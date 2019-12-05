package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/weft"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
	"net/http"
	"strings"
	"time"
)

/*
	This endpoint lists all the potential metadata fields
*/
func metaHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {

	//Get the domain from the path
	path := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(path) != 3 {
		return weft.NoMatch(r, h, b)
	}

	domain := path[1]
	_, ok := domainMap[domain]
	if !ok {
		return valid.Error{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("domain '%v' is not valid", domain),
		}
	}

	var out proto.Message
	var err error

	switch path[2] {
	case "list": //list all metadata keys and values in the domain
		out, err = metaList(r, h, b, domain)
	case "entries": //list the metadata for a specific 'key'
		out, err = metaEntries(r, h, b, domain)
	default:
		return weft.NoMatch(r, h, b)
	}

	if err != nil {
		return err
	}

	return returnProto(out, r, h, b)
}

func querySimple(keyQ, valQ, domain string, now time.Time) ([]string, error) {
	out := make([]string, 0)
	result, err := db.Query("SELECT DISTINCT(record_key) FROM dapper.metadata WHERE record_domain=$1 AND field=$2 AND value=$3 AND timespan @> $4::TIMESTAMPTZ;", domain, keyQ, valQ, now)
	if err != nil {
		return out, fmt.Errorf("failed to execute simple query: %v", err)
	}

	for result.Next() {
		var key string
		err = result.Scan(&key)
		if err != nil {
			return out, fmt.Errorf("failed to scan for simple query result: %v", err)
		}
		out = append(out, key)
	}
	return out, nil
}

func tagsSimple(tag, domain string, now time.Time) ([]string, error) {
	out := make([]string, 0)
	result, err := db.Query("SELECT DISTINCT(record_key) FROM dapper.metadata WHERE record_domain=$1 AND field=$2 AND istag AND timespan @> $3::TIMESTAMPTZ;", domain, tag, now)
	if err != nil {
		return out, fmt.Errorf("failed to execute tag query: %v", err)
	}

	for result.Next() {
		var key string
		err = result.Scan(&key)
		if err != nil {
			return out, fmt.Errorf("failed to scan for tag query results: %v", err)
		}
		out = append(out, key)
	}
	return out, nil
}

