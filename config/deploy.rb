set :stages, ['staging', 'production']
set :default_stage, 'staging'
require "capistrano/ext/multistage"

set :application, "cwfy"
set :repository,  "git@github.com:smartchicago/chicagoworksforyou.git"
set :scm, :git

set :user, "ec2-user"
set :use_sudo, false

after 'deploy:update', 'deploy:compile:api'
after 'deploy:update', 'deploy:compile:worker'
after 'deploy:update', 'deploy:restart'

namespace :db do
  desc "prepare a snapshot of the CWFY database, store to S3"
  task :snapshot do
    timestamp = Time.now.strftime("%Y-%m-%d-%H%M")
    backup_file = "/tmp/cwfy-#{stage}-#{timestamp}.dump"
    run "pg_dump -O -C -c --format=custom -f #{backup_file} #{database} && \
      s3cmd --no-encrypt --acl-public --reduced-redundancy put #{backup_file} s3://cwfy-database-backups/#{stage}.dump && \
      rm -f #{backup_file}"
  end
  
  desc "download latest snapshot and load into local database"
  task :restore do
    run_locally "dropdb #{database} && \
      createdb #{database} && \
      curl -o /tmp/cwfy-restore-#{stage}.dump http://cwfy-database-backups.s3.amazonaws.com/#{stage}.dump
      pg_restore -d #{database} -O -c /tmp/cwfy-restore-#{stage}.dump && \
      rm -f /tmp/cwfy-restore-#{stage}.dump"
  end
  
end

namespace :deploy do
  namespace :compile do
    task :api do
      out = "server"
      run_locally "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/#{out} -ldflags '-X main.version `git rev-parse --short HEAD`' api/server.go"
      top.upload "/tmp/#{out}", "#{release_path}/bin/#{out}", mode: "0755", via: :scp
      run_locally "rm -f /tmp/#{out}"
    end

    task :worker do
      out = "fetch"
      run_locally "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/#{out} -ldflags '-X main.version `git rev-parse --short HEAD`' api/fetch.go"
      top.upload "/tmp/#{out}", "#{release_path}/bin/#{out}", mode: "0755", via: :scp
      run_locally "rm -f /tmp/#{out}"
    end
  end
  
  task (:restart) { sudo "supervisorctl restart #{stage}:*", pty: true } 
  task (:start) { sudo "sudo supervisorctl start #{stage}:*", pty: true }
  task (:stop) { sudo "sudo supervisorctl stop #{stage}:*", pty: true }
end