package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/weft"
	"github.com/golang/protobuf/proto"
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

	fmt.Println(path)

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

/*
	TODO: Be able to query metadata, need some requirements so I can design the API and the resulting SQL queries.
*/

func metaEntries(r *http.Request, h http.Header, b *bytes.Buffer, domain string) (proto.Message, error) {
	v, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"key", "aggregate"}, valid.Query)
	if err != nil {
		return nil, err
	}

	key := v.Get("key")
	key = strings.Replace(key, "*", "%", -1) //% is the wildcard in postgres but easier to send an * in urls.

	if key == "" {
		key = "%"
	}

	now := time.Now().UTC()

	result, err := db.Query("SELECT record_key, ST_X(geom::geometry) as longitude, ST_Y(geom::geometry) as latitude FROM dapper.metageom WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::timestamp;", domain, key, now) //TODO: Allow starttime/endtime queries
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

	result, err = db.Query("SELECT record_key, field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND record_key ILIKE $2 AND timespan @> $3::timestamp ORDER BY record_key;", domain, key, now) //TODO: Allow starttime/endtime queries
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
			kms = &dapperlib.KeyMetadataSnapshot{
				Domain:   domain,
				Key:      key,
				Moment:   now.Unix(),
				Metadata: make(map[string]string),
				Tags:     make([]string, 0),
				Location: locMap[key],
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

	result, err := db.Query("SELECT DISTINCT field, value, istag FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::timestamp ORDER BY field, value;", domain, now) //TODO: Allow starttime/endtime queries
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

	result, err = db.Query("SELECT DISTINCT record_key FROM dapper.metadata WHERE record_domain=$1 AND timespan @> $2::timestamp ORDER BY record_key;", domain, now) //TODO: Allow starttime/endtime queries
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