func metaEntries(r *http.Request, h http.Header, b *bytes.Buffer, domain string) (proto.Message, error) {
	v, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"key", "aggregate", "query", "tags"}, valid.Query)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	key := v.Get("key")
	query := v.Get("query")
	tags := v.Get("tags")
	var qRes []string

	count := 0
	if key != "" {
		count++
	}
	if query != "" {
		count++
	}
	if tags != "" {
		count++
	}
	if count > 1 {
		return nil, valid.Error{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("only one of 'key', 'query', 'tags' may be provided"),
		}
	}

	if query != "" {
		keyQ, valQ, err := valid.ParseQuery(query)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("failed to parse query string: %v", err),
			}
		}
		qRes, err = querySimple(keyQ, valQ, domain, now)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed to perform simple query: %v", err),
			}
		}
	} else if tags != "" {
		tagsList := strings.Split(tags, ",")
		if len(tagsList) != 1 {
			return nil, valid.Error{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("only one tag may be specified at this time"),
			}
		}
		qRes, err = tagsSimple(tagsList[0], domain, now)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("failed simple tag query: %v", err),
			}
		}
	} else if key != "" {
		key = strings.Replace(key, "*", "%", -1) //% is the wildcard in postgres but easier to send an * in urls.
	} else {
		key = "%"
	}

	var result *sql.Rows
	//Get locations
	if key != "" {
		result, err = db.Query("SELECT record_key, ST_X(geom::geometry) as longitude, ST_Y(geom::geometry) as latitude FROM dapper.metageom WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::TIMESTAMPTZ;", domain, key, now) //TODO: Allow starttime/endtime queries
	} else if query != "" || tags != "" {
		result, err = db.Query("SELECT record_key, ST_X(geom::geometry) as longitude, ST_Y(geom::geometry) as latitude FROM dapper.metageom WHERE record_domain=$1 AND record_key = ANY($2) AND timespan @> $3::TIMESTAMPTZ;", domain, pq.Array(qRes), now)
	} else {
		err = fmt.Errorf("was not set to either key or query")
	}
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("location metadata query failed: %v", err),
		}
	}

	locMap := make(map[string]*dapperlib.Point)
	for result.Next() {
		var key string
		var lon, lat float32
		err = result.Scan(&key, &lon, &lat)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("scanning loc entry failed: %v", err),
			}
		}

		locMap[key] = &dapperlib.Point{
			Latitude:  lat,
			Longitude: lon,
		}
	}

	//get metadata

	if key != "" {
		result, err = db.Query("SELECT record_key, field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::TIMESTAMPTZ ORDER BY record_key;", domain, key, now) //TODO: Allow starttime/endtime queries
	} else if query != "" || tags != "" {
		result, err = db.Query("SELECT record_key, field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND record_key = ANY($2) AND timespan @> $3::TIMESTAMPTZ ORDER BY record_key;", domain, pq.Array(qRes), now)
	} else {
		err = fmt.Errorf("was not set to either key or query")
	}
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("metadata entry query failed: %v", err),
		}
	}

	out := &dapperlib.KeyMetadataSnapshotList{Metadata: make([]*dapperlib.KeyMetadataSnapshot, 0)}

	var kms *dapperlib.KeyMetadataSnapshot
	for result.Next() {
		var key, field, value string
		var istag bool
		err = result.Scan(&key, &field, &value, &istag)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("scanning metadata entries failed: %v", err),
			}
		}

		if kms == nil || kms.Key != key {
			// find Links
			frList := make([]*dapperlib.LinkSpan, 0)
			toList := make([]*dapperlib.LinkSpan, 0)
			fr, err := db.Query("SELECT from_key, rel_type, from_locality, to_locality, ST_X(geom::geometry) AS from_longitude, ST_Y(geom::geometry) AS from_latitude FROM dapper.metarel r, dapper.metageom g WHERE r.record_domain=$1 AND r.to_key=$2 AND r.timespan @> $3::TIMESTAMPTZ AND g.record_key=r.from_key ORDER BY r.from_key;", domain, key, now)
			if err != nil {
				return nil, valid.Error{
					Code: http.StatusInternalServerError,
					Err:  fmt.Errorf("metarel entry query failed: %v", err),
				}
			}

			for fr.Next() {
				var k, rel, fromLoc, toLoc string
				var fromX, fromY float32
				err = fr.Scan(&k, &rel, &fromLoc, &toLoc, &fromX, &fromY)
				if err != nil {
					return nil, valid.Error{
						Code: http.StatusInternalServerError,
						Err:  fmt.Errorf("scanning metarel entries failed: %v", err),
					}
				}
				frList = append(frList, &dapperlib.LinkSpan{
					// Domain:  domain,	// No need to output domain
					FromKey:      k,
					FromLocality: fromLoc,
					ToKey:        key,
					ToLocality:   toLoc,
					GeoJson:      lineString(fromX, fromY, locMap[key].Longitude, locMap[key].Latitude),
					// RelType:      ty, // TODO: This would be a duplicate info as device already has it
				})
			}

			to, err := db.Query("SELECT to_key,rel_type, from_locality, to_locality, ST_X(geom::geometry) AS to_longitude, ST_Y(geom::geometry) AS to_latitude FROM dapper.metarel r, dapper.metageom g WHERE r.record_domain=$1 AND r.from_key=$2 AND r.timespan @> $3::TIMESTAMPTZ AND g.record_key=r.to_key ORDER BY r.to_key;", domain, key, now)
			if err != nil {
				return nil, valid.Error{
					Code: http.StatusInternalServerError,
					Err:  fmt.Errorf("metarel entry query failed: %v", err),
				}
			}

			for to.Next() {
				var k, rel, fromLoc, toLoc string
				var toX, toY float32
				err = to.Scan(&k, &rel, &fromLoc, &toLoc, &toX, &toY)
				if err != nil {
					return nil, valid.Error{
						Code: http.StatusInternalServerError,
						Err:  fmt.Errorf("scanning metarel entries failed: %v", err),
					}
				}
				toList = append(toList, &dapperlib.LinkSpan{
					// Domain:  domain, // No need to output domain
					FromKey:      key,
					FromLocality: fromLoc,
					ToKey:        k,
					ToLocality:   toLoc,
					GeoJson:      lineString(locMap[key].Longitude, locMap[key].Latitude, toX, toY),
					// RelType:      ty, // TODO: This would be a duplicate info as device already has it
				})
			}
			kms = &dapperlib.KeyMetadataSnapshot{
				Domain:   domain,
				Key:      key,
				Moment:   now.Unix(),
				Metadata: make(map[string]string),
				Tags:     make([]string, 0),
				Location: locMap[key],
				Links:    append(frList, toList...),
			}
			out.Metadata = append(out.Metadata, kms)
		}

		if istag {
			kms.Tags = append(kms.Tags, field)
		} else {
			kms.Metadata[field] = value
		}
	}

	aggregate := v.Get("aggregate")
	if aggregate != "" {
		out = aggregateKMS(out, aggregate)
	}

	return out, nil
}

