server "cwfy-api.smartchicagoapps.org", :app, :db, primary: true
set :branch, 'master'
set :deploy_to, "/var/www/cwfy/production"
set :stage, "production"
set :database, "cwfy-production"