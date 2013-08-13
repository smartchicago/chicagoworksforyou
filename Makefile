.PHONY: api workers

workers:
	go run api/fetch.go api/environment.go api/service_request.go --environment development

api:
	go run api/server.go api/environment.go api/helpers.go api/*_handler.go --environment development