func aggregateKMS(in *dapperlib.KeyMetadataSnapshotList, aggr string) *dapperlib.KeyMetadataSnapshotList {

	aggrMap := make(map[string]*dapperlib.KeyMetadataSnapshot)

	for _, kms := range in.Metadata {
		aggrVal := kms.Metadata[aggr]
		if aggrVal == "" {
			continue
		}
		aggrKey := fmt.Sprintf("%s:%s", aggr, aggrVal)

		aggrKMS, ok := aggrMap[aggrKey]
		if !ok {
			kms.Key = aggrKey
			aggrMap[aggrKey] = kms
			continue
		}

		//compare metadata fields
		for k, v := range kms.Metadata {
			aggrVal, ok := aggrKMS.Metadata[k]
			if !ok {
				continue
			}
			if v != aggrVal {
				delete(aggrKMS.Metadata, k)
			}
		}

		//concatenate tag fields
		for _, t := range kms.Tags {
			found := false
			for _, tComp := range aggrKMS.Tags {
				if t == tComp {
					found = true
					break
				}
			}
			if !found {
				aggrKMS.Tags = append(aggrKMS.Tags, t)
			}
		}

		//compare location
		if aggrKMS.Location != nil && (kms.Location.Latitude != aggrKMS.Location.Latitude ||
			kms.Location.Longitude != aggrKMS.Location.Longitude) {
			aggrKMS.Location = nil
		}
	}

	out := &dapperlib.KeyMetadataSnapshotList{Metadata: make([]*dapperlib.KeyMetadataSnapshot, 0)}

	for _, v := range aggrMap {
		out.Metadata = append(out.Metadata, v)
	}

	return out
}

func metaList(r *http.Request, h http.Header, b *bytes.Buffer, domain string) (proto.Message, error) {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	result, err := db.Query("SELECT DISTINCT field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::TIMESTAMPTZ ORDER BY field, value;", domain, now) //TODO: Allow starttime/endtime queries
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("metadata values query failed: %v", err),
		}
	}

	out := &dapperlib.DomainMetadataList{
		Domain:   domain,
		Keys:     make([]string, 0),
		Metadata: make(map[string]*dapperlib.MetadataValuesList),
		Tags:     make([]string, 0),
	}

	for result.Next() {
		var field, value string
		var istag bool
		err = result.Scan(&field, &value, &istag)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("scanning metadata values failed: %v", err),
			}
		}
		if istag {
			out.Tags = append(out.Tags, field)
		} else {
			meta, ok := out.Metadata[field]
			if !ok {
				meta = &dapperlib.MetadataValuesList{
					Name:   field,
					Values: make([]string, 0),
				}
				out.Metadata[field] = meta
			}
			meta.Values = append(meta.Values, value)
		}
	}

	result, err = db.Query("SELECT DISTINCT record_key FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::TIMESTAMPTZ ORDER BY record_key;", domain, now) //TODO: Allow starttime/endtime queries
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("metadata keys query failed: %v", err),
		}
	}

	for result.Next() {
		var key string
		err = result.Scan(&key)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("scanning metadata keys failed: %v", err),
			}
		}
		out.Keys = append(out.Keys, key)
	}

	return out, nil
}

func lineString(x0, y0, x1, y1 float32) string {
	return fmt.Sprintf("{\"type\":\"LineString\", \"coordinates\":[[%.5f, %.5f], [%.5f, %.5f]]}", x0, y0, x1, y1)
}
