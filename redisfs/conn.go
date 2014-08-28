package redisfs

import "strconv"
import "github.com/garyburd/redigo/redis"

func NewRedisConn(host string, port int, auth string) (redis.Conn, error) {
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
