APPNAME=redis-mount

define build
	echo $(APPNAME)-$(1)-$(2); \
	GO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -o "bin/$(APPNAME)-$(1)-$(2)" "main.go";
endef

all: clean test build

test:
	@go test -v ./redisfs

build:
	@$(call build,linux,amd64)
	@$(call build,linux,386)
	@$(call build,linux,arm)
	@$(call build,darwin,amd64)

get-deps:
	@go get github.com/poying/go-chalk
	@go get github.com/codegangsta/cli
	@go get github.com/hanwen/go-fuse/fuse
	@go get github.com/garyburd/redigo/redis
	@go get github.com/smartystreets/goconvey/convey

clean:
	-@rm -r bin
