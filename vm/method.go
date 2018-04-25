package vm

type Method struct {
	name      string
	code      Code
	owner     Pubkey
	privilege Privilege

	contract *Contract
}
