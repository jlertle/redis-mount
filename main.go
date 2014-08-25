package main

import "os"
import "fmt"
import "strings"
import "strconv"
import "path/filepath"
import "github.com/poying/go-chalk"
import "github.com/codegangsta/cli"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/garyburd/redigo/redis"
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

	conn, err := newRedisConn(
		ctx.String("host"),
		ctx.Int("port"),
		ctx.String("auth"))

	if err != nil {
		return nil, err
	}

	fs := &redisfs.RedisFs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		Conn:       conn,
		Dirs:       make(map[string][]string),
		Sep:        ctx.String("sep"),
	}

	if err != nil {
		return nil, err
	}

	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mnt, nfs.Root(), nil)

	if err != nil {
		return nil, err
	}

	return server, nil
}

func newRedisConn(host string, port int, auth string) (redis.Conn, error) {
	address := host + ":" + strconv.Itoa(port)
	conn, err := redis.Dial("tcp", address)

	if err != nil {
		return nil, err
	}

	if len(auth) > 0 {
		if _, err := conn.Do("AUTH", auth); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
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
