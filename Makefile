.PHONY: proto gen

proto: gen

gen:
	protoc --go_out=. --go_opt=module=github.com/IvanOplesnin/BotTradeService.git \
		--go-grpc_out=. --go-grpc_opt=module=github.com/IvanOplesnin/BotTradeService.git \
		-I proto \
		proto/*.proto
