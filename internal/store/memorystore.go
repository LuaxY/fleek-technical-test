package store

import "sync"

type Store interface {
	Add(key string, data interface{})
	Get(key string, data *interface{}) (exist bool)
	All(f func(key string, data interface{}))
	Delete(key string)
}

type MemoryStore struct {
	// datas map is used to store file info like encryption key and metadata of file hash
	datas map[string]interface{}

	mutex sync.RWMutex
}

func NewMemoryStore() (*MemoryStore, error) {
	ms := MemoryStore{
		datas: make(map[string]interface{}),
	}

	return &ms, nil
}

func (ms *MemoryStore) Add(key string, data interface{}) {
	ms.mutex.Lock()
	ms.datas[key] = data
	ms.mutex.Unlock()
}

func (ms *MemoryStore) Get(key string, data *interface{}) (exist bool) {
	ms.mutex.RLock()
	*data, exist = ms.datas[key]
	ms.mutex.RUnlock()
	return
}

func (ms *MemoryStore) All(f func(key string, data interface{})) {
	ms.mutex.RLock()
	// return a copy
	for k, v := range ms.datas {
		f(k, v)
	}
	ms.mutex.RUnlock()
}

func (ms *MemoryStore) Delete(key string) {
	ms.mutex.Lock()
	delete(ms.datas, key)
	ms.mutex.Unlock()
}
