run:
	@go run . serve

init: reset
	@go run . init

reset:
	@rm -rf ./pb_data ./dist

build:
	@rm -rf ./dist/
	@go build -o ./dist/server
	@go build -C internal/cli/install -o ../../../dist/
	@go build -C internal/cli/uninstall -o ../../../dist/

prod-build:
	@rm -rf ./dist/
	@go build -o ./server
	@./server init

prod-run:
	@./server serve --http=0.0.0.0:8080 --encryptionEnv ${PB_DATA_ENCRYPTION_KEY}

uninstall:
	@cd dist && sudo ./uninstall

install: build uninstall
	@cd dist && sudo ./install
