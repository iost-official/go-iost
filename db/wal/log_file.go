package wal

type LogFile struct {
	// dir to put files
	dir string
	// size of files to make, in bytes
	size int64

	streamFile StreamFile
	encoder    encoder
	decoder    decoder

}