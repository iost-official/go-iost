package tx

/**
 * Describtion: tx
 * User: wangyu
 * Date: 18-7-30
 */

// Action 的实现
type Action struct {
	Contract    string  // 合约地址，为空则视为调用系统合约
	ActionName  string	// 方法名称
	Data        string  // json
}



