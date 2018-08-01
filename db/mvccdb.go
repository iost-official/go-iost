package db

import (
	"github.com/iost-official/Go-IOS-Protocol/db/mvcc"
)

const (
	SEPARATOR = "/"
)

type MVCCDB struct {
	head      *mvcc.TrieNode
	stage     *mvcc.TrieNode
	storage   Database
	revisions map[string]*mvcc.TrieNode
}

func (m *MVCCDB) Get(table string, key string) (string, error) {
	tk := table + SEPARATOR + key
	tv, ok := m.stage.Get(tk, 0)
	if !ok {
		dk := []byte(tk)
		dv, err := m.storage.Get(dk)
		return string(dv[:]), err
	}
	return tv, nil
}

func (m *MVCCDB) Put(table string, key string, value string) error {
	tk := table + SEPARATOR + key
	m.stage.Put(tk, value, 0)
	return nil
}

// TODO Del operation should be a update in mutiple version trie
func (m *MVCCDB) Del(table string, key string) error {
	tk := table + SEPARATOR + key
	_ = m.stage.Del(tk, 0)
	return nil
}

func (m *MVCCDB) Has(table string, key string) (bool, error) {
	return false, nil
}

func (m *MVCCDB) Keys(table string, prefix string) ([]string, error) {
	return nil, nil
}

func (m *MVCCDB) Tables(table string) ([]string, error) {
	return nil, nil
}

func (m *MVCCDB) Commit() (string, error) {
	return "", nil
}

func (m *MVCCDB) Rollback() error {
	return nil
}

func (m *MVCCDB) Tag(tag string) error {
	return nil
}

func (m *MVCCDB) Fork(revision string) *MVCCDB {
	mvccdb := &MVCCDB{
		head:      m.head,
		stage:     m.head.Fork(),
		storage:   m.storage,
		revisions: m.revisions,
	}

	return mvccdb
}

func (m *MVCCDB) Flush(revision string) error {
	return nil
}
