build-all: swag build-go

build-go:
	go build -o mediabox cmd/mediabox/mediabox.go
	go build -o mediabox-agent cmd/agent/agent.go

swag:
	swag init -g api.go --dir ./internal/api/ -o ./internal/api/docs
	swag fmt -g ./internal/api/api.go

build-dev:
	docker build --build-arg SOURCE_MIRROR=$$SOURCE_MIRROR --build-arg USER_ID=$$(id -u) --build-arg GROUP_ID=$$(id -g) -t mediabox-dev:latest .

terminal:
	docker run -it --rm -v $$(pwd):/project/src -w /project/src --user $$(id -u):$$(id -g) mediabox-dev:latest

clean:
	rm -rf mediabox mediabox-agent* logs/*
