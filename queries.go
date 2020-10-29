package main

import (
	"sync"
	"container/list"
)

const MAX_CACHE_SIZE = 2000
const MIN_QUERY_SIZE = 20

type CachedQuery struct {
	Data []byte 
	Timestamp int
}

type CachedQueries struct {
	Queries map[string]CachedQuery
	Queue *list.List
	mutex sync.RWMutex
}

func (c *CachedQueries) Clear() {
	c.Queue = c.Queue.Init()
	c.Queries = make(map[string]CachedQuery)
}

func (c *CachedQueries) newQuery(key string, data []byte ) {
	if (len(data) < MIN_QUERY_SIZE) {
		return;
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if (c.Queue.Len() > MAX_CACHE_SIZE) {
		e := c.Queue.Front()
		delete(c.Queries, e.Value.(string))
		c.Queue.Remove(e)
	}

	query := CachedQuery{Data: data}
	c.Queue.PushBack(key)
	c.Queries[key] = query
}

func (c *CachedQueries) getQuery(key string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if val, ok := c.Queries[key]; ok {
		return val.Data, true
	}

	return nil, false
}

