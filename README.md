redis-mount [![Build Status](http://img.shields.io/travis/poying/redis-mount.svg?style=flat)](https://travis-ci.org/poying/redis-mount)
===========

redis-mount lets you use Redis as a filesystem.

```bash
redis-mount 0.0.0
$ redis-mount ~/redis

--host, -h   localhost    Redis host name
--port, -p   6379         Redis port number
--auth, -a                Redis password
--sep, -s    :            Redis key separator
```

## What we can do with it?

1. Use `grep` to search for text in redis values.
2. Pass data to other programs. ex: `$ cat redis-key | pretty-print`

![screenshot](./screenshot.gif)

## Installation

### Download binary file

* [mac-amd64](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-darwin-amd64)
* [linux-amd64](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-amd64)
* [linux-386](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-386)
* [linux-arm](https://github.com/poying/redis-mount/releases/download/20140824/redis-mount-linux-arm)

### Build from source

#### Requirement

* fuse

It is easy to build redis-mount from the source code. It takes four steps:

1. Install `fuse` ([linux](http://fuse.sourceforge.net/), [mac](http://osxfuse.github.io/)).
2. Get the redis-mount source code from GitHub
  
  ```bash
  $ git clone https://github.com/poying/redis-mount.git
  ```
  
3. Change to the directory with the redis-mount source code and run
  
  ```bash
  $ make get-deps
  ```
  
  to install dependencies.

4. Run `go build` and then you can see a binary file in current directory.

```bash
$ git clone git@github.com:poying/redis-mount.git
$ cd redis-mount
```

### Run Unit Tests

```bash
$ make test
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
