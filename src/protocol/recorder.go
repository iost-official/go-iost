package protocol

import "IOS/src/iosbase"

type Recorder struct {
	*RuntimeData

	txPool iosbase.TxPool
}

func (r *Recorder) init(rd *RuntimeData, bc iosbase.BlockChain, sp iosbase.StatePool) error {
	r.RuntimeData = rd
	r.blockChain = bc
	r.statePool = sp
	return nil
}

func (r *Recorder) verifyTx(tx iosbase.Tx) error {
	return nil
}

func (r *Recorder) verifyBlock(block *iosbase.Block) error {
	return nil
}

func (r *Recorder) makeBlock() *iosbase.Block {
	return &iosbase.Block{}
}

func (r *Recorder) makeEmptyBlock() {
	r.blockChain.Push(iosbase.Block{})
}

func (r *Recorder) admitBlock(block *iosbase.Block) {
}

func (r *Recorder) publishTx(tx iosbase.Tx) {

}

func (r *Recorder) recorderLoop() {
	// every Period : 1/ require new Block; 2/ get view to find out if
}
