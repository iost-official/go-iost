package p2p

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	testPrivKey    = "CAESYAQITj2STtBfy+bszLOrp3XvgMmCVdv32X7IPPiHPPdJ0MH8a3hlN8U1GCWynchgcW+IeUWBLLwz7PIzHwCHT1DQwfxreGU3xTUYJbKdyGBxb4h5RYEsvDPs8jMfAIdPUA=="
	testInvalidKey = "abcde"
	testCacheFile  = "testkey.cache"
)

func TestGetKeyFromFile(t *testing.T) {
	Convey("valid key in file", t, func() {
		err := ioutil.WriteFile(testCacheFile, []byte(testPrivKey), 0666)
		Convey("create a file", func() {
			So(err, ShouldBeNil)
		})

		Convey("check private key", func() {
			pk, err := getKeyFromFile(testCacheFile)
			So(err, ShouldBeNil)
			So(pk, ShouldNotBeNil)
		})

		err = os.Remove(testCacheFile)
		Convey("delete the file", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("invalid key in file", t, func() {
		err := ioutil.WriteFile(testCacheFile, []byte(testInvalidKey), 0666)
		Convey("create a file", func() {
			So(err, ShouldBeNil)
		})

		Convey("check private key", func() {
			pk, err := getKeyFromFile(testCacheFile)
			So(err, ShouldNotBeNil)
			So(pk, ShouldBeNil)
		})

		err = os.Remove(testCacheFile)
		Convey("delete the file", func() {
			So(err, ShouldBeNil)
		})
	})

	Convey("nonexistent file", t, func() {
		Convey("empty file name", func() {
			pk, err := getKeyFromFile("")
			So(err, ShouldNotBeNil)
			So(pk, ShouldBeNil)
		})
		Convey("nonexistent file name", func() {
			pk, err := getKeyFromFile(testCacheFile)
			So(err, ShouldNotBeNil)
			So(pk, ShouldBeNil)
		})
	})
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func TestGetOrCreateKey(t *testing.T) {
	Convey("file should be nonexistent", t, func() {
		isExist, err := fileExists(testCacheFile)
		So(err, ShouldBeNil)
		So(isExist, ShouldBeFalse)
	})
	Convey("create key and save", t, func() {
		privKey, err := getOrCreateKey(testCacheFile)
		So(err, ShouldBeNil)
		So(privKey, ShouldNotBeNil)
	})

	Convey("file should be existent", t, func() {
		isExist, err := fileExists(testCacheFile)
		So(err, ShouldBeNil)
		So(isExist, ShouldBeTrue)
	})

	Convey("delete the file", t, func() {
		err := os.Remove(testCacheFile)
		So(err, ShouldBeNil)
	})
}
