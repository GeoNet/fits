DELETE FROM dapper.records where record_domain='test_db_archive' or record_domain='test_api';
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key1', 'field1', NOW(), '1.1', false, NOW());
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key1', 'field2', NOW(), '1.1', false, NOW()- interval '1 hour');
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key1', 'field3', NOW(), '1.1', false, NOW());
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key2', 'field1', NOW(), '1.1', false, NOW());
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key2', 'field2', NOW(), '1.1', false, NOW());
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_db_archive', 'test_key2', 'field3', NOW()- interval '16 day', '1.1', true, NOW() );
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_api', 'rfap5g-soundstage', 'temperature', NOW(), '1.1', false, NOW());
INSERT INTO dapper.records(record_domain, record_key, field, time, value, archived, modtime) VALUES ('test_api', 'rfap5g-soundstage', 'voltage', NOW(), '1.1', false, NOW()- interval '1 hour');

DELETE FROM dapper.metadata where record_domain='test_api';
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', 'locality', 'towai', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', 'model', 'MikroTik Routerboard', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', 'hostname', 'rf2soundstage-towai', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', 'ipaddr', '10.236.80.25', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', 'x1', '', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', true);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rf2soundstage-towai', '5G', '', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', true);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', 'locality', 'soundstage', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', 'model', 'MikroTik Routerboard', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', 'hostname', 'rfap5g-soundstage', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', 'ipaddr', '10.236.0.5', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', 'wifi', '', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', true);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'rfap5g-soundstage', '5G', '', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', true);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'wanrt-soundstage', 'locality', 'soundstage', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'wanrt-soundstage', 'model', 'MikroTik Routerboard', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'wanrt-soundstage', 'hostname', 'wanrt-soundstage', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);
INSERT INTO dapper.metadata(record_domain, record_key, field, value, timespan, istag) VALUES ('test_api', 'wanrt-soundstage', 'ipaddr', '10.236.0.5', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")', false);

DELETE FROM dapper.metageom where record_domain='test_api';
INSERT INTO dapper.metageom(record_domain, record_key, geom, timespan) VALUES ('test_api', 'rfap5g-soundstage', '0101000020E6100000000000C0DADD6540000000A0249944C0', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")');
INSERT INTO dapper.metageom(record_domain, record_key, geom, timespan) VALUES ('test_api', 'wanrt-soundstage', '0101000020E6100000000000C0DADD6540000000A0249944C0', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")');
INSERT INTO dapper.metageom(record_domain, record_key, geom, timespan) VALUES ('test_api', 'rf2soundstage-towai', '0101000020E6100000000000C0DADD6540000000A0249944C0', '["2019-11-14 19:21:53+00","9999-01-01 00:00:00+00")');

DELETE FROM dapper.metarel where record_domain='test_api';
INSERT INTO dapper.metarel(record_domain, from_key, to_key, rel_type, timespan) VALUES ('test_api', 'rf2soundstage-towai', 'rfap5g-soundstage', '5G', '["2020-03-12 02:25:06+00","9999-01-01 00:00:00+00")');
