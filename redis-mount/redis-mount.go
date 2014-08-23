package redisMount

import "fmt"
import "path/filepath"
import "github.com/hanwen/go-fuse/fuse/pathfs"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/poying/redis-mount/redis-mount/redisfs"

func Mount(host string, port int, auth string, mnt string) (*pathfs.PathNodeFs, error) {
	mnt, err := filepath.Abs(mnt)

	if (err != nil) {
		fmt.Printf("Redis connect failed. (%s)", err)
		return nil, err
	}

	fs := redisfs.New(host, port, auth)
	_, err = fs.ConnectRedis()

	if (err != nil) {
		return nil, err
	}

	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mnt, nfs.Root(), nil)

	if (err != nil) {
		fmt.Printf("Mount failed. (%s)", err)
		return nil, err
	}

  server.Serve()

	return nfs, nil
}
