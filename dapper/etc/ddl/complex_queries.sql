/* OR Query */
SELECT DISTINCT q1.record_key
FROM (SELECT record_key
      FROM dapper.metadata
      WHERE record_domain = 'fdmp'
        AND (
              (field = 'model' AND LOWER(value) LIKE '%cusp%')
              OR (LOWER(field) = 'canary' AND istag = TRUE)
          )
        AND timespan @> NOW()::timestamp) AS q1;

/* AND Query */


SELECT DISTINCT(record_key)
FROM dapper.metadata
WHERE record_domain = 'fdmp'
  AND (field = 'model' AND LOWER(value) LIKE '%cusp%')
  AND record_key IN (SELECT DISTINCT(record_key)
                     FROM dapper.metadata
                     WHERE record_domain = 'fdmp'
                       AND (field = 'canary' AND istag = TRUE)
                       AND timespan @> NOW()::timestamp)
  AND record_key IN (SELECT DISTINCT(record_key)
                      FROM dapper.metadata
                      WHERE record_domain = 'fdmp'
                        AND (field = 'locality' AND value ILIKE '%taita%')
                        AND timespan @> NOW()::timestamp)
  AND timespan @> NOW()::timestamp;