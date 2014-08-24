all: clean test build

test:
	@go test -v ./redisfs

build:
	@go build main.go

clean:
	@-rm main.a

.PHONY: main.a
