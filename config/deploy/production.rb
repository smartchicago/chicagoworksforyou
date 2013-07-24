server "cwfy-api.smartchicagoapps.org", :app, :db, primary: true
set :deploy_to, "/var/www/cwfy/production"
set :supervisor_group, "production"