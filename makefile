build:
	@rm -rf ./dist/
	@go build -C app -o ../dist/
	@go build -C cli/reset -o ../../dist/
	@go build -C cli/install -o ../../dist/
	@go build -C cli/uninstall -o ../../dist/

uninstall:
	@cd dist && sudo ./uninstall

install: build uninstall
	@cd dist && sudo ./install
