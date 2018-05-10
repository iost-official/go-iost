package message

//go:generate gencode go -schema=structs.schema -package=message


func (d *Message) GetTime() int64 {
	return d.Time
}

func (d *Message) GetFrom() string {
	return d.From
}

func (d *Message) GetTo() string {
	return d.To
}

func (d *Message) GetReqType() int32 {
	return d.ReqType
}

func (d *Message) GetPriority() int8 {
	return d.Priority
}

func (d *Message) GetBody() []byte {
	return d.Body
}
