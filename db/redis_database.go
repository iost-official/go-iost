package db

import (
	"strconv"
	"time"

	"errors"

	"github.com/gomodule/redigo/redis"
)

const (
	Conn = "tcp"
)

var DBAddr string = "127.0.0.1"
var DBPort int16 = 6379

type RedisDatabase struct {
	connPool *redis.Pool
}

func NewRedisDatabase() (*RedisDatabase, error) {

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(Conn, DBAddr+":"+strconv.FormatUint(uint64(DBPort), 10))
		},
	}

	return &RedisDatabase{connPool: pool}, nil
}

func (rdb *RedisDatabase) Put(key []byte, value []byte) error {
	conn := rdb.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", interface{}(key), interface{}(value))
	return err
}

func (rdb *RedisDatabase) PutHM(key []byte, args ...[]byte) error {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = key
	for i, v := range args {
		newArgs[i+1] = v
	}
	conn := rdb.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("HMSET", newArgs...)
	return err
}

func (rdb *RedisDatabase) Get(key []byte) ([]byte, error) {
	conn := rdb.connPool.Get()
	defer conn.Close()
	rtn, err := conn.Do("GET", interface{}(key))
	if err != nil {

		return nil, err
	}
	if rtn == nil {
		return nil, nil
	}
	return rtn.([]byte), err
}

func (rdb *RedisDatabase) GetHM(key []byte, args ...[]byte) ([][]byte, error) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = key
	for i, v := range args {
		newArgs[i+1] = v
	}
	conn := rdb.connPool.Get()
	defer conn.Close()
	value, ok := redis.Values(conn.Do("HMGET", newArgs...))
	if ok == nil {
		params := make([][]byte, 0)
		for _, v := range value {
			if v == nil {
				params = append(params, nil)
			} else {
				params = append(params, v.([]byte))
			}
		}
		return params, nil
	}
	return nil, nil
}

func (rdb *RedisDatabase) Has(key []byte) (bool, error) {
	conn := rdb.connPool.Get()
	defer conn.Close()
	_, ok := conn.Do("EXISTS", key)
	return ok == nil, nil
}

func (rdb *RedisDatabase) Delete(key []byte) error {
	conn := rdb.connPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

func (rdb *RedisDatabase) Close() {
	rdb.connPool.Close()
}

func (rdb *RedisDatabase) Type(key string) (string, error) {
	conn := rdb.connPool.Get()
	defer conn.Close()
	rtn, err := conn.Do("TYPE", key)
	if err != nil {
		return "", err
	}
	s, ok := rtn.(string)
	if !ok {
		return "", errors.New("return no string")
	}
	return s, nil
}

func (rdb *RedisDatabase) GetAll(key string) (map[string]string, error) {
	conn := rdb.connPool.Get()
	defer conn.Close()
	return redis.StringMap(conn.Do("HGETALL", key))
}
