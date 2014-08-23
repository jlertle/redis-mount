package redisMount

import "path/filepath"
import . "github.com/visionmedia/go-debug"
import "github.com/hanwen/go-fuse/fuse/pathfs"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/poying/redis-mount/redis-mount/redisfs"

var debug = Debug("redis-mount")

func Mount(host string, port int, auth string, mnt string) (*pathfs.PathNodeFs, error) {
	mnt, err := filepath.Abs(mnt)

	if (err != nil) {
		return nil, err
	}

	debug("mount %s:%d %s", host, port, mnt)

	fs := redisfs.New(host, port, auth)
	_, err = fs.ConnectRedis()

	if (err != nil) {
		return nil, err
	}

	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mnt, nfs.Root(), nil)

	if (err != nil) {
		debug("mount failed")
		return nil, err
	}

  server.Serve()

	return nfs, nil
}
