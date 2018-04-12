package iostdb

import (
	"errors"

	"github.com/gomodule/redigo/redis"
)

const (
	Conn   = "tcp"
	DBAddr = "localhost:6379"
)

type RedisDatabase struct {
	cli redis.Conn
}

func NewRedisDatabase() (*RedisDatabase, error) {
	dial, _ := redis.Dial(Conn, DBAddr)
	return &RedisDatabase{cli: dial}, nil
}

func (rdb *RedisDatabase) Put(args ...interface{}) error {
	_, err := rdb.cli.Do("HMSET", args...)
	return err
}

func (rdb *RedisDatabase) Get(args ...interface{}) (interface{}, error) {
	value, ok := redis.Values(rdb.cli.Do("HMGET", args...))
	if ok == nil {
		return value, nil
	}
	return nil, errors.New("Not found")
}

func (rdb *RedisDatabase) Has(args ...interface{}) (bool, error) {
	_, ok := redis.Values(rdb.cli.Do("HMGET", args...))
	return ok == nil, nil
}

func (rdb *RedisDatabase) Delete(key interface{}) error {
	_, err := rdb.cli.Do("DEL", key)
	return err
}

func (rdb *RedisDatabase) Close() {
	rdb.cli = nil
}

type UTXORedis struct {
	db      *RedisDatabase
	subKeys []interface{}
}

func NewUTXORedis(keys ...interface{}) (*UTXORedis, error) {
	rdb, _ := NewRedisDatabase()
	return &UTXORedis{db: rdb, subKeys: keys}, nil
}

func (ur *UTXORedis) Put(args ...interface{}) error {
	params := make([]interface{}, 0)
	for k, v := range args {
		if k != 0 {
			params = append(params, ur.subKeys[k-1])
		}
		params = append(params, v)
	}
	ur.db.Put(params...)
	return nil
}

func (ur *UTXORedis) Get(args ...interface{}) (interface{}, error) {
	params := make([]interface{}, 0)
	for _, v := range args {
		params = append(params, v)
	}
	for _, v := range ur.subKeys {
		params = append(params, v)
	}
	rtn, err := ur.db.Get(params...)
	return rtn, err

}

func (ur *UTXORedis) Has(args ...interface{}) (bool, error) {
	params := make([]interface{}, 0)
	for _, v := range args {
		params = append(params, v)
	}
	for _, v := range ur.subKeys {
		params = append(params, v)
	}
	return ur.db.Has(params...)
}
