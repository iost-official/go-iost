package vm

type BlockValue interface{}

type Program struct {
	Owner  string
	Script string
}

func Exec(program Program) (oldState, newState map[string]BlockValue) {
	return nil, nil
}

func getStatus(addr, key string) (BlockValue, error) {
	return nil, nil
}
