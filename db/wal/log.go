package wal

import "errors"

func (log *Log) Check(crc uint64) error {
	if log.Checksum == crc {
		return nil
	}
	log.Reset()
	return errors.New("WAL: crc mismatch")
}