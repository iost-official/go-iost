package protocol

import "IOS/src/iosbase"

type DposView struct {
	primary iosbase.Member
	backup  []iosbase.Member
}

func NewDposView(chain iosbase.BlockChain) DposView {
	var view DposView

	return view
}

func (v *DposView) GetPrimaryID() string {
	return v.primary.ID
}

func (v *DposView) GetBackupID() []string {
	var s []string
	for _, m := range v.backup {
		s = append(s, m.ID)
	}
	return s
}

func (v *DposView) isPrimary(ID string) bool {
	if ID == v.primary.ID {
		return true
	} else {
		return false
	}
}

func (v *DposView) isBackup(ID string) bool {
	ans := false
	for _, m := range v.backup {
		if ID == m.ID {
			ans = true
		}
	}
	return ans
}
