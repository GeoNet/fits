CREATE TABLE public.map180_layers (
	mapPK SERIAL PRIMARY KEY,
	region INT NOT NULL,
	zoom INT NOT NULL,
	type INT NOT NULL
);

SELECT addgeometrycolumn('public', 'map180_layers', 'geom', 3857, 'MULTIPOLYGON', 2);

CREATE INDEX ON public.map180_layers (zoom);
CREATE INDEX ON public.map180_layers (region);
CREATE INDEX ON public.map180_layers (type);
CREATE INDEX ON public.map180_layers USING gist (geom);

GRANT SELECT ON public.map180_layers TO PUBLIC;

-- land = type 0
-- lakes = type 1

-- World. Region 0
insert into public.map180_layers (region,zoom,type,geom) select 0,0,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-85,180,85, 4326), geom),3857))
 from public.ne50land;
 -- lakes
 insert into public.map180_layers (region,zoom,type,geom) select 0,0,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-85,180,85, 4326), geom),3857))
 from public.ne50lakes;


-- New Zealand.  Region 1 
-- NE50  Left and right of 180.  
insert into public.map180_layers (region,zoom,type,geom) select 1,0,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom),3857))
 from public.ne50land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom));
 insert into public.map180_layers (region,zoom,type,geom) select 1,0,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom),3857)) 
 	from public.ne50land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom));
 insert into public.map180_layers (region,zoom,type,geom) select 1,0,1,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom),3857))
 from public.ne50lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom));
 insert into public.map180_layers (region,zoom,type,geom) select 1,0,1,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-48,-160,-27, 4326), geom),3857)) 
 	from public.ne50lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(-180,-48,-160,-27, 4326), geom));	

-- NE10  Left and right of 180.  
insert into public.map180_layers (region,zoom,type,geom) select 1,1,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom),3857))
 from public.ne10land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom));
 insert into public.map180_layers (region,zoom,type,geom) select 1,1,0,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom),3857)) 
 	from public.ne10land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom)); 	
 insert into public.map180_layers (region,zoom,type,geom) select 1,1,1,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom),3857))
 from public.ne10lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(160,-48,180,-27, 4326), geom));
 insert into public.map180_layers (region,zoom,type,geom) select 1,1,1,  ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom),3857)) 
 	from public.ne10lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(-180,-48,-175,-27, 4326), geom)); 	

-- NZTOPO
--
-- zoom 2: 1500k small feaures removed for performance 
--
insert into public.map180_layers (region,zoom,type,geom) select 1,2,0,  
	ST_Multi(ST_Transform(geom,3857)) from public.nztopo_1500k_land where st_area(geom) *111*111 > 0.2 ;
insert into public.map180_layers (region,zoom,type,geom) select 1,2,1,  
	ST_Multi(ST_Transform(geom,3857)) from public.nztopo_1500k_lakes  where st_area(geom) *111*111 > 0.5;

--  Raoul is missing from 1500k.  Add it using filtered 50k
insert into public.map180_layers (region,zoom,type,geom) select 1,2,0,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom),3857)) 
 from public.nztopo_150k_land where 
 Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom)) and st_area(geom) *111 *111 > 0.5;


-- Chathams missing from 1500k, use 1250k files.
insert into public.map180_layers (region,zoom,type,geom) select 1,2,0,  ST_Multi(ST_Transform(geom,3857))
 from public.nztopo_1250k_chathams_land where st_area(geom) *111*111 > 0.5 ;
 insert into public.map180_layers (region,zoom,type,geom) select 1,2,1,  ST_Multi(ST_Transform(geom,3857))
 from public.nztopo_1250k_chathams_lagoon ;


--
-- zoom 3: 150k small feaures removed for performance 
--
insert into public.map180_layers (region,zoom,type,geom) select 1,3,0,  
	ST_Multi(ST_Transform(geom,3857)) from public.nztopo_150k_land where st_area(geom) *111*111 > 0.2;
insert into public.map180_layers (region,zoom,type,geom) select 1,3,1,  
	ST_Multi(ST_Transform(geom,3857)) from public.nztopo_150k_lakes where st_area(geom) *111*111 > 0.5 ;


-- Delete Raoul at zoom 3.  Then put data back with all small features.
delete from public.map180_layers where not ST_IsEmpty(ST_Intersection(ST_Transform(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326),3857),geom)) 
	and zoom =3;
insert into public.map180_layers (region,zoom,type,geom) select 1,3,0,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom),3857)) 
 from public.nztopo_150k_land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom)) ;	
insert into public.map180_layers (region,zoom,type,geom) select 1,3,1,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom),3857)) 
 from public.nztopo_150k_lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(182.0,-29.30,182.14,-29.22, 4326), geom)) ;

-- Delete White Island at zoom 3.  Then put data back with all small features.
delete from public.map180_layers where not ST_IsEmpty(ST_Intersection(ST_Transform(ST_MakeEnvelope(177.164,-37.54,177.20,-37.505, 4326),3857),geom)) 
	and zoom =3;
insert into public.map180_layers (region,zoom,type,geom) select 1,3,0,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(177.164,-37.54,177.20,-37.505, 4326), geom),3857)) 
 from public.nztopo_150k_land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(177.164,-37.54,177.20,-37.505, 4326), geom)) ;	
insert into public.map180_layers (region,zoom,type,geom) select 1,3,1,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(177.164,-37.54,177.20,-37.505, 4326), geom),3857)) 
 from public.nztopo_150k_lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(177.164,-37.54,177.20,-37.505, 4326), geom)) ;

-- Delete Chathams at zoom 3.  Then put back with all small features.
-- 183,-44.5,184,-43.5
delete from public.map180_layers where not ST_IsEmpty(ST_Intersection(ST_Transform(ST_MakeEnvelope(183,-44.5,184,-43.5, 4326),3857),geom)) 
	and zoom =3;
insert into public.map180_layers (region,zoom,type,geom) select 1,3,0,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(183,-44.5,184,-43.5, 4326), geom),3857)) 
 from public.nztopo_150k_land where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(183,-44.5,184,-43.5, 4326), geom)) ;	
insert into public.map180_layers (region,zoom,type,geom) select 1,3,1,  
	ST_Multi(ST_Transform(ST_Intersection(ST_MakeEnvelope(183,-44.5,184,-43.5, 4326), geom),3857)) 
 from public.nztopo_150k_lakes where Not ST_IsEmpty(ST_Intersection(ST_MakeEnvelope(183,-44.5,184,-43.5, 4326), geom)) ;

DROP TABLE  public.ne50land;
DROP TABLE  public.ne50lakes;
DROP TABLE  public.ne10minorislands;
DROP TABLE  public.ne10land;
DROP TABLE  public.ne10lakes;
DROP TABLE  public.nztopo_1250k_chathams_land;
DROP TABLE  public.nztopo_1250k_chathams_lagoon;
DROP TABLE  public.nztopo_1500k_land;
DROP TABLE  public.nztopo_150k_land;
DROP TABLE  public.nztopo_1500k_lakes;
DROP TABLE  public.nztopo_150k_lakes;







