.PHONY: proto gen up down run

proto: gen

gen:
	protoc --go_out=. --go_opt=module=github.com/IvanOplesnin/BotTradeService.git \
		--go-grpc_out=. --go-grpc_opt=module=github.com/IvanOplesnin/BotTradeService.git \
		-I proto \
		proto/*.proto

up:
	goose -env .env up

down:
	goose -env .env down

run:
	ENV_FILE=./.env ./run.sh