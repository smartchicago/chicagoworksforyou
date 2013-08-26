rebuild:
	bundle exec compass compile src
	bundle exec jekyll build

server:
	bundle exec compass watch src &
	bundle exec jekyll serve --watch

STAGING_BUCKET    = s3://cwfy-staging.smartchicagoapps.org/
PRODUCTION_BUCKET = s3://chicagoworksforyou.com/

S3CMD = s3cmd -c .s3cfg \
  --acl-public \
	--no-delete-removed \
	--reduced-redundancy \
	--progress \
	--rexclude '\.git|.DS_Store' \
	sync ./site/

all: staging

deploy:
	$(S3CMD) $(PRODUCTION_BUCKET)

staging:
	bundle exec jekyll build --config _config.yml,_config_staging.yml
	$(S3CMD) $(STAGING_BUCKET)

production:
	bundle exec jekyll build --config _config.yml,_config_production.yml
	$(S3CMD) $(PRODUCTION_BUCKET)
