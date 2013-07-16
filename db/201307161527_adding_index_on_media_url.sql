DROP INDEX IF EXISTS sr_media_url;
UPDATE service_requests SET media_url = NULL WHERE media_url = '';
CREATE INDEX sr_media_url ON service_requests(media_url);