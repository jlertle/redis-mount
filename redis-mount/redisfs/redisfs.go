package redisfs

import "os"
import "path"
import "regexp"
import "strings"
import "strconv"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/garyburd/redigo/redis"
import . "github.com/visionmedia/go-debug"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/hanwen/go-fuse/fuse/pathfs"

var debug = Debug("redisfs")

type RedisFs struct {
	pathfs.FileSystem
	Host string
	Port int
	Auth string
	conn redis.Conn
}

func New(host string, port int, auth string) *RedisFs {
	fs := &RedisFs{
		Host: host,
		Port: port,
		Auth: auth,
		FileSystem: pathfs.NewDefaultFileSystem(),
	}
	return fs
}

func (fs *RedisFs) ConnectRedis() (*RedisFs, error) {
	address := fs.Host + ":" + strconv.Itoa(fs.Port)
	conn, err := redis.Dial("tcp", address)

	debug("connect to %s", address)

	if err != nil {
		debug("connect failed")
		return nil, err
	}

	fs.conn = conn;
	fs.FileSystem = pathfs.NewDefaultFileSystem()

	if len(fs.Auth) > 0 {
		debug("auth")
		if _, err := conn.Do("AUTH", fs.Auth); err != nil {
			debug("auth failed");
			conn.Close()
			return nil, err
		}
	}

	return fs, nil
}

func (fs *RedisFs) GetAttr(name string, ctx *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}

	if string(name[0]) == "." {
		return nil, fuse.ENOENT
	}

	key := nameToKey(name)

	debug("GetAttr %s %s", name, key)

	content, err1 := redis.String(fs.conn.Do("GET", key))
	list, err2 := redis.Strings(fs.conn.Do("KEYS", key + ":*"))

	switch {
	case err2 == nil && len(list) > 0:
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
		break;
	case err1 == nil:
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644,
			Size: uint64(len(content)),
		}, fuse.OK
		break;
	}

	return nil, fuse.ENOENT
}

func (fs *RedisFs) OpenDir(name string, ctx *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	pattern := nameToPattern(name)

	debug("OpenDir %s %s", name, pattern)

	res, err := redis.Strings(fs.conn.Do("KEYS", pattern))

	if err != nil {
		return nil, fuse.ENOENT
	}

	entries := resToEntries(nameToKey(name), res)

	return entries, fuse.OK
}

func (fs *RedisFs) Open(name string, flags uint32, ctx *fuse.Context) (file nodefs.File, code fuse.Status) {
	key := nameToKey(name)

	debug("Open %s %s", name, key)

	content, err := redis.String(fs.conn.Do("GET", key))

	if err != nil {
		return nil, fuse.ENOENT
	}

	return nodefs.NewDataFile([]byte(content)), fuse.OK
}

func nameToPattern(name string) string {
	pattern := nameToKey(name)

	if name == "" {
		pattern += "*"
	} else {
		pattern += ":*"
	}

	return pattern;
}

func resToEntries(root string, list []string) []fuse.DirEntry {
	m := make(map[string]bool)
	entries := make([]fuse.DirEntry, 0)
	offset := len(root)
	sepCount := strings.Count(root, string(os.PathSeparator)) + 1

	if  offset != 0 {
		offset += 1
	}

	for _, el := range list {
		key := el[offset:]

		switch strings.Count(key, ":") {
		case 0:
			entries = append(entries, fuse.DirEntry{
				Name: keyToName(key),
				Mode: fuse.S_IFREG,
			})
			break
		case sepCount:
			key = path.Clean(path.Join(keyToName(key), ".."))
			_, ok := m[key]
			if !ok {
				m[key] = true
				entries = append(entries, fuse.DirEntry{
					Name: key,
					Mode: fuse.S_IFDIR,
				})
			}
		}
	}

	return entries
}

func nameToKey(name string) string {
	re := regexp.MustCompile(string(os.PathSeparator))
	key := re.ReplaceAllLiteralString(name, ":")
	key = decodePathSeparator(key)
	return key
}

func keyToName(key string) string {
	name := encodePathSeparator(key)
	re := regexp.MustCompile(":")
	name = re.ReplaceAllLiteralString(name, string(os.PathSeparator))
	return name
}

func encodePathSeparator(str string) string {
	re := regexp.MustCompile(string(os.PathSeparator))
	str = re.ReplaceAllLiteralString(str, "\uffff")
	return str;
}

func decodePathSeparator(str string) string {
	re := regexp.MustCompile("\uffff")
	str = re.ReplaceAllLiteralString(str, string(os.PathSeparator))
	return str;
}
