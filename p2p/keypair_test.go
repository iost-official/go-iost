package p2p

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testPrivKey    = "CAESYAQITj2STtBfy+bszLOrp3XvgMmCVdv32X7IPPiHPPdJ0MH8a3hlN8U1GCWynchgcW+IeUWBLLwz7PIzHwCHT1DQwfxreGU3xTUYJbKdyGBxb4h5RYEsvDPs8jMfAIdPUA=="
	testInvalidKey = "abcde"
	testCacheFile  = "testkey.cache"
)

func TestGetKeyFromFile(t *testing.T) {
	// valid key in file
	err := os.WriteFile(testCacheFile, []byte(testPrivKey), 0666)
	assert.Nil(t, err)
	pk, err := getKeyFromFile(testCacheFile)
	assert.Nil(t, err)
	assert.NotNil(t, pk)
	assert.Nil(t, os.Remove(testCacheFile))

	// invalid key in file
	err = os.WriteFile(testCacheFile, []byte(testInvalidKey), 0666)
	assert.Nil(t, err)
	pk, err = getKeyFromFile(testCacheFile)
	assert.NotNil(t, err)
	assert.Nil(t, pk)
	assert.Nil(t, os.Remove(testCacheFile))

	// nonexistent file
	pk, err = getKeyFromFile("")
	assert.NotNil(t, err)
	assert.Nil(t, pk)

	// nonexistent file name
	pk, err = getKeyFromFile(testCacheFile)
	assert.NotNil(t, err)
	assert.Nil(t, pk)
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
	// file should be nonexistent
	isExist, err := fileExists(testCacheFile)
	assert.Nil(t, err)
	assert.False(t, isExist)

	// create key and save
	privKey, err := getOrCreateKey(testCacheFile)
	assert.Nil(t, err)
	assert.NotNil(t, privKey)

	// file should be existent
	isExist, err = fileExists(testCacheFile)
	assert.Nil(t, err)
	assert.True(t, isExist)

	assert.Nil(t, os.Remove(testCacheFile))
}
