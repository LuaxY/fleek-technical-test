package store

import "sync"

// Store interface used to store information that can be retrieved during process
type Store interface {
	Add(key string, data interface{})
	Get(key string, data *interface{}) (exist bool)
	All(f func(key string, data interface{}))
	Delete(key string)
}

// MemoryStore simple in memory map of any king of data linked to string key,
// data will remain accessible during session but lost when program stop or restart
type MemoryStore struct {
	// datas map is used to store file info like encryption key and metadata of file hash
	datas map[string]interface{}

	mutex sync.RWMutex
}

// NewMemoryStore return simple in memory store
func NewMemoryStore() (*MemoryStore, error) {
	ms := MemoryStore{
		datas: make(map[string]interface{}),
	}

	return &ms, nil
}

// Add adds new data in memory storing linked to it's string key
func (ms *MemoryStore) Add(key string, data interface{}) {
	ms.mutex.Lock()
	ms.datas[key] = data
	ms.mutex.Unlock()
}

// Get returns data of associated string key
// return boolean to tel user if data is present in store
func (ms *MemoryStore) Get(key string, data *interface{}) (exist bool) {
	ms.mutex.RLock()
	*data, exist = ms.datas[key]
	ms.mutex.RUnlock()
	return
}

// All returns copy of all data in memory store
func (ms *MemoryStore) All(f func(key string, data interface{})) {
	ms.mutex.RLock()
	// return a copy
	for k, v := range ms.datas {
		f(k, v)
	}
	ms.mutex.RUnlock()
}

// Delete remove data of provided key from memory store
func (ms *MemoryStore) Delete(key string) {
	ms.mutex.Lock()
	delete(ms.datas, key)
	ms.mutex.Unlock()
}
