package pbft

import "IOS/src/iosbase"

type View struct {
	primary iosbase.Member
	backup  []iosbase.Member
}


func GetViewID(chain iosbase.BlockChain) int {
	return 0
}

func NewView(chain *iosbase.BlockChain) View {
	var view View
	return view
}

func (v *View) GetPrimaryID() string{
	return v.primary.ID
}

func (v *View) GetBackupID() []string{
	var s  []string
	for _, m := range v.backup {
		s = append(s, m.ID)
	}
	return s
}
