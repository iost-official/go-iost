package vm

type Context struct {
	Base        *Context
	Publisher   string
	Signers     []string
	ParentHash  []byte
	Timestamp   int64
	BlockHeight int64
	Witness     string
}

func NewContext(ctx *Context) *Context {
	return &Context{
		Base: ctx,
	}
}

func BaseContext() *Context {
	return &Context{Base: nil}
}
