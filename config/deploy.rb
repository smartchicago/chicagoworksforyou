set :stages, [:staging, :production]
set :default_stage, :staging
require "capistrano/ext/multistage"

set :application, "cwfy"
set :repository,  "git@github.com:smartchicago/chicagoworksforyou.git"
set :scm, :git

set :user, "ec2-user"
set :use_sudo, false

after 'deploy:update', 'deploy:compile:api'
after 'deploy:update', 'deploy:compile:worker'
after 'deploy:update', 'deploy:restart'

namespace :deploy do
  namespace :compile do
    task :api do
      out = "server"
      run_locally "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/#{out} api/server.go"
      top.upload "/tmp/#{out}", "#{release_path}/bin/#{out}", mode: "0755", via: :scp
      run_locally "rm -f /tmp/#{out}"
    end

    task :worker do
      out = "fetch"
      run_locally "export GOOS=linux && export GOARCH=amd64 && /usr/local/bin/go build -o /tmp/#{out} api/fetch.go"
      top.upload "/tmp/#{out}", "#{release_path}/bin/#{out}", mode: "0755", via: :scp
      run_locally "rm -f /tmp/#{out}"
    end
  end
  
  task (:restart) { sudo "supervisorctl restart #{supervisor_group}:*", pty: true } 
  task (:start) { sudo "sudo supervisorctl start #{supervisor_group}:*", pty: true }
  task (:stop) { sudo "sudo supervisorctl stop #{supervisor_group}:*", pty: true }
end

task :asdf do
  run_locally 'pwd'
end