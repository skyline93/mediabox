SOURCE_MIRROR ?= http://mirrors.aliyun.com/debian
UID := $(shell id -u)

build-all: swag build-go

build-go:
	go build -o mediabox cmd/mediabox/mediabox.go
	go build -o mediabox-agent cmd/agent/agent.go

swag:
	swag init -g api.go --dir ./internal/api/ -o ./internal/api/docs
	swag fmt -g ./internal/api/api.go

build-dev:
	docker build --build-arg SOURCE_MIRROR=$(SOURCE_MIRROR) --build-arg USER_ID=$$(id -u) --build-arg GROUP_ID=$$(id -g) -t mediabox-dev:latest .

terminal:
	docker compose exec -it mediabox zsh

build-prd:
	docker build -t mediabox:latest -f docker/Dockerfile .
	docker build -t mediabox-frontend:latest -f docker/Dockerfile-frontend .

clean:
	docker compose down
	rm -rf mediabox mediabox-agent* logs/* mediabox-data
