package iost

type Serializable interface {
	Bytes() []byte
}

type Binary struct {
	bytes  []byte
	length int
}

func BinaryFromBase58(s string) Binary {
	bits := Base58Decode(s)
	return Binary{bits, len(bits)}
}

func (bin *Binary) ToBase58() string {
	return Base58Encode(bin.bytes)
}

func (bin *Binary) putUInt(u uint32) *Binary {
	var bytes [4]byte
	for i := 0; i < 4; i++ {
		bytes[i] = byte((u >> uint32((3-i)*8)) & 0xff)
	}
	bin.bytes = append(bin.bytes, bytes[:]...)
	bin.length += 4
	return bin
}

func (bin *Binary) putULong(u uint64) *Binary {
	var bytes [8]byte
	for i := 0; i < 8; i++ {
		bytes[i] = byte((u >> uint32((7-i)*8)) & 0xff)
	}
	bin.bytes = append(bin.bytes, bytes[:]...)
	bin.length += 8
	return bin
}

func (bin *Binary) putBytes(b []byte) *Binary {
	bin.bytes = append(bin.bytes, b...)
	bin.length += len(b)
	return bin
}

func (bin *Binary) put(o Serializable) *Binary {
	bytes := o.Bytes()
	bin.bytes = append(bin.bytes, bytes...)
	bin.length += len(bytes)
	return bin
}

func (bin *Binary) Bytes() []byte {
	return bin.bytes
}

func (bin Binary) Hash160() Binary {
	bin.bytes = Hash160(bin.bytes)
	return bin
}

func Equal(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
