# Deployment

The CWFY application consists of two components: the API backend, written in Go, and the AngularJS/Jekyll/HTML frontend. The two components are contained in this repository, but are deployed separately.

## Frontend Deployment

Frontend code deployment is handled via a Makefile command. This command invokes jekyll, to rebuild the site, and then the [s3cmd utility](http://s3tools.org/s3cmd) to copy the files to the appropriate S3 bucket. s3cmd will look in the `frontend/` directory for a .s3cfg configuration file. To create this file, run `s3cmd -c .s3cfg --configure` in the `frontend/` directory. You will be prompted to input your AWS access key and secret key. This is required for you to be able to connect to S3 and deploy the code. 

For example:

    $ make staging
    bundle exec jekyll build --config _config.yml,_config_staging.yml
    Configuration file: _config.yml
    Configuration file: _config_staging.yml
                Source: src
           Destination: site
          Generating... done.
    s3cmd -c .s3cfg --acl-public --no-delete-removed --reduced-redundancy --progress --rexclude '\.git|.DS_Store' sync ./site/ s3://cwfy-staging.smartchicagoapps.org/
    ./site/js/serviceApp.js -> s3://cwfy-staging.smartchicagoapps.org/js/serviceApp.js  [1 of 2]
     6440 of 6440   100% in    0s    40.70 kB/s  done
    ./site/js/wardApp.js -> s3://cwfy-staging.smartchicagoapps.org/js/wardApp.js  [2 of 2]
     16776 of 16776   100% in    0s    79.69 kB/s  done
    Done. Uploaded 23216 bytes in 0.4 seconds, 59.35 kB/s

## Backend Deployment

The API backend may be deployed using [Capistrano](https://github.com/capistrano/capistrano). There are two deployment environments defined for CWFY: staging and production. The only distinction between the two are the database names and the deployment paths on the CWFY server. The deploy command will compile the Go binaries on your local machine, then copy them to the server via scp. See the REQUIREMENTS file for information about build requirements. You must be authorized to SSH into the CWFY server to deploy the code.

To deploy:

        cap <environment> deploy

For example:

        $ cap production deploy
            triggering load callbacks
          * 2013-08-20 14:09:46 executing `production'
            triggering start callbacks for `deploy'
          * 2013-08-20 14:09:46 executing `multistage:ensure'
          * 2013-08-20 14:09:46 executing `deploy'
          * 2013-08-20 14:09:46 executing `deploy:update'
         ** transaction: start
          * 2013-08-20 14:09:46 executing `deploy:update_code'
            executing locally: "git ls-remote git@github.com:smartchicago/chicagoworksforyou.git master"
          * executing "git clone -q -b master git@github.com:smartchicago/chicagoworksforyou.git /PATH/TO/cwfy/production/releases/20130820190947 && cd /PATH/TO/cwfy/production/releases/20130820190947 && git checkout -q -b deploy 81851b874c4f7a1cc08224f69ba647f2031263cf && (echo 81851b874c4f7a1cc08224f69ba647f2031263cf > /PATH/TO/cwfy/production/releases/20130820190947/REVISION)"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
            command finished in 1465ms
          * 2013-08-20 14:09:50 executing `deploy:finalize_update'
          * executing "chmod -R g+w /PATH/TO/cwfy/production/releases/20130820190947"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
            command finished in 134ms
          * 2013-08-20 14:09:50 executing `deploy:create_symlink'
          * executing "rm -f /PATH/TO/cwfy/production/current && ln -s /PATH/TO/cwfy/production/releases/20130820190947 /PATH/TO/cwfy/production/current"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
            command finished in 154ms
         ** transaction: commit
            triggering after callbacks for `deploy:update'
          * 2013-08-20 14:09:50 executing `deploy:compile:api'
            executing locally: "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/server -ldflags \"-X main.version `git rev-parse --short HEAD`\" api/server.go api/helpers.go api/environment.go api/*_handler.go"
            servers: ["cwfy-api.smartchicagoapps.org"]
         ** scp upload /tmp/server -> /PATH/TO/cwfy/production/releases/20130820190947/bin/server
            [cwfy-api.smartchicagoapps.org] /tmp/server
          * scp upload complete
          * executing "chmod 0755 /PATH/TO/cwfy/production/releases/20130820190947/bin/server"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
            command finished in 188ms
            executing locally: "rm -f /tmp/server"
          * 2013-08-20 14:10:01 executing `deploy:compile:worker'
            executing locally: "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/fetch -ldflags \"-X main.version `git rev-parse --short HEAD`\" api/fetch.go api/environment.go api/service_request.go"
            servers: ["cwfy-api.smartchicagoapps.org"]
         ** scp upload /tmp/fetch -> /PATH/TO/cwfy/production/releases/20130820190947/bin/fetch
            [cwfy-api.smartchicagoapps.org] /tmp/fetch
          * scp upload complete
          * executing "chmod 0755 /PATH/TO/cwfy/production/releases/20130820190947/bin/fetch"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
            command finished in 138ms
            executing locally: "rm -f /tmp/fetch"
          * 2013-08-20 14:10:10 executing `deploy:restart'
          * executing "sudo -p 'sudo password: ' supervisorctl restart production:*"
            servers: ["cwfy-api.smartchicagoapps.org"]
            [cwfy-api.smartchicagoapps.org] executing command
         ** [out :: cwfy-api.smartchicagoapps.org] cwfy-api-production: stopped
         ** [out :: cwfy-api.smartchicagoapps.org] cwfy-worker-production: stopped
         ** [out :: cwfy-api.smartchicagoapps.org] cwfy-api-production: started
         ** [out :: cwfy-api.smartchicagoapps.org] cwfy-worker-production: started
            command finished in 3873ms

### Other backend commands

Database snapshot: trigger a snapshot of the database for a given environment, store the result in S3, and return the public URL of the dump file.

        $ cap staging db:snapshot
            triggering load callbacks
          * 2013-08-20 14:00:17 executing `staging'
            triggering start callbacks for `db:snapshot'
          * 2013-08-20 14:00:17 executing `multistage:ensure'
          * 2013-08-20 14:00:17 executing `db:snapshot'
          * executing "pg_dump -O -C -c --format=custom -f /tmp/cwfy-staging-2013-08-20-1400.dump cwfy &&       s3cmd --no-encrypt --acl-public --reduced-redundancy put /tmp/cwfy-staging-2013-08-20-1400.dump s3://cwfy-database-backups/staging.dump &&       rm -f /tmp/cwfy-staging-2013-08-20-1400.dump"
            servers: ["cwfy-api-staging.smartchicagoapps.org"]
            [cwfy-api-staging.smartchicagoapps.org] executing command
        *** [err :: cwfy-api-staging.smartchicagoapps.org] WARNING: Module python-magic is not available. Guessing MIME types based on file extensions.
         ** [out :: cwfy-api-staging.smartchicagoapps.org] File '/tmp/cwfy-staging-2013-08-20-1400.dump' stored as 's3://cwfy-database-backups/staging.dump' (221319803 bytes in 11.2 seconds, 18.76 MB/s) [1 of 1]
         ** [out :: cwfy-api-staging.smartchicagoapps.org] Public URL of the object is: http://cwfy-database-backups.s3.amazonaws.com/staging.dump
            command finished in 145827ms


Database restore: download a copy of a database snapshot and replace a local database with the contents of the dump file. Note: this operation may take a good deal of time (more then 20 minutes) to complete. The database dump file is hundreds of megabytes and may take a while to download, depending on your Internet connection.

        $ cap staging db:restore
            triggering load callbacks
          * 2013-08-20 14:35:05 executing `staging'
            triggering start callbacks for `db:restore'
          * 2013-08-20 14:35:05 executing `multistage:ensure'
          * 2013-08-20 14:35:05 executing `db:restore'
            executing locally: "dropdb cwfy &&       createdb cwfy &&       curl -o /tmp/cwfy-restore-staging.dump http://cwfy-database-backups.s3.amazonaws.com/staging.dump\n      pg_restore -d cwfy -O -c /tmp/cwfy-restore-staging.dump &&       rm -f /tmp/cwfy-restore-staging.dump"

        (PostgreSQL and curl warnings are omitted)
        
## Server configuration

CWFY runs on a Amazon Web Services m1.medium instance. The instance contains a PostgreSQL 9.2.4 database server with PostGIS 2.0.3 extensions. The CWFY API and fetch worker are managed by [supervisord](http://supervisord.org/). The there is a cron job (file: `/etc/cron.daily/ebs-snapshot`) scheduled to run nightly to create backup snapshots of the instance's EBS volume. This script is configured to maintain 14 days of snapshot history.