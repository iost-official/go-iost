package pob

import (
	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/consensus/common"
	. "github.com/iost-official/Go-IOS-Protocol/core/tx"

	"errors"
	"fmt"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"

	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
	"encoding/binary"
)

func genGenesis(initTime int64) (*block.Block, error) {

	main := lua.NewMethod(vm.Public, "", 0, 0)

	var code string
	for k, v := range GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	lc := lua.NewContract(vm.ContractInfo{Prefix: "", GasLimit: 0, Price: 0, Publisher: ""}, code, main)

	tx := Tx{
		Time:     0,
		Nonce:    0,
		Contract: &lc,
	}

	genesis := &block.Block{
		Head: block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Content: make([]Tx, 0),
	}
	genesis.Content = append(genesis.Content, tx)
	return genesis, nil
}

func genBlock(acc Account, bc block.Chain, pool state.Pool) *block.Block {
	lastBlk := bc.Top()
	blk := block.Block{Content: []Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: lastBlk.HeadHash(),
		Number:     lastBlk.Head.Number + 1,
		Witness:    acc.ID,
		Time:       GetCurrentTimestamp().Slot,
	}}

	vc := vm.NewContext(vm.BaseContext())
	vc.Timestamp = blk.Head.Time
	vc.ParentHash = blk.Head.ParentHash
	vc.BlockHeight = blk.Head.Number
	vc.Witness = vm.IOSTAccount(acc.ID)

	// TODO
	blk.Head.TreeHash = blk.CalculateTreeHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()

	blockcache.CleanStdVerifier()

	generatedBlockCount.Inc()

	Data.ClearServi(blk.Head.Witness)

	return &blk
}

func generateHeadInfo(head block.BlockHead) []byte {
	var info, numberInfo, versionInfo []byte
	info = make([]byte, 8)
	versionInfo = make([]byte, 4)
	numberInfo = make([]byte, 4)
	binary.BigEndian.PutUint64(info, uint64(head.Time))
	binary.BigEndian.PutUint32(versionInfo, uint32(head.Version))
	binary.BigEndian.PutUint32(numberInfo, uint32(head.Number))
	info = append(info, versionInfo...)
	info = append(info, numberInfo...)
	info = append(info, head.ParentHash...)
	info = append(info, head.TreeHash...)
	info = append(info, head.Info...)
	return common.Sha256(info)
}

func (p *PoB) blockVerify(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
	// verify block head
	if err := blockcache.VerifyBlockHead(blk, parent); err != nil {
		return nil, err
	}

	// verify block witness
	if witnessOfTime(Timestamp{Slot: blk.Head.Time}) != blk.Head.Witness {
		return nil, errors.New("wrong witness")
	}

	headInfo := generateHeadInfo(blk.Head)
	var signature common.Signature
	signature.Decode(blk.Head.Signature)

	if blk.Head.Witness != common.Base58Encode(signature.Pubkey) {
		return nil, errors.New("wrong pubkey")
	}

	// verify block witness signature
	if !common.VerifySignature(headInfo, signature) {
		return nil, errors.New("wrong signature")
	}
	newPool, err := blockcache.StdBlockVerifier(blk, pool)
	if err != nil {
		return nil, err
	}
	return newPool, nil
}