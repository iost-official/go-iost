package db

import (
	"os"
	"os/exec"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	DBPATH = "mvccdb"
)

type MVCCDBTestSuite struct {
	suite.Suite
	mvccdb *MVCCDB
}

func (suite *MVCCDBTestSuite) SetupTest() {
	mvccdb, err := NewMVCCDB(DBPATH)
	require.Nil(suite.T(), err, "Create MVCCDB should not fail")
	suite.mvccdb = mvccdb
	suite.mvccdb.Put("table01", "key01", "value01")
	suite.mvccdb.Put("table01", "key02", "value02")
	suite.mvccdb.Put("table01", "key03", "value03")
	suite.mvccdb.Put("table01", "key04", "value04")
	suite.mvccdb.Put("table01", "key05", "value05")
	suite.mvccdb.Put("table01", "iost01", "value06")
	suite.mvccdb.Put("table01", "iost02", "value07")
	suite.mvccdb.Put("table01", "iost03", "value08")
	suite.mvccdb.Put("table01", "iost04", "value09")
	suite.mvccdb.Put("table01", "iost05", "value10")
	suite.mvccdb.Commit()
}

func (suite *MVCCDBTestSuite) TestGet() {
	var value string
	var err error
	value, err = suite.mvccdb.Get("table01", "key01")
	suite.Nil(err)
	suite.Equal("value01", value)
	value, err = suite.mvccdb.Get("table01", "key05")
	suite.Nil(err)
	suite.Equal("value05", value)
}

func (suite *MVCCDBTestSuite) TestPut() {
	var value string
	var err error
	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)

	err = suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
}

func (suite *MVCCDBTestSuite) TestDel() {
	var value string
	var err error
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value04", value)

	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestHas() {
	var ok bool
	var err error
	ok, err = suite.mvccdb.Has("table01", "key01")
	suite.Nil(err)
	suite.True(ok)
	ok, err = suite.mvccdb.Has("table01", "key06")
	suite.Nil(err)
	suite.False(ok)
}

func (suite *MVCCDBTestSuite) TestCommit() {
	var value string
	var err error

	err = suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)
	suite.mvccdb.Commit()

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestRollback() {
	var value string
	var err error

	err = suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)
	suite.mvccdb.Rollback()

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value04", value)
}

func (suite *MVCCDBTestSuite) TestCheckout() {
	var value string
	var err error

	err = suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)
	suite.mvccdb.Commit()
	suite.mvccdb.Tag("tag1")

	err = suite.mvccdb.Put("table01", "key06", "value066")
	suite.Nil(err)
	err = suite.mvccdb.Put("table01", "key04", "value044")
	suite.Nil(err)
	suite.mvccdb.Commit()
	suite.mvccdb.Tag("tag2")

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value066", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value044", value)

	suite.mvccdb.Checkout("tag1")

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestFork() {
	var value string
	var err error

	mvccdb2 := suite.mvccdb.Fork()

	err = mvccdb2.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = mvccdb2.Del("table01", "key04")
	suite.Nil(err)
	mvccdb2.Commit()

	value, err = mvccdb2.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = mvccdb2.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value04", value)
}

func (suite *MVCCDBTestSuite) TestFlush() {
	var value string
	var err error

	err = suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)

	suite.mvccdb.Commit()
	suite.mvccdb.Tag("qwertyuiopasdfghjkl;zxcvbnm,.afd")

	err = suite.mvccdb.Flush("qwertyuiopasdfghjkl;zxcvbnm,.afd")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Equal(ErrKeyNotFound, err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TearDownTest() {
	err := suite.mvccdb.Close()
	suite.Nil(err, "Close MVCCDB should not fail")

	cmd := exec.Command("rm", "-r", DBPATH)
	err = cmd.Run()
	require.Nil(suite.T(), err, "Remove database should not fail")
}

func TestMVCCDBTestSuite(t *testing.T) {
	if (len(os.Args) > 1) && (os.Args[1] == "debug") {
		log.SetLevel(log.DebugLevel)
	}
	suite.Run(t, new(MVCCDBTestSuite))
}
