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

func (rdb *RedisDatabase) Put(key string, args ...string) error {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = key
	for i, v := range args {
		newArgs[i+1] = v
	}
	_, err := rdb.cli.Do("HMSET", newArgs...)
	return err
}

func (rdb *RedisDatabase) Get(key string, args ...string) (interface{}, error) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = key
	for i, v := range args {
		newArgs[i+1] = v
	}
	value, ok := redis.Values(rdb.cli.Do("HMGET", newArgs...))
	if ok == nil {
		return value, nil
	}
	return nil, errors.New("Not found")
}

func (rdb *RedisDatabase) Has(key string) (bool, error) {
	_, ok := rdb.cli.Do("EXISTS", key)
	return ok == nil, nil
}

func (rdb *RedisDatabase) Delete(key string) error {
	_, err := rdb.cli.Do("DEL", key)
	return err
}

func (rdb *RedisDatabase) Close() {
	rdb.cli = nil
}

/*
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
}*/
