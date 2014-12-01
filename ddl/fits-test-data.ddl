insert into fits.network (networkID, description) VALUES ('TN1', 'Test network 1');
insert into fits.network (networkID, description) VALUES ('TN2', 'Test network 2');

-- Add the same site info twice and make sure it doesn't change.
select fits.add_site('TN1',  'TEST1', $$Test site 1$$, 172.79019, -42.21496, -999.9, 0.0);
select fits.add_site('TN1',  'TEST1', $$Test site 1$$, 172.79019, -42.21496, -999.9, 0.0);
-- Add a site and then update the height.  Make sure it does change.
select fits.add_site('TN1', 'TEST2', $$Test site 2$$, 172.79019, -42.21496, -999.9, 0.0);
select fits.add_site('TN1', 'TEST2', $$Test site 2$$, 172.79019, -42.21496, -111.1, 0.0);
-- Add another site TEST2 but in the TN2 network
select fits.add_site('TN2', 'TEST2', $$Test site 2$$, 172.79019, -42.21496, -999.99, 0.0);

insert into fits.unit(symbol, name) VALUES ('m', 'metre');
insert into fits.unit(symbol, name) VALUES ('K', 'Kelvin');

insert into fits.type (typeID, name, description, unitPK) VALUES ('t1', 'Type 1', 'Test data type 1', 1);
insert into fits.type (typeID, name, description, unitPK) VALUES ('t2', 'Type 1', 'Test data type 2', 2);

insert into fits.method (methodID, name, description, reference) VALUES ('m1', 'Method 1', 'Test data method 1', 'a link to more information about method 1');
insert into fits.method (methodID, name, description, reference) VALUES ('m2', 'Method 2', 'Test data method 2', 'a link to more information about method 2');

-- Being lazy with sequence values and assuming these rows are being insterted into a clean freshly created DB.
insert into fits.type_method (typePK, methodPK) VALUES (1,1);	
insert into fits.type_method (typePK, methodPK) VALUES (1,2);
insert into fits.type_method (typePK, methodPK) VALUES (2,1);

insert into fits.system(systemID, description) VALUES ('none', 'No external system reference');	
insert into fits.system(systemID, description) VALUES ('lab', 'Some external lab system');	

insert into fits.sample(sampleID, systemPK) VALUES ('none', 1);
insert into fits.sample(sampleID, systemPK) VALUES ('0001', 2);

-- Add some observations
select fits.add_observation('TN1', 'TEST1', 't1', 'm1', 'none', 'none', '2000-01-06T12:00:00.000000Z'::timestamptz, 1.52, 0);
select fits.add_observation('TN1', 'TEST1', 't1', 'm1', 'none', 'none',  '2000-01-07T12:00:00.000000Z'::timestamptz, 2.52, 0);
select fits.add_observation('TN1', 'TEST1', 't1', 'm2', 'none', 'none',  '2000-01-08T12:00:00.000000Z'::timestamptz, 3.52, 0);
-- Add an observation and update it.
select fits.add_observation('TN1', 'TEST1', 't1', 'm1', 'none', 'none',  '2000-01-09T12:00:00.000000Z'::timestamptz, 3.52, 0);
select fits.add_observation('TN1', 'TEST1', 't1', 'm1', 'none', 'none',  '2000-01-09T12:00:00.000000Z'::timestamptz, 4.52, 1.1);
-- Add observation at same site, same type, same time but different method.
select fits.add_observation('TN1', 'TEST2', 't1', 'm1', 'none', 'none',  '2000-01-08T12:00:00.000000Z'::timestamptz, 4.52, 1.1);
select fits.add_observation('TN1', 'TEST2', 't1', 'm2', 'none', 'none',  '2000-01-08T12:00:00.000000Z'::timestamptz, 4.02, 0.1);
-- Add some observations for sample 0001.
select fits.add_observation('TN1', 'TEST2', 't1', 'm2', '0001', 'lab',  '2001-01-08T12:00:00.000000Z'::timestamptz, 9.02, 0.1);
select fits.add_observation('TN1', 'TEST2', 't1', 'm1', '0001', 'lab',  '2001-01-08T12:00:00.000000Z'::timestamptz, 9.12, 0.01);
select fits.add_observation('TN1', 'TEST2', 't2', 'm1', '0001', 'lab',  '2001-01-08T12:00:00.000000Z'::timestamptz, 9.12, 0.01);
