package db

import (
	"strconv"

	"github.com/gomodule/redigo/redis"
)

const (
	Conn = "tcp"
)

var DBAddr string = "127.0.0.1"
var DBPort int16 = 6379

type RedisDatabase struct {
	cli redis.Conn
}

func NewRedisDatabase() (*RedisDatabase, error) {
	//fmt.Println(strings.TrimSpace(DBAddr),":",strconv.FormatUint(uint64(DBPort),10))
	dial, err := redis.Dial(Conn, DBAddr+":"+strconv.FormatUint(uint64(DBPort), 10))

	return &RedisDatabase{cli: dial}, err
}

func (rdb *RedisDatabase) Put(key []byte, value []byte) error {
	_, err := rdb.cli.Do("SET", interface{}(key), interface{}(value))
	return err
}

func (rdb *RedisDatabase) PutHM(key []byte, args ...[]byte) error {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = key
	for i, v := range args {
		newArgs[i+1] = v
	}
	_, err := rdb.cli.Do("HMSET", newArgs...)
	return err
}

func (rdb *RedisDatabase) Get(key []byte) ([]byte, error) {
	rtn, err := rdb.cli.Do("GET", interface{}(key))
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
	value, ok := redis.Values(rdb.cli.Do("HMGET", newArgs...))
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
	_, ok := rdb.cli.Do("EXISTS", key)
	return ok == nil, nil
}

func (rdb *RedisDatabase) Delete(key []byte) error {
	_, err := rdb.cli.Do("DEL", key)
	return err
}

func (rdb *RedisDatabase) Close() {
	rdb.cli = nil
}
