package db

import (
	"errors"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/db/mvcc"
	"github.com/iost-official/Go-IOS-Protocol/db/storage"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

//go:generate mockgen -destination mocks/mock_mvccdb.go -package db_mock github.com/iost-official/Go-IOS-Protocol/db MVCCDB

const (
	SEPARATOR = '/'
)

var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrTableNotValid = errors.New("table name is not valid")
)

type Storage interface {
	Get(key []byte) ([]byte, error)
	Put(key []byte, value []byte) error
	Del(key []byte) error
	Keys(prefix []byte) ([][]byte, error)
	BeginBatch() error
	CommitBatch() error
	Close() error
}

type MVCCDB interface {
	Get(table string, key string) (string, error)
	Put(table string, key string, value string) error
	Del(table string, key string) error
	Has(table string, key string) (bool, error)
	Keys(table string, prefix string) ([]string, error)
	Commit()
	Rollback()
	Checkout(t string) bool
	Tag(t string)
	CurrentTag() string
	Fork() MVCCDB
	Flush(t string) error
	Close() error
}

func NewMVCCDB(path string) (MVCCDB, error) {
	return NewCacheMVCCDB(path, mvcc.TrieCache)
}

type Item struct {
	table   string
	key     string
	value   string
	deleted bool
}

type Commit struct {
	mvcc.Cache
	Tags []string
}

func NewCommit(cacheType mvcc.CacheType) *Commit {
	return &Commit{
		Cache: mvcc.NewCache(cacheType),
		Tags:  make([]string, 0),
	}
}

func (c *Commit) Fork() *Commit {
	return &Commit{
		Cache: c.Cache.Fork().(mvcc.Cache),
		Tags:  make([]string, 0),
	}
}

type CacheMVCCDB struct {
	head      *Commit
	stage     *Commit
	tags      map[string]*Commit
	commits   []*Commit
	stagerw   *sync.RWMutex
	tagsrw    *sync.RWMutex
	commitsrw *sync.RWMutex
	storage   Storage
}

func NewCacheMVCCDB(path string, cacheType mvcc.CacheType) (*CacheMVCCDB, error) {
	storage, err := storage.NewLevelDB(path)
	if err != nil {
		return nil, err
	}
	tag, err := storage.Get([]byte(string(SEPARATOR) + "tag"))
	if err != nil {
		tag = []byte("")
	}
	head := NewCommit(cacheType)
	stage := head.Fork()
	tags := make(map[string]*Commit)
	commits := make([]*Commit, 0)

	head.Tags = []string{string(tag)}
	tags[head.Tags[0]] = head
	commits = append(commits, head)
	mvccdb := &CacheMVCCDB{
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

func (m *CacheMVCCDB) isValidTable(table string) bool {
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

func (m *CacheMVCCDB) Get(table string, key string) (string, error) {
	if !m.isValidTable(table) {
		return "", ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	m.stagerw.RLock()
	v := m.stage.Get(k)
	m.stagerw.RUnlock()
	if v == nil {
		v, err := m.storage.Get(k)
		if err != nil {
			ilog.Debugf("Failed to get from storage: %v", err)
			return "", ErrKeyNotFound
		}
		return string(v[:]), nil
	}
	i, ok := v.(*Item)
	if !ok {
		return "", errors.New("can't assert Item type")
	}
	if i.deleted {
		return "", ErrKeyNotFound
	}
	return i.value, nil
}

func (m *CacheMVCCDB) Put(table string, key string, value string) error {
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

func (m *CacheMVCCDB) Del(table string, key string) error {
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

func (m *CacheMVCCDB) Has(table string, key string) (bool, error) {
	if !m.isValidTable(table) {
		return false, ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	m.stagerw.RLock()
	v := m.stage.Get(k)
	m.stagerw.RUnlock()
	if v == nil {
		v, err := m.storage.Get(k)
		if err != nil {
			ilog.Debugf("Failed to get from storage: %v", err)
			return false, nil
		}
		return v != nil, nil
	}
	i, ok := v.(*Item)
	if !ok {
		return false, errors.New("can't assert Item type")
	}
	if i.deleted {
		return false, nil
	}
	return true, nil
}

func (m *CacheMVCCDB) Keys(table string, prefix string) ([]string, error) {
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

func (m *CacheMVCCDB) Commit() {
	m.commitsrw.Lock()
	m.commits = append(m.commits, m.stage)
	m.commitsrw.Unlock()
	m.head = m.stage
	m.stage = m.head.Fork()
}

func (m *CacheMVCCDB) Rollback() {
	m.stage = m.head.Fork()
}

func (m *CacheMVCCDB) Checkout(t string) bool {
	m.tagsrw.RLock()
	head, ok := m.tags[t]
	m.tagsrw.RUnlock()
	if !ok {
		return false
	}
	m.head = head
	m.stage = m.head.Fork()
	return true
}

func (m *CacheMVCCDB) Tag(t string) {
	m.tagsrw.Lock()
	m.tags[t] = m.head
	m.tagsrw.Unlock()
	m.head.Tags = append(m.head.Tags, t)
}

func (m *CacheMVCCDB) CurrentTag() string {
	return m.head.Tags[0]
}

func (m *CacheMVCCDB) Fork() MVCCDB {
	mvccdb := &CacheMVCCDB{
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

func (m *CacheMVCCDB) Flush(t string) error {
	trie := m.tags[t]

	if err := m.storage.BeginBatch(); err != nil {
		return err
	}
	err := m.storage.Put([]byte(string(SEPARATOR)+"tag"), []byte(t))
	if err != nil {
		return err
	}
	for _, v := range trie.All([]byte("")) {
		item, ok := v.(*Item)
		if !ok {
			return errors.New("can't assert Item type")
		}
		if item.deleted {
			err := m.storage.Del([]byte(item.table + string(SEPARATOR) + item.key))
			if err != nil {
				return err
			}
		} else {
			err := m.storage.Put([]byte(item.table+string(SEPARATOR)+item.key), []byte(item.value))
			if err != nil {
				return err
			}
		}
	}
	if err := m.storage.CommitBatch(); err != nil {
		return err
	}
	ilog.Debugf("Commits length: %v", len(m.commits))
	for k, v := range m.commits {
		if v == trie {
			m.commits = m.commits[k:]
			break
		} else {
			for _, t := range v.Tags {
				delete(m.tags, t)
			}
			v.Free()
		}
	}
	return nil
}

func (m *CacheMVCCDB) Close() error {
	return m.storage.Close()
}
