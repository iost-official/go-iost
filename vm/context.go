package vm

type Context struct {
	Base       *Context
	Publisher  IOSTAccount
	Signers    []IOSTAccount
	ParentHash []byte
	Timestamp  int64
}

func NewContext(ctx *Context) *Context {
	return &Context{
		Base: ctx,
	}
}

//func (b Context) ParentHash() []byte {
//	return []byte{0}
//}
//
//func (b Context) Base() *Context {
//	return b.base
//}

func BaseContext() *Context {
	return &Context{}
}
