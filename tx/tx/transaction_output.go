package transaction


// datastructure of transaction's output
// TXOutput 包含两部分
// Value: 有多少币，就是存储在 Value 里面
// ScriptPubKey: 对输出进行锁定
// 在当前实现中，ScriptPubKey 将仅用一个字符串来代替
type TXOutput struct{
	Value int
	ScriptPubKey string
}


// list of transaction's output
type TXOutputs struct{
	Outputs []TXOutput
}


func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}