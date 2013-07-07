.PHONY: api workers

workers:
	go run workers/fetch.go --environment development

api:
	go run api/server.go --environment development
