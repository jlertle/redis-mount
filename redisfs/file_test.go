package redisfs_test

import "time"
import "errors"
import "testing"
import "github.com/hanwen/go-fuse/fuse"
import "github.com/garyburd/redigo/redis"
import "github.com/poying/redis-mount/redisfs"
import . "github.com/smartystreets/goconvey/convey"

func TestRedisFile(t *testing.T) {
	Convey("Write", t, func() {
		conn := dialTestDB()

		Convey("should work", func() {
			data := []byte("Ghost Island Taiwan")

			_, err := conn.Do("SET", "writing", "")

			file := redisfs.NewRedisFile(conn, "writing")
			_, code := file.Write(data, 0)

			So(code, ShouldEqual, fuse.OK)
			res, err := redis.String(conn.Do("GET", "writing"))

			if err != nil {
				panic(err)
			}

			So(res, ShouldEqual, string(data))
		})

		Reset(func() {
			conn.Close()
		})
	})

	Convey("Read", t, func() {
		conn := dialTestDB()

		Convey("should work", func() {
			file := redisfs.NewRedisFile(conn, "reading")
			data := []byte("QQ")
			_, err := conn.Do("SET", "reading", string(data))

			if err != nil {
				panic(err)
			}

			buf := make([]byte, 100)
			res, code := file.Read(buf, 0)

			So(code, ShouldEqual, fuse.OK)
			So(res.Size(), ShouldEqual, len(data))
		})

		Reset(func() {
			conn.Close()
		})
	})
}

// https://github.com/garyburd/redigo/blob/master/redis/test_test.go

type testConn struct {
	redis.Conn
}

func (t testConn) Close() error {
	_, err := t.Conn.Do("SELECT", "9")
	if err != nil {
		return nil
	}
	_, err = t.Conn.Do("FLUSHDB")
	if err != nil {
		return err
	}
	return t.Conn.Close()
}

func dialTestDB() redis.Conn {
	c, err := redis.DialTimeout("tcp", ":6379", 0, 1*time.Second, 1*time.Second)
	if err != nil {
		panic(err)
	}

	_, err = c.Do("SELECT", "9")
	if err != nil {
		panic(err)
	}

	n, err := redis.Int(c.Do("DBSIZE"))
	if err != nil {
		panic(err)
	}

	if n != 0 {
		panic(errors.New("database #9 is not empty, test can not continue"))
	}

	return testConn{c}
}
