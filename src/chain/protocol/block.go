package protocol

import "common"

type blockHeader struct{
	previous common.BlockIdType
	timestamp uint32
	witness
}

type SignedBlock struct {

}
