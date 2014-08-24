all: clean test build

test:
	@go test -v ./redisfs

build:
	@go build main.go

get-deps:
	@go get github.com/poying/go-chalk
	@go get github.com/codegangsta/cli
	@go get github.com/hanwen/go-fuse/fuse
	@go get github.com/garyburd/redigo/redis
	@go get github.com/smartystreets/goconvey/convey

clean:
	-@rm main

.PHONY: main.a
