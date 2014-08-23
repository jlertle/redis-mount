package redisfs

import "os"
import "fmt"
import "path"
import "regexp"
import "strings"
import "github.com/poying/go-chalk"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/garyburd/redigo/redis"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/hanwen/go-fuse/fuse/pathfs"

type RedisFs struct {
	pathfs.FileSystem
	Conn redis.Conn
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
	content, err1 := redis.String(fs.Conn.Do("GET", key))
	list, err2 := redis.Strings(fs.Conn.Do("KEYS", key + ":*"))

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

func (fs *RedisFs) OpenDir(name string, ctx *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	pattern := nameToPattern(name)
	res, err := redis.Strings(fs.Conn.Do("KEYS", pattern))

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	entries := resToEntries(nameToKey(name), res)

	return entries, fuse.OK
}

func (fs *RedisFs) Open(name string, flags uint32, ctx *fuse.Context) (nodefs.File, fuse.Status) {
	key := nameToKey(name)
	_, err := fs.Conn.Do("EXISTS", key)

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	return NewRedisFile(fs.Conn, key), fuse.OK
}

func (fs *RedisFs) Create(name string, flags uint32, mode uint32, ctx *fuse.Context) (nodefs.File, fuse.Status) {
	key := nameToKey(name)
	_, err := fs.Conn.Do("SET", key, "")

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	return NewRedisFile(fs.Conn, key), fuse.OK
}

func (fs *RedisFs) Unlink(name string, ctx *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.OK
	}

	key := nameToKey(name)
	_, err := fs.Conn.Do("DEL", key)

	if err != nil {
		printError(err)
		return fuse.ENOENT
	}

	return fuse.OK
}

func (fs *RedisFs) Rmdir(name string, ctx *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.OK
	}

	pattern := nameToPattern(name)
	list, err := redis.Strings(fs.Conn.Do("KEYS", pattern))

	if err != nil {
		printError(err)
		return fuse.ENOENT
	}

	for _, el := range list {
		_, err := fs.Conn.Do("DEL", el)
		if err != nil {
			printError(err)
			return fuse.ENOENT
		}
	}

	return fuse.OK
}

func (fs *RedisFs) Mkdir(name string, mode uint32, ctx *fuse.Context) fuse.Status {
	key := nameToKey(name) + ":.redis-mount-folder"
	_, err := fs.Conn.Do("SET", key, 1)

	if err != nil {
		printError(err)
		return fuse.ENOENT
	}

	return fuse.OK
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

func printError(err error) {
	fmt.Printf("  %s: %s\n", chalk.Magenta("Error"), err)
}
