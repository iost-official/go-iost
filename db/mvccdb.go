package db

import (
	"fmt"
	"sync"

	"github.com/iost-official/go-iost/db/kv"
	"github.com/iost-official/go-iost/db/mvcc"
)

//go:generate mockgen -destination mocks/mock_mvccdb.go -package db_mock github.com/iost-official/go-iost/db MVCCDB

// constant of mvccdb
const (
	SEPARATOR = '/'
)

// error of mvccdb
var (
	ErrTableNotValid = fmt.Errorf("table name is not valid")
)

// MVCCDB is the interface of mvccdb
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

// NewMVCCDB return new mvccdb
func NewMVCCDB(path string) (MVCCDB, error) {
	return NewCacheMVCCDB(path, mvcc.MapCache)
}

// Item is the value of cache
type Item struct {
	table   string
	key     string
	value   string
	deleted bool
}

// Commit is the cache of specify tag
type Commit struct {
	mvcc.Cache
	Tags []string
}

// NewCommit returns new commit
func NewCommit(cacheType mvcc.CacheType) *Commit {
	return &Commit{
		Cache: mvcc.NewCache(cacheType),
		Tags:  make([]string, 0),
	}
}

// Fork will fork the commit
// thread safe between all forks of the commit
func (c *Commit) Fork() *Commit {
	return &Commit{
		Cache: c.Cache.Fork().(mvcc.Cache),
		Tags:  make([]string, 0),
	}
}

// CommitManager is the commit manager, support get, delete etc.
type CommitManager struct {
	tags    map[string]*Commit
	commits []*Commit
	rwmu    *sync.RWMutex
}

// NewCommitManager returns new commit manager
func NewCommitManager() *CommitManager {
	tags := make(map[string]*Commit)
	commits := make([]*Commit, 0)
	rwmu := new(sync.RWMutex)
	return &CommitManager{
		tags:    tags,
		commits: commits,
		rwmu:    rwmu,
	}
}

// Add will add a commit
func (m *CommitManager) Add(c *Commit) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.commits = append(m.commits, c)
}

// Get will get a commit by tag
func (m *CommitManager) Get(t string) *Commit {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.tags[t]
}

// AddTag will make the commit with the tag
func (m *CommitManager) AddTag(c *Commit, t string) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	c.Tags = append(c.Tags, t)
	m.tags[t] = c
}

// GetTags returns tags of the commit
func (m *CommitManager) GetTags(c *Commit) []string {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return c.Tags
}

// FreeBefore will free the momery of commits before the commit
func (m *CommitManager) FreeBefore(c *Commit) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	for k, v := range m.commits {
		if v == c {
			m.commits = m.commits[k:]
			break
		} else {
			for _, t := range v.Tags {
				delete(m.tags, t)
			}
			v.Free()
		}
	}
}

// CacheMVCCDB is the mvcc db with cache
type CacheMVCCDB struct {
	head    *Commit
	rwmu    sync.RWMutex
	stage   *Commit
	storage *kv.Storage
	cm      *CommitManager
}

// NewCacheMVCCDB returns new CacheMVCCDB
func NewCacheMVCCDB(path string, cacheType mvcc.CacheType) (*CacheMVCCDB, error) {
	storage, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to new storage: %v", err)
	}
	tag, err := storage.Get([]byte(string(SEPARATOR) + "tag"))
	if err != nil {
		return nil, fmt.Errorf("failed to get from storage: %v", err)
	}
	head := NewCommit(cacheType)
	stage := head.Fork()
	cm := NewCommitManager()

	cm.AddTag(head, string(tag))
	cm.Add(head)
	mvccdb := &CacheMVCCDB{
		head:    head,
		stage:   stage,
		storage: storage,
		cm:      cm,
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

// Get returns the value of specify key and table
func (m *CacheMVCCDB) Get(table string, key string) (string, error) {
	if !m.isValidTable(table) {
		return "", ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	v := m.stage.Get(k)
	if v == nil {
		v, err := m.storage.Get(k)
		if err != nil {
			return "", fmt.Errorf("failed to get from storage: %v", err)
		}
		return string(v[:]), nil
	}
	i, ok := v.(*Item)
	if !ok {
		return "", fmt.Errorf("can't assert Item type")
	}
	if i.deleted {
		return "", nil
	}
	return i.value, nil
}

// Put will insert the key-value pair into the table
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
	m.stage.Put(k, v)
	return nil
}

// Del will remove the specify key in the table
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
	m.stage.Put(k, v)
	return nil
}

// Has returns whether the specified key exists in the table
func (m *CacheMVCCDB) Has(table string, key string) (bool, error) {
	if !m.isValidTable(table) {
		return false, ErrTableNotValid
	}
	k := []byte(table + string(SEPARATOR) + key)
	v := m.stage.Get(k)
	if v == nil {
		return m.storage.Has(k)
	}
	i, ok := v.(*Item)
	if !ok {
		return false, fmt.Errorf("can't assert Item type")
	}
	if i.deleted {
		return false, nil
	}
	return true, nil
}

// Keys returns the list of key prefixed with prefix in the table
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

// Commit will commit current state of mvccdb
func (m *CacheMVCCDB) Commit() {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.cm.Add(m.stage)
	m.head = m.stage
	m.stage = m.head.Fork()
}

// Rollback will rollback the state of mvccdb
func (m *CacheMVCCDB) Rollback() {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.stage = m.head.Fork()
}

// Checkout will checkout the specify tag of mvccdb
func (m *CacheMVCCDB) Checkout(t string) bool {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	head := m.cm.Get(t)
	if head == nil {
		return false
	}
	m.head = head
	m.stage = m.head.Fork()
	return true
}

// Tag will add tag to current state of mvccdb
func (m *CacheMVCCDB) Tag(t string) {
	m.Commit()

	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	m.cm.AddTag(m.head, t)
}

// CurrentTag will returns current tag of mvccdb
func (m *CacheMVCCDB) CurrentTag() string {
	// TODO how to write better in this place
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	tags := m.cm.GetTags(m.head)
	return tags[len(tags)-1]
}

// Fork will fork the mvcdb
// thread safe between all forks of the mvccdb
func (m *CacheMVCCDB) Fork() MVCCDB {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	mvccdb := &CacheMVCCDB{
		head:    m.head,
		stage:   m.head.Fork(),
		storage: m.storage,
		cm:      m.cm,
	}
	return mvccdb
}

// Flush will persist the current state of mvccdb
func (m *CacheMVCCDB) Flush(t string) error {
	commit := m.cm.Get(t)
	if commit == nil {
		return fmt.Errorf("not found tag: %v", t)
	}
	if err := m.storage.BeginBatch(); err != nil {
		return err
	}
	err := m.storage.Put([]byte(string(SEPARATOR)+"tag"), []byte(t))
	if err != nil {
		return err
	}
	for _, v := range commit.All([]byte("")) {
		item, ok := v.(*Item)
		if !ok {
			return fmt.Errorf("can't assert Item type")
		}
		if item.deleted {
			err := m.storage.Delete([]byte(item.table + string(SEPARATOR) + item.key))
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
	m.cm.FreeBefore(commit)
	return nil
}

// Close will close the mvccdb
func (m *CacheMVCCDB) Close() error {
	return m.storage.Close()
}
