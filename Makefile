build-all: swag build-go

build-go:
	go build -o mediabox cmd/mediabox/mediabox.go
	go build -o mediabox-agent cmd/agent/agent.go

swag:
	swag init -g api.go --dir ./internal/api/ -o ./internal/api/docs
	swag fmt -g ./internal/api/api.go

clean:
	rm -rf mediabox mediabox-agent* logs/*
