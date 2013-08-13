server "cwfy-api-staging.smartchicagoapps.org", :app, :db, primary: true
set :branch, 'development'
set :deploy_to, "/var/www/cwfy/staging"
set :stage, "staging"
set :database, "cwfy"