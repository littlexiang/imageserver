package imageserver

import (
	"sync"
)

type SafeMap struct {
	lock *sync.Mutex
	bm   map[string]*ReqQueue
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock: new(sync.Mutex),
		bm:   make(map[string]*ReqQueue),
	}
}

//Set value if key is not set yet
func (m *SafeMap) SetOnce(k string) (*ReqQueue, bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.bm[k]; ok {
		m.bm[k].length++
		return m.bm[k], false
	}
	m.bm[k] = NewReqQueue()
	return m.bm[k], true
}

//Delete key and do something
func (m *SafeMap) DoAndDelete(k string, function func()) {
	m.lock.Lock()
	defer m.lock.Unlock()
	function()
	delete(m.bm, k)
}

/*
//Get from maps return the k's value
func (m *SafeMap) Get(k string) interface{} {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if val, ok := m.bm[k]; ok {
		return val
	}
	return nil
}


// Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (m *SafeMap) Set(k string, v interface{}) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if val, ok := m.bm[k]; !ok {
		m.bm[k] = v
	} else if val != v {
		m.bm[k] = v
	} else {
		return false
	}
	return true
}

// Returns true if k is exist in the map.
func (m *SafeMap) Check(k string) bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if _, ok := m.bm[k]; !ok {
		return false
	}
	return true
}

func (m *SafeMap) Delete(k string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.bm, k)
}
*/
