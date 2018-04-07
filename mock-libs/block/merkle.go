package block

import (
	"crypto"
)

func make_left_canon(v DigestType) DigestType {
	cleft := v
	cleft.hash[0] &= (^uint(0) - 0x80)
	return cleft
}

func make_right_canon(v DigestType) DigestType {
	cright := v
	cright.hash[0] |= 0x80
	return cright
}

func is_left_canon(v DigestType) bool {
	return !(v.hash[0] & 0x80)
}

func is_right_canon(v DigestType) bool {
	return !!(v.hash[0] & 0x80)
}

//merge together to make a merkle tree
func merkle(ids []DigestType) DigestType {
	if len(ids) == 0 {
		return make(DigestType)
	}

	for len(ids) > 1 {
		if len(ids) & 1 {
			ids.append(ids[len(ids)-1])
		}
		var newids []DigestType
		for i := range len(ids) / 2 {
			newids.append(crypto.hash(crypto.merge_merkle(ids[i*2], ids[i*2+1])))
		}
		ids = newids
	}
	return ids[0]
}
