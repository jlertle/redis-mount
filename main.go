package main

import "os"
import "fmt"
import "strings"
import "path/filepath"
import "github.com/poying/go-chalk"
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
	Name:  "host, h",
	Value: "localhost",
	Usage: "Redis host name",
}

// redis port number
var PortFlag = cli.IntFlag{
	Name:  "port, p",
	Value: 6379,
	Usage: "Redis port number",
}

// redis port number
var DbFlag = cli.IntFlag{
	Name:  "db, d",
	Value: 0,
	Usage: "Redis database",
}

// redis password
var AuthFlag = cli.StringFlag{
	Name:  "auth, a",
	Usage: "Redis password",
}

// redis key separator
var SepFlag = cli.StringFlag{
	Name:  "sep, s",
	Value: ":",
	Usage: "Redis key separator",
}

func main() {
	App = cli.NewApp()
	App.HideHelp = true
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
		PrintHelpMessage()
		return
	}

	server, err := mount(ctx)

	if err != nil {
		fmt.Printf("\n  %s: %s\n\n", chalk.Magenta("Error"), err)
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
		Host:				ctx.String("host"),
		Port:       ctx.Int("port"),
		Db:         ctx.Int("db"),
		Auth:				ctx.String("auth"),
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

func PrintHelpMessage() {
	println()
	fmt.Printf("  %s %s\n", chalk.Cyan(App.Name), chalk.Green(App.Version))
	println("  $ redis-mount ~/redis")
	println()

	fmt.Printf("  %-12s %-12v %s\n",
		prefixNames(HostFlag.Name), HostFlag.Value, HostFlag.Usage)

	fmt.Printf("  %-12s %-12v %s\n",
		prefixNames(PortFlag.Name), PortFlag.Value, PortFlag.Usage)

	fmt.Printf("  %-12s %-12v %s\n",
		prefixNames(DbFlag.Name), DbFlag.Value, DbFlag.Usage)

	fmt.Printf("  %-12s %-12v %s\n",
		prefixNames(AuthFlag.Name), AuthFlag.Value, AuthFlag.Usage)

	fmt.Printf("  %-12s %-12v %s\n",
		prefixNames(SepFlag.Name), SepFlag.Value, SepFlag.Usage)

	println()
}

func prefixNames(fullName string) (prefixed string) {
	first := true
	parts := strings.Split(fullName, ",")

	for _, name := range parts {
		name = strings.Trim(name, " ")

		if len(name) == 1 {
			prefixed += "-" + name
		} else {
			prefixed += "--" + name
		}

		if first {
			first = false
			prefixed += ", "
		}
	}

	return
}
