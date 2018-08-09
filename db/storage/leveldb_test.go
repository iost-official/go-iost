package storage

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	DBPATH = "leveldb"
)

type LevelDBTestSuite struct {
	suite.Suite
	ldb *LevelDB
}

func (suite *LevelDBTestSuite) SetupTest() {
	ldb, err := NewLevelDB(DBPATH)
	suite.Require().Nil(err)
	suite.ldb = ldb
	suite.ldb.Put([]byte("key01"), []byte("value01"))
	suite.ldb.Put([]byte("key02"), []byte("value02"))
	suite.ldb.Put([]byte("key03"), []byte("value03"))
	suite.ldb.Put([]byte("key04"), []byte("value04"))
	suite.ldb.Put([]byte("key05"), []byte("value05"))
	suite.ldb.Put([]byte("iost01"), []byte("value06"))
	suite.ldb.Put([]byte("iost02"), []byte("value07"))
	suite.ldb.Put([]byte("iost03"), []byte("value08"))
	suite.ldb.Put([]byte("iost04"), []byte("value09"))
	suite.ldb.Put([]byte("iost05"), []byte("value10"))
}

func (suite *LevelDBTestSuite) TestGet() {
	var value []byte
	var err error
	value, err = suite.ldb.Get([]byte("key00"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)
	value, err = suite.ldb.Get([]byte("key01"))
	suite.Nil(err)
	suite.Equal([]byte("value01"), value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
}

func (suite *LevelDBTestSuite) TestPut() {
	var value []byte
	var err error
	value, err = suite.ldb.Get([]byte("key06"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)
	value, err = suite.ldb.Get([]byte("key07"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)

	err = suite.ldb.Put([]byte("key07"), []byte("value07"))
	suite.Nil(err)

	value, err = suite.ldb.Get([]byte("key06"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)
	value, err = suite.ldb.Get([]byte("key07"))
	suite.Nil(err)
	suite.Equal([]byte("value07"), value)
}

func (suite *LevelDBTestSuite) TestDel() {
	var value []byte
	var err error
	value, err = suite.ldb.Get([]byte("key04"))
	suite.Nil(err)
	suite.Equal([]byte("value04"), value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)

	err = suite.ldb.Del([]byte("key04"))
	suite.Nil(err)

	value, err = suite.ldb.Get([]byte("key04"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Equal([]byte{}, value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
}

func (suite *LevelDBTestSuite) TestKeys() {
	var keys [][]byte
	var err error

	keys, err = suite.ldb.Keys([]byte("key"))
	suite.Nil(err)
	suite.ElementsMatch(
		[][]byte{
			[]byte("key01"),
			[]byte("key02"),
			[]byte("key03"),
			[]byte("key04"),
			[]byte("key05"),
		},
		keys,
	)
	keys, err = suite.ldb.Keys([]byte("iost"))
	suite.Nil(err)
	suite.ElementsMatch(
		[][]byte{
			[]byte("iost01"),
			[]byte("iost02"),
			[]byte("iost03"),
			[]byte("iost04"),
			[]byte("iost05"),
		},
		keys,
	)
}

func (suite *LevelDBTestSuite) TestBatch() {
	var value []byte
	var err error

	value, err = suite.ldb.Get([]byte("key04"))
	suite.Nil(err)
	suite.Equal([]byte("value04"), value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
	value, err = suite.ldb.Get([]byte("key06"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)

	err = suite.ldb.BeginBatch()
	suite.Nil(err)
	err = suite.ldb.BeginBatch()
	suite.NotNil(err)

	err = suite.ldb.Del([]byte("key04"))
	suite.Nil(err)
	err = suite.ldb.Put([]byte("key06"), []byte("value06"))
	suite.Nil(err)

	err = suite.ldb.CommitBatch()
	suite.Nil(err)
	err = suite.ldb.CommitBatch()
	suite.NotNil(err)

	value, err = suite.ldb.Get([]byte("key04"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Equal([]byte{}, value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
	value, err = suite.ldb.Get([]byte("key06"))
	suite.Nil(err)
	suite.Equal([]byte("value06"), value)
}

func (suite *LevelDBTestSuite) TestRecover() {
	var value []byte
	var err error

	value, err = suite.ldb.Get([]byte("key04"))
	suite.Nil(err)
	suite.Equal([]byte("value04"), value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
	value, err = suite.ldb.Get([]byte("key06"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)

	err = suite.ldb.BeginBatch()
	suite.Nil(err)

	err = suite.ldb.Del([]byte("key04"))
	suite.Nil(err)
	err = suite.ldb.Put([]byte("key06"), []byte("value06"))
	suite.Nil(err)

	err = suite.ldb.Close()
	suite.Nil(err)
	ldb, err := NewLevelDB(DBPATH)
	suite.Require().Nil(err)
	suite.ldb = ldb

	err = suite.ldb.CommitBatch()
	suite.NotNil(err)

	value, err = suite.ldb.Get([]byte("key04"))
	suite.Nil(err)
	suite.Equal([]byte("value04"), value)
	value, err = suite.ldb.Get([]byte("key05"))
	suite.Nil(err)
	suite.Equal([]byte("value05"), value)
	value, err = suite.ldb.Get([]byte("key06"))
	suite.Equal(leveldb.ErrNotFound, err)
	suite.Nil(value)
}

func (suite *LevelDBTestSuite) TearDownTest() {
	err := suite.ldb.Close()
	suite.Nil(err)
	cmd := exec.Command("rm", "-r", DBPATH)
	err = cmd.Run()
	suite.Require().Nil(err)
}

func TestLevelDBTestSuite(t *testing.T) {
	suite.Run(t, new(LevelDBTestSuite))
}
