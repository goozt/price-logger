run:
	@go run . serve

init: reset
	@go run . init

reset:
	@rm -rf ./pb_data ./dist

build:
	@go build -o ./dist/server
	@./dist/server init

prod:
	@./dist/server start --http=0.0.0.0:8080

stop:
	@./dist/server stop --port=8080

serve:
	@./dist/server serve --http=0.0.0.0:8080

test: stop reset build prod

.PHONY: run init reset build prod stop serve test
