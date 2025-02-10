build:
	@go build -C app -o ../dist/
	@go build -C cli/reset -o ../../dist/
	@go build -C cli/install -o ../../dist/
	@go build -C cli/uninstall -o ../../dist/
