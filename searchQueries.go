package main

import (
	"sync"
	"container/list"
)

type CachedSearchQuery struct {
	Data []BlockResults
	Timestamp int
}

type CachedSearchQueries struct {
	Queries map[string]CachedSearchQuery
	Queue *list.List
	mutex sync.RWMutex
}

func (c *CachedSearchQueries) Clear() {
	c.Queue = c.Queue.Init()
	c.Queries = make(map[string]CachedSearchQuery)
}

func (c *CachedSearchQueries) newQuery(key string, data []BlockResults) {
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

	query := CachedSearchQuery{Data: data}
	c.Queue.PushBack(key)
	c.Queries[key] = query
}

func (c *CachedSearchQueries) getQuery(key string) ([]BlockResults, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if val, ok := c.Queries[key]; ok {
		return val.Data, true
	}

	return nil, false
}

