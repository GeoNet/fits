CREATE FUNCTION fits.add_site(siteID_n TEXT, name_n TEXT, longitude_n NUMERIC, latitude_n NUMERIC, height_n NUMERIC, ground_relationship_n NUMERIC) RETURNS VOID AS
$$
DECLARE
tries INTEGER = 0;
BEGIN
LOOP
UPDATE fits.site 
SET height = height_n, ground_relationship = ground_relationship_n, name = name_n, location =  ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(longitude_n, latitude_n), 4326))) 
WHERE siteID = siteID_n;
IF found THEN
RETURN;
END IF;

BEGIN
INSERT INTO fits.site(siteID, name, location, height, ground_relationship)
VALUES (siteID_n, name_n, ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint(longitude_n, latitude_n), 4326))), height_n, ground_relationship_n);
RETURN;
EXCEPTION WHEN unique_violation THEN
--  Loop once more to see if a different insert happened after the update but before our insert.
tries = tries + 1;
if tries > 1 THEN
RETURN;
END IF;
END;
END LOOP;
END;
$$
LANGUAGE plpgsql;

CREATE FUNCTION fits.add_observation(siteID_n TEXT, typeID_n TEXT, methodID_n TEXT, sampleID_n TEXT, systemID_n TEXT, time_n TIMESTAMP(6) WITH TIME ZONE, value_n NUMERIC, error_n NUMERIC ) RETURNS VOID AS
$$
DECLARE
tries INTEGER = 0;
BEGIN
LOOP
UPDATE fits.observation 
SET value = value_n, error = error_n 
WHERE observation.sitepk = (select sitepk from fits.site where siteID = siteID_n)
AND observation.samplepk = (select samplepk from fits.sample join fits.system using (systempk) where systemID = systemID_n and sampleID = sampleID_n) 
AND observation.typepk = (select typepk from fits.type where typeID = typeID_n) 
AND observation.methodpk = (select methodpk from fits.method where methodID = methodID_n) 
AND observation.time = time_n ;
IF found THEN
RETURN;
END IF;

BEGIN
INSERT INTO fits.observation(sitepk, typepk, methodpk, samplepk, time, value, error) SELECT site.sitepk, typepk, methodpk, ss.samplepk, time_n, value_n, error_n
from fits.site, fits.type, fits.method, (fits.sample join fits.system using (systempk)) as ss
where site.siteID = siteID_n
and method.methodID = methodID_n
and type.typeID = typeID_n
and ss.sampleID = sampleID_n
and ss.systemID = systemID_n; 
RETURN;
EXCEPTION WHEN unique_violation THEN
--  Loop once more to see if a different insert happened after the update but before our insert.
tries = tries + 1;
if tries > 1 THEN
RETURN;
END IF;
END;
END LOOP;
END;
$$
LANGUAGE plpgsql;