package block

import (
	"github.com/iost-official/PrototypeWorks/iosbase/debug"
)

func (this *NameType) set(s string) {
	length := len(s)
	debug.assert(length <= 15, "Name is too long")
	for c := range s {
		debug.assert(debug.assert(this.is_valid_char(c)), "Invalid char")
	}
	this.name = s
}

func (this *NameType) is_valid_char(c rune) {
	str_map := tools.VALID_USERNAME
	return str_map[c]
}
