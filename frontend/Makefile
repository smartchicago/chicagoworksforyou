rebuild:
	bundle exec compass compile src
	bundle exec jekyll build

server:
	bundle exec compass watch src &
	bundle exec jekyll serve --watch

STAGING_BUCKET    = s3://cwfy-staging.smartchicagoapps.org/

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
	$(S3CMD) $(STAGING_BUCKET)
