package vm

type Method struct {
	Name      string
	Code      Code
	Owner     Pubkey
	Contract *Contract
}
