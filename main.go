package main

import "os"
import "fmt"
import "path/filepath"
import "github.com/codegangsta/cli"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/hanwen/go-fuse/fuse/pathfs"
import "github.com/hanwen/go-fuse/fuse/nodefs"
import "github.com/poying/redis-mount/redisfs"

var App *cli.App

// app name
var Name = "redis-mount"

// app version
var Version = "0.0.0"

// redis host name
var HostFlag = cli.StringFlag{
	Name:  "host",
	Value: "localhost",
	Usage: "set redis host name",
}

// redis port number
var PortFlag = cli.IntFlag{
	Name:  "port",
	Value: 6379,
	Usage: "set redis port number",
}

// redis port number
var DbFlag = cli.IntFlag{
	Name:  "db",
	Value: 0,
	Usage: "set redis database",
}

// redis password
var AuthFlag = cli.StringFlag{
	Name:  "auth",
	Usage: "set redis password",
}

// redis key separator
var SepFlag = cli.StringFlag{
	Name:  "sep",
	Value: ":",
	Usage: "set redis key separator",
}

// help message template
var AppHelpTemplate = "" +
	"\n" +
	"  \u001b[36m{{.Name}}\u001b[39m \u001b[33m{{.Version}}\u001b[39m\n" +
	"  $ mkdir /tmp/redis && {{.Name}} /tmp/redis\n" +
	"\n" +
	"  {{range .Flags}}{{.}}\n" +
	"  {{end}}\n"

func main() {
	cli.AppHelpTemplate = AppHelpTemplate

	App = cli.NewApp()
	App.Name = Name
	App.Version = Version

	App.Flags = []cli.Flag{
		HostFlag,
		PortFlag,
		DbFlag,
		AuthFlag,
		SepFlag,
	}

	App.Action = run

	App.Run(os.Args)
}

func run(ctx *cli.Context) {
	if len(ctx.Args()) == 0 {
		cli.ShowAppHelp(ctx)
		return
	}

	server, err := mount(ctx)

	if err != nil {
		fmt.Printf("\n  \u001b[35m%s\u001b[39m: %s\n\n", "Error", err)
		return
	}

	server.Serve()
}

func mount(ctx *cli.Context) (*fuse.Server, error) {
	mnt, err := filepath.Abs(ctx.Args().Get(0))

	if err != nil {
		return nil, err
	}

	fs := &redisfs.RedisFs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		Host:       ctx.String("host"),
		Port:       ctx.Int("port"),
		Db:         ctx.Int("db"),
		Auth:       ctx.String("auth"),
		Dirs:       make(map[string][]string),
		Sep:        ctx.String("sep"),
	}

	fs.Init()

	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mnt, nfs.Root(), nil)

	if err != nil {
		return nil, err
	}

	return server, nil
}
