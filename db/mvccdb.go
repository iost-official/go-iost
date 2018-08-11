package db

import (
	"errors"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db/mvcc/trie"
	"github.com/iost-official/Go-IOS-Protocol/db/storage"
)

const (
	SEPARATOR = '/'
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrTableNotValid = errors.New("table name is not valid")
)

type Item struct {
	table   string
	key     string
	value   string
	deleted bool
}

type Storage interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Del(key []byte) error
	Keys(prefix []byte) ([][]byte, error)
	BeginBatch() error
	CommitBatch() error
	Close() error
}

type MVCCDB struct {
	head      *trie.Trie
	stage     *trie.Trie
	tags      map[string]*trie.Trie
	commits   []*trie.Trie
	stagerw   *sync.RWMutex
	tagsrw    *sync.RWMutex
	commitsrw *sync.RWMutex
	storage   Storage
}

func NewMVCCDB(path string) (*MVCCDB, error) {
	storage, err := storage.NewLevelDB(path)
	if err != nil {
		return nil, err
	}
	tag, err := storage.Get([]byte(string(SEPARATOR) + "tag"))
	if err != nil {
		tag = []byte("")
	}
	head := trie.New()
	stage := head.Fork()
	tags := make(map[string]*trie.Trie)
	commits := make([]*trie.Trie, 1)

	head.SetTag(string(tag))
	tags[string(tag)] = head
	commits = append(commits, head)
	mvccdb := &MVCCDB{
		head:      head,
		stage:     stage,
		tags:      tags,
		commits:   commits,
		stagerw:   new(sync.RWMutex),
		tagsrw:    new(sync.RWMutex),
		commitsrw: new(sync.RWMutex),
		storage:   storage,
	}
	return mvccdb, nil
}

func (m *MVCCDB) isValidTable(table string) bool {
	if table == "" {
		return false
	}
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
	k := []byte(table + string(SEPARATOR) + key)
	m.stagerw.RLock()
	v, ok := m.stage.Get(k).(*Item)
	m.stagerw.RUnlock()
	if !ok {
		return "", errors.New("can't assert Item type")
	}
	if v == nil {
		v, err := m.storage.Get(k)
		return string(v[:]), err
	}
	if v.deleted {
		return "", ErrKeyNotFound
	}
	return v.value, nil
}

func (m *MVCCDB) Put(table string, key string, value string) error {
	if !m.isValidTable(table) {
		return ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	v := &Item{
		table:   table,
		key:     key,
		value:   value,
		deleted: false,
	}
	m.stagerw.RLock()
	m.stage.Put(k, v)
	m.stagerw.RUnlock()
	return nil
}

func (m *MVCCDB) Del(table string, key string) error {
	if !m.isValidTable(table) {
		return ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	v := &Item{
		table:   table,
		key:     key,
		value:   "",
		deleted: true,
	}
	m.stagerw.RLock()
	m.stage.Put(k, v)
	m.stagerw.RUnlock()
	return nil
}

func (m *MVCCDB) Has(table string, key string) (bool, error) {
	if !m.isValidTable(table) {
		return false, ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	m.stagerw.RLock()
	v, ok := m.stage.Get(k).(*Item)
	m.stagerw.RUnlock()
	if !ok {
		return false, errors.New("can't assert Item type")
	}
	if v == nil {
		v, err := m.storage.Get(k)
		return v != nil, err
	}
	if v.deleted {
		return false, nil
	}
	return true, nil
}

func (m *MVCCDB) Keys(table string, prefix string) ([]string, error) {
	//if !m.isValidTable(table) {
	//	return nil, ErrTableNotValid
	//}
	//p := []byte(table + string(SEPARATOR) + prefix)
	//m.stagerw.RLock()
	//vlist := m.stage.Keys(p)
	//m.stagerw.RUnlock()
	//keys, err := m.storage.Keys(p)
	//if err != nil {
	//	return nil, err
	//}
	//// TODO use iterator instead of keys
	//for key := range keys {
	//
	//}
	//	if !ok {
	//		return nil, error.New("can't assert Item type")
	//	}
	return nil, nil
}

func (m *MVCCDB) Commit() {
	m.commitsrw.Lock()
	m.commits = append(m.commits, m.stage)
	m.commitsrw.Unlock()
	m.head = m.stage
	m.stage = m.head.Fork()
}

func (m *MVCCDB) Rollback() {
	m.stage = m.head.Fork()
}

func (m *MVCCDB) Checkout(t string) {
	m.tagsrw.RLock()
	m.head = m.tags[t]
	m.tagsrw.RUnlock()
	m.stage = m.head.Fork()
}

func (m *MVCCDB) Tag(t string) {
	m.tagsrw.Lock()
	m.tags[t] = m.head
	m.tagsrw.Unlock()
	m.head.SetTag(t)
}

func (m *MVCCDB) Fork() *MVCCDB {
	mvccdb := &MVCCDB{
		head:      m.head,
		stage:     m.head.Fork(),
		tags:      m.tags,
		commits:   m.commits,
		stagerw:   m.stagerw,
		tagsrw:    m.tagsrw,
		commitsrw: m.commitsrw,
		storage:   m.storage,
	}
	return mvccdb
}

func (m *MVCCDB) Flush(t string) error {
	trie := m.tags[t]

	if err := m.storage.BeginBatch(); err != nil {
		return err
	}
	for _, v := range trie.Keys([]byte("")) {
		item, ok := v.(*Item)
		if !ok {
			return errors.New("can't assert Item type")
		}
		if item.deleted {
			err := m.storage.Del([]byte(item.key))
			if err != nil {
				return err
			}
		} else {
			err := m.storage.Put([]byte(item.key), []byte(item.value))
			if err != nil {
				return err
			}
		}
	}
	if err := m.storage.CommitBatch(); err != nil {
		return err
	}

	for k, v := range m.commits {
		if v == trie {
			m.commits = m.commits[k:]
			break
		} else {
			delete(m.tags, v.Tag())
			v.Free()
		}
	}
	return nil
}

func (m *MVCCDB) Close() error {
	return m.storage.Close()
}
