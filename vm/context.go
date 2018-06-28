package vm

type Context struct {
	Base        *Context
	Publisher   IOSTAccount
	Signers     []IOSTAccount
	ParentHash  []byte
	Timestamp   int64
	BlockHeight int64
	Witness     IOSTAccount
}

func NewContext(ctx *Context) *Context {
	return &Context{
		Base: ctx,
	}
}

func BaseContext() *Context {
	return &Context{Base: nil}
}
