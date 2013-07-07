.PHONY: api workers

workers:
	go run workers/fetch.go

api:
	go run api/server.go --environment development
