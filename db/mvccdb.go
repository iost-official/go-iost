package db

import (
	"fmt"
	"sort"
	"strings"
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
	KeysByRange(table string, from string, to string, limit int) ([]string, error)
	Checkout(t string) bool
	Commit(t string)
	CurrentTag() string
	Fork() MVCCDB
	Flush(t string) error
	Size() (int64, error)
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
	Tag string
}

// NewCommit returns new commit
func NewCommit(cache mvcc.Cache, tag string) *Commit {
	return &Commit{
		Cache: cache,
		Tag:   tag,
	}
}

// ForkCache will fork a cache from the commit
func (c *Commit) ForkCache() mvcc.Cache {
	return c.Cache.Fork().(mvcc.Cache)
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

// Add will add a commit with tag
func (m *CommitManager) Add(c *Commit) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.commits = append(m.commits, c)
	m.tags[c.Tag] = c
}

// Get will get a commit by tag
func (m *CommitManager) Get(t string) *Commit {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.tags[t]
}

// Tags will return last 10 tags
func (m *CommitManager) Tags() []string {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	start := len(m.commits) - 10
	if start < 0 {
		start = 0
	}

	res := make([]string, 0)
	for i := start; i < len(m.commits); i++ {
		res = append(res, m.commits[i].Tag)
	}
	return res
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
			delete(m.tags, v.Tag)
			v.Free()
		}
	}
}

// CacheMVCCDB is the mvcc db with cache
type CacheMVCCDB struct {
	head    *Commit
	stage   mvcc.Cache
	storage *kv.Storage
	cm      *CommitManager
	rwmu    sync.RWMutex
}

// NewCacheMVCCDB returns new CacheMVCCDB
func NewCacheMVCCDB(path string, cacheType mvcc.CacheType) (*CacheMVCCDB, error) {
	storage, err := kv.NewStorage(path, kv.LevelDBStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to new storage: %v", err)
	}
	stage := mvcc.NewCache(cacheType)
	cm := NewCommitManager()

	mvccdb := &CacheMVCCDB{
		head:    nil,
		stage:   stage,
		storage: storage,
		cm:      cm,
	}

	tag, err := storage.Get([]byte(string(SEPARATOR) + "tag"))
	if err != nil {
		return nil, fmt.Errorf("failed to get init tag from storage: %v", err)
	}
	mvccdb.Commit(string(tag))

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
	// fmt.Printf("Get %v %v\n", table, key)
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
		return string(v), nil
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
	// fmt.Printf("Put %v %v %v\n", table, key, value)
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
func (m *CacheMVCCDB) KeysByRange(table string, from string, to string, limit int) ([]string, error) {
	if !m.isValidTable(table) {
		return nil, ErrTableNotValid
	}

	var deletedKeys []string
	var cachedKeys []string
	for _, v := range m.stage.All([]byte("")) {
		item, ok := v.(*Item)
		if !ok {
			continue
		}
		if !(from <= item.key && item.key < to) {
			continue
		}
		if item.deleted {
			deletedKeys = append(deletedKeys, item.key)
		} else {
			cachedKeys = append(cachedKeys, item.key)
		}
	}

	fromBytes := []byte(table + string(SEPARATOR) + from)
	toBytes := []byte(table + string(SEPARATOR) + to)
	keys, err := m.storage.KeysByRange(fromBytes, toBytes, limit)
	if err != nil {
		return nil, err
	}
	results := make([]string, 0, len(keys)+len(cachedKeys))
	for _, item := range keys {
		keyWithoutPrefix := strings.TrimPrefix(string(item), table+string(SEPARATOR))
		deleted := false
		for _, s := range deletedKeys {
			if s == keyWithoutPrefix {
				deleted = true
				break
			}
		}
		if !deleted {
			results = append(results, keyWithoutPrefix)
		}
	}
	results = append(results, cachedKeys...)
	sort.Strings(results)
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
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
	m.stage = m.head.ForkCache()
	return true
}

// Commit will commit the stage and add tag to current state of mvccdb
func (m *CacheMVCCDB) Commit(t string) {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.head = NewCommit(m.stage, t)
	m.stage = m.head.ForkCache()
	m.cm.Add(m.head)
}

// CurrentTag will return current tag of mvccdb
func (m *CacheMVCCDB) CurrentTag() string {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	return m.head.Tag
}

// Fork will fork the mvcdb
// thread safe between all forks of the mvccdb
func (m *CacheMVCCDB) Fork() MVCCDB {
	m.rwmu.RLock()
	defer m.rwmu.RUnlock()

	mvccdb := &CacheMVCCDB{
		head:    m.head,
		stage:   m.head.ForkCache(),
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

// Size returns the size of mvccdb
func (m *CacheMVCCDB) Size() (int64, error) {
	return m.storage.Size()
}

// Close will close the mvccdb
func (m *CacheMVCCDB) Close() error {
	return m.storage.Close()
}
