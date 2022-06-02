


.PHONY: build
build:
	go mod tidy && go build -o ./get ./

