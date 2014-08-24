redis-mount [![Build Status](http://img.shields.io/travis/poying/redis-mount.svg?style=flat)](https://travis-ci.org/poying/redis-mount)
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

## Download

* [mac-amd64](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-darwin-amd64)
* [linux-amd64](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-amd64)
* [linux-386](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-386)
* [linux-arm](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-arm)

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
