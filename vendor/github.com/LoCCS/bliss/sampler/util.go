package sampler

func combineUint16(buf []byte, i int) uint16 {
	return uint16(buf[i]) + (uint16(buf[i+1]) << 8)
}

func combineUint64(buf []byte, i int) uint64 {
	return uint64(buf[i]) + (uint64(buf[i+1]) << 8) +
		(uint64(buf[i+2]) << 16) + (uint64(buf[i+3]) << 24) +
		(uint64(buf[i+4]) << 32) + (uint64(buf[i+5]) << 40) +
		(uint64(buf[i+6]) << 48) + (uint64(buf[i+7]) << 56)
}
