package db

import (
	"errors"

	"github.com/iost-official/Go-IOS-Protocol/db/mvcc"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	SEPARATOR = '/'
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrTableNotValid = errors.New("table name is not valid")
)

type MVCCDB struct {
	head      *mvcc.TrieNode
	stage     *mvcc.TrieNode
	storage   Database
	revisions map[string]*mvcc.TrieNode
}

func (m *MVCCDB) isValidTable(table string) bool {
	for _, b := range table {
		if b == SEPARATOR {
			return false
		}
	}
	return true
}

func (m *MVCCDB) Get(table string, key string) (string, error) {
	if !m.isValidTable(table) {
		return "", ErrTableNotValid
	}
	tk := []byte(table + string(SEPARATOR) + key)
	tv, ok := m.stage.Get(tk, 0)
	if !ok {
		dk := tk
		dv, err := m.storage.Get(dk)
		return string(dv[:]), err
	}
	if tv.Deleted {
		return "", ErrKeyNotFound
	}
	return string(tv.Raw[:]), nil
}

func (m *MVCCDB) Put(table string, key string, value string) error {
	if !m.isValidTable(table) {
		return ErrTableNotValid
	}
	tk := []byte(table + string(SEPARATOR) + key)
	tv := &mvcc.Value{
		Revision: m.stage.Revision(),
		Raw:      []byte(value),
		Deleted:  false,
	}
	m.stage.Put(tk, tv, 0)
	return nil
}

func (m *MVCCDB) Del(table string, key string) error {
	if !m.isValidTable(table) {
		return ErrTableNotValid
	}
	tk := []byte(table + string(SEPARATOR) + key)
	tv := &mvcc.Value{
		Revision: m.stage.Revision(),
		Raw:      nil,
		Deleted:  true,
	}
	m.stage.Put(tk, tv, 0)
	return nil
}

func (m *MVCCDB) Has(table string, key string) (bool, error) {
	return false, nil
}

func (m *MVCCDB) Keys(table string, prefix string) ([]string, error) {
	return nil, nil
}

func (m *MVCCDB) Commit() string {
	m.revisions[m.stage.Revision()] = m.stage
	m.head = m.stage
	// TODO Why does creating uuid fail?
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Create uuid faied: %v", err)
	}
	r := u.String()
	m.stage = m.head.Fork(r)
	return r
}

func (m *MVCCDB) Rollback() {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Create uuid faied: %v", err)
	}
	r := u.String()
	m.stage = m.head.Fork(r)
}

func (m *MVCCDB) Checkout(revision string) {
	m.head = m.revisions[revision]
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Create uuid faied: %v", err)
	}
	r := u.String()
	m.stage = m.head.Fork(r)
}

func (m *MVCCDB) Fork() *MVCCDB {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("Create uuid faied: %v", err)
	}
	r := u.String()
	mvccdb := &MVCCDB{
		head:      m.head,
		stage:     m.head.Fork(r),
		storage:   m.storage,
		revisions: m.revisions,
	}

	return mvccdb
}

func (m *MVCCDB) Flush(revision string) error {
	return nil
}
