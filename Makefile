.PHONY: api workers

workers:
	go run api/fetch.go --environment development

api:
	go run api/server.go --environment development
