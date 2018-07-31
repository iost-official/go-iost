package chainbase

type BTreeNode struct {
}

type Chainbase struct {
	commits map[string]*BTreeNode
}

func (m *Chainbase) Get(table string, key string) (string, error) {
	return "", nil
}

func (m *Chainbase) Put(table string, value string) error {
	return nil
}

func (m *Chainbase) Del(table string, key string) error {
	return nil
}

func (m *Chainbase) Has(table string, key string) (bool, error) {
	return false, nil
}

func (m *Chainbase) Keys(table string, prefix string) ([]string, error) {
	return nil, nil
}

func (m *Chainbase) Tables(table string) ([]string, error) {
	return nil, nil
}

func (m *Chainbase) Commit() (string, error) {
	return "", nil
}

func (m *Chainbase) Rollback() error {
	return nil
}

func (m *Chainbase) Tag(tag string) error {
	return nil
}

func (m *Chainbase) Checkout(commit string) (*Chainbase, error) {
	return nil, nil
}

func (m *Chainbase) Flush(commit string) error {
	return nil
}
