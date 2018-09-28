package db

import (
	"os/exec"
	"testing"

	"os"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	DBPATH = "mvccdb"
)

type MVCCDBTestSuite struct {
	suite.Suite
	mvccdb MVCCDB
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
	suite.Nil(err)
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
	suite.Nil(err)
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

	err := suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)
	suite.mvccdb.Commit()

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestRollback() {
	var value string

	err := suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)
	suite.mvccdb.Rollback()

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value04", value)
}

func (suite *MVCCDBTestSuite) TestCheckout() {
	var value string

	err := suite.mvccdb.Put("table01", "key06", "value06")
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
	suite.Nil(err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestFork() {
	var value string

	mvccdb2 := suite.mvccdb.Fork()

	err := mvccdb2.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = mvccdb2.Del("table01", "key04")
	suite.Nil(err)
	mvccdb2.Commit()

	value, err = mvccdb2.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = mvccdb2.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("", value)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("value04", value)
}

func (suite *MVCCDBTestSuite) TestFlush() {
	var value string

	err := suite.mvccdb.Put("table01", "key06", "value06")
	suite.Nil(err)
	err = suite.mvccdb.Del("table01", "key04")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("", value)

	suite.mvccdb.Commit()
	suite.mvccdb.Tag("qwertyuiopasdfghjkl;zxcvbnm,.afd")

	err = suite.mvccdb.Flush("qwertyuiopasdfghjkl;zxcvbnm,.afd")
	suite.Nil(err)

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
	suite.Equal("", value)
}

func (suite *MVCCDBTestSuite) TestRecovery() {
	var value string
	var err error

	value, err = suite.mvccdb.Get("table01", "key05")
	suite.Nil(err)
	suite.Equal("value05", value)

	suite.mvccdb.Commit()
	suite.mvccdb.Tag("qwertyuiopasdfghjkl;zxcvbnm,.afd")

	err = suite.mvccdb.Flush("qwertyuiopasdfghjkl;zxcvbnm,.afd")
	suite.Nil(err)

	ilog.Info("Close mvccdb")
	err = suite.mvccdb.Close()
	suite.Nil(err, "Close MVCCDB should not fail")

	mvccdb, err := NewMVCCDB(DBPATH)
	require.Nil(suite.T(), err, "Create MVCCDB should not fail")
	suite.mvccdb = mvccdb

	value, err = suite.mvccdb.Get("table01", "key05")
	suite.Nil(err)
	suite.Equal("value05", value)

	suite.Equal("qwertyuiopasdfghjkl;zxcvbnm,.afd", suite.mvccdb.CurrentTag())
}

func (suite *MVCCDBTestSuite) TestFlushAndCheckout() {
	var value string

	err := suite.mvccdb.Put("table01", "key06", "value06")
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

	err = suite.mvccdb.Flush("tag1")
	suite.Nil(err)
	suite.mvccdb.Checkout("tag1")

	value, err = suite.mvccdb.Get("table01", "key06")
	suite.Nil(err)
	suite.Equal("value06", value)
	value, err = suite.mvccdb.Get("table01", "key04")
	suite.Nil(err)
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
	suite.Run(t, new(MVCCDBTestSuite))
}

func TestPutTimeout(t *testing.T) {
	t.Skip() // todo repair or find out why
	d, err := NewMVCCDB("mvcc_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.Close()
		os.RemoveAll("mvcc_test")
	}()
	start := time.Now()
	d.Put("t", "a", "b")
	du := time.Now().Sub(start)
	if du > 0 {
		t.Error("time usage: ", du)
	}

	start = time.Now()
	d.Put("t", "b", "c")
	du = time.Now().Sub(start)
	if du > 0 {
		t.Error("time usage: ", du)
	}
}
