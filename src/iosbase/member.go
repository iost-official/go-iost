package iosbase

type Member struct {
	ID     string
	Pubkey []byte
	Seckey []byte
}

func NewMember(seckey []byte) (Member, error) {
	var m Member
	m.Seckey = seckey
	m.Pubkey = CalcPubkey(seckey)
	m.ID = Base58Encode(m.Pubkey)
	return m, nil
}
