redis-mount
===========

Use Redis as a filesystem

![screenshot](./screenshot.gif)

```bash
  redis-mount 0.0.0
  $ redis-mount ~/redis

  --host, -h   localhost    Redis host name
  --port, -p   6379         Redis port number
  --auth, -a                Redis password
```

## Build Requirement

* fuse

## Install

```bash
$ go get github.com/poying/redis-mount
```

## Unmount

Linux

```bash
$ fusermount -u /tmp/redis
```

MacOS

```bash
$ diskutil unmount /tmp/redis
```

## License

(The MIT License)

Copyright (c) 2014 Po-Ying Chen &lt;poying.me@gmail.com&gt;.
