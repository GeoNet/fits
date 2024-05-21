package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/weft"
	"github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

const sqlReverseFindKey = `SELECT DISTINCT(record_key) FROM dapper.metadata WHERE record_domain=$1 AND field=$2 AND value=$3 AND timespan @> $4::TIMESTAMPTZ;`
const sqlListKeys = `SELECT DISTINCT record_key FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::TIMESTAMPTZ ORDER BY record_key;`
const sqlMetaFields = `SELECT DISTINCT field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::TIMESTAMPTZ ORDER BY field, value;`
const sqlTag = `SELECT DISTINCT(record_key) FROM dapper.metadata WHERE record_domain=$1 AND field=$2 AND istag AND timespan @> $3::TIMESTAMPTZ;`
const sqlILike = `SELECT record_key, field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::TIMESTAMPTZ ORDER BY record_key;`
const sqlAny = `SELECT record_key, field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND record_key = ANY($2) AND timespan @> $3::TIMESTAMPTZ ORDER BY record_key;`
const sqlGeomILike = `SELECT record_key, ST_X(geom::geometry) as longitude, ST_Y(geom::geometry) as latitude FROM dapper.metageom WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::TIMESTAMPTZ;`
const sqlGeomAny = `SELECT record_key, ST_X(geom::geometry) as longitude, ST_Y(geom::geometry) as latitude FROM dapper.metageom WHERE record_domain=$1 AND record_key = ANY($2) AND timespan @> $3::TIMESTAMPTZ;`
const sqlRelILike = `SELECT from_key, to_key, rel_type FROM dapper.metarel WHERE record_domain=$1 AND (from_key ILIKE $2 OR to_key ILIKE $2) AND timespan @> $3::TIMESTAMPTZ;`
const sqlRelAny = `SELECT from_key, to_key, rel_type FROM dapper.metarel WHERE record_domain=$1 AND (from_key = ANY($2) OR to_key = ANY($2)) AND timespan @> $3::TIMESTAMPTZ;`

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
	result, err := db.Query(sqlReverseFindKey, domain, keyQ, valQ, now)
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
	result, err := db.Query(sqlTag, domain, tag, now)
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
		result, err = db.Query(sqlGeomILike, domain, key, now) //TODO: Allow starttime/endtime queries
	} else if query != "" || tags != "" {
		result, err = db.Query(sqlGeomAny, domain, pq.Array(qRes), now)
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

	// preload all relations so we don't query in the loop
	if key != "" {
		result, err = db.Query(sqlRelILike, domain, key, now)
	} else if query != "" || tags != "" {
		result, err = db.Query(sqlRelAny, domain, pq.Array(qRes), now)
	}
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("relation metadata query failed: %v", err),
		}
	}

	allRels := make(map[string]map[string]string) // map<[fromKey], map<toKey, reltype>>
	revRels := make(map[string]map[string]string) // map<[toKey], map<fromKey, reltype>>
	for result.Next() {
		var from, to, rel string
		err = result.Scan(&from, &to, &rel)
		if err != nil {
			return nil, valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("relation metadata query failed: %v", err),
			}
		}
		var r map[string]string
		var ok bool
		if r, ok = allRels[from]; !ok {
			r = make(map[string]string)
		}
		r[to] = rel
		allRels[from] = r

		if r, ok = revRels[to]; !ok {
			r = make(map[string]string)
		}
		r[from] = rel
		revRels[to] = r // We'll only do map lookups so it's safe to reuse the same "d" object
	}

	//get metadata

	if key != "" {
		result, err = db.Query(sqlILike, domain, key, now) //TODO: Allow starttime/endtime queries
	} else if query != "" || tags != "" {
		result, err = db.Query(sqlAny, domain, pq.Array(qRes), now)
	} else {
		err = fmt.Errorf("was not set to either key or query")
	}
	if err != nil {
		return nil, valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("metadata entry query failed: %v", err),
		}
	}

	out := &dapperlib.KeyMetadataSnapshotList{
		Metadata: make([]*dapperlib.KeyMetadataSnapshot, 0),
	}

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

		rels := make([]*dapperlib.SnapshotRelation, 0)
		for t, r := range allRels[key] {
			rels = append(rels, &dapperlib.SnapshotRelation{FromKey: key, ToKey: t, RelType: r})
		}
		for t, r := range revRels[key] {
			rels = append(rels, &dapperlib.SnapshotRelation{FromKey: t, ToKey: key, RelType: r})
		}
		if kms == nil || kms.Key != key {
			kms = &dapperlib.KeyMetadataSnapshot{
				Domain:    domain,
				Key:       key,
				Moment:    now.Unix(),
				Metadata:  make(map[string]string),
				Tags:      make([]string, 0),
				Location:  locMap[key],
				Relations: rels,
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
	keyMap := make(map[string]*dapperlib.KeyMetadataSnapshot)
	aggrMap := make(map[string]*dapperlib.KeyMetadataSnapshot)

	// Build the map with original key, we need this to lookup when adding relations later
	for _, kms := range in.Metadata {
		keyMap[kms.Key] = kms
	}

	for _, kms := range in.Metadata {
		aggrVal := kms.Metadata[aggr]
		if aggrVal == "" {
			continue
		}
		aggrKey := fmt.Sprintf("%s:%s", aggr, aggrVal)

		aggrKMS, ok := aggrMap[aggrKey]
		if !ok {
			kms.Key = aggrKey
			kms.Relations = rekeyRelations(kms, aggr, keyMap)
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

		//bring in linked relations
		kms.Relations = rekeyRelations(kms, aggr, keyMap)
		for _, l := range kms.Relations {
			found := false

			for _, al := range aggrKMS.Relations {
				if l.FromKey == al.FromKey && l.ToKey == al.ToKey && l.RelType == al.RelType {
					found = true
					break
				}
			}
			if !found {
				aggrKMS.Relations = append(aggrKMS.Relations, l)
			}
		}

		aggrMap[aggrKey] = aggrKMS
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

	result, err := db.Query(sqlMetaFields, domain, now) //TODO: Allow starttime/endtime queries
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

	result, err = db.Query(sqlListKeys, domain, now) //TODO: Allow starttime/endtime queries
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

func rekeyRelations(kms *dapperlib.KeyMetadataSnapshot, aggr string, keyMap map[string]*dapperlib.KeyMetadataSnapshot) []*dapperlib.SnapshotRelation {
	newRels := make([]*dapperlib.SnapshotRelation, 0)
	for _, l := range kms.Relations {
		// The keys in aggrKMS is already altered as "<aggr>:<meta-value>"
		lFrom := keyMap[l.FromKey].Metadata[aggr]
		lTo := keyMap[l.ToKey].Metadata[aggr]

		if lFrom == "" || lTo == "" {
			continue
		}

		lFrom = fmt.Sprintf("%s:%s", aggr, lFrom)
		lTo = fmt.Sprintf("%s:%s", aggr, lTo)

		newRels = append(newRels, &dapperlib.SnapshotRelation{
			FromKey: lFrom,
			ToKey:   lTo,
			RelType: l.RelType,
		})
	}

	return newRels
}
