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
	conn redis.Conn
	dirs map[string][]string
}

func NewRedisFs(fs pathfs.FileSystem, conn redis.Conn) *RedisFs {
	return &RedisFs{
		FileSystem: fs,
		conn: conn,
		dirs: make(map[string][]string),
	}
}

func (fs *RedisFs) GetAttr(name string, ctx *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}

	// ignore hidden files
	if string(name[0]) == "." {
		return nil, fuse.ENOENT
	}

	// find dir in memory
	dirs, ok := fs.dirs[path.Dir(name)]
	baseName := path.Base(name)

	if ok {
		exist, _ := stringInSlice(baseName, dirs)
		if exist {
			return &fuse.Attr{
				Mode: fuse.S_IFDIR | 0755,
			}, fuse.OK
		}
	}

	// find attr in redis
	key := nameToKey(name)
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

func (fs *RedisFs) OpenDir(name string, ctx *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	pattern := nameToPattern(name)
	res, err := redis.Strings(fs.conn.Do("KEYS", pattern))

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	entries := resToEntries(nameToKey(name), res)

	if name == "" {
		name = "."
	}

	if list, ok := fs.dirs[name]; ok {
		for _, key := range list {
			entries = append(entries, fuse.DirEntry{
				Name: key,
				Mode: fuse.S_IFDIR,
			})
		}
	}

	return entries, fuse.OK
}

func (fs *RedisFs) Open(name string, flags uint32, ctx *fuse.Context) (nodefs.File, fuse.Status) {
	key := nameToKey(name)
	_, err := fs.conn.Do("EXISTS", key)

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	return NewRedisFile(fs.conn, key), fuse.OK
}

func (fs *RedisFs) Create(name string, flags uint32, mode uint32, ctx *fuse.Context) (nodefs.File, fuse.Status) {
	key := nameToKey(name)
	_, err := fs.conn.Do("SET", key, "")

	if err != nil {
		printError(err)
		return nil, fuse.ENOENT
	}

	return NewRedisFile(fs.conn, key), fuse.OK
}

func (fs *RedisFs) Unlink(name string, ctx *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.OK
	}

	key := nameToKey(name)
	_, err := fs.conn.Do("DEL", key)

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

	// check if name is in memory
	dirName := path.Dir(name)
	dir, ok := fs.dirs[dirName]
	baseName := path.Base(name)
	
	if ok {
		exist, index := stringInSlice(baseName, dir)
		if exist {
			fs.dirs[dirName] = append(dir[:index], dir[index + 1:]...)
			return fuse.OK
		}
	}

	// if name isn't in memory then find it in redis
	pattern := nameToPattern(name)
	list, err := redis.Strings(fs.conn.Do("KEYS", pattern))

	if err != nil {
		printError(err)
		return fuse.ENOENT
	}

	for _, el := range list {
		_, err := fs.conn.Do("DEL", el)
		if err != nil {
			printError(err)
			return fuse.ENOENT
		}
	}

	return fuse.OK
}

func (fs *RedisFs) Mkdir(name string, mode uint32, ctx *fuse.Context) fuse.Status {
	dir := path.Join(name, "..")

	_, ok := fs.dirs[dir]

	if !ok {
		fs.dirs[dir] = make([]string, 0, 10)
	}

	fs.dirs[dir] = append(fs.dirs[dir], path.Base(name))

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

func stringInSlice(target string, list []string) (bool, int) {
	for i, str := range list {
		if str == target {
			return true, i
		}
	}
	return false, -1
}
