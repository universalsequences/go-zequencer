package main

import (
	"sort"
	"strings"
)

func SortByIndices(cache map[string]EventCache) Cache {
	// keys are event types which map to the cache for
	// that event (i.e a list of events)
	indexedCache := make(Cache)
	for eventType, eventCache := range cache {
		// check if theres an index
		indices := make(Indices)
		if keys, ok := TableIndices[eventType]; ok {
			// then sort by the fields in the index
			for i, key := range keys {
				if i == 0 { 
					indices[key] = SortByKey(eventCache, key, keys[1])
				} else {
					indices[key] = SortByKey(eventCache, key, keys[0])
				}
			}
		} 
		indices["blockNumber"] = eventCache
		indexedCache[eventType] = indices
	}
	return indexedCache
}

// sort by 2 keys with secondary one breaking any ties
func SortByKey(cache EventCache, key string, secondaryKey string) EventCache {
	sorted := []map[string]interface{}{}
	for _, element := range cache {
		if element[key] == nil {
			element[key] = 0.0
		} 
		if element[secondaryKey] == nil {
			element[secondaryKey] = 0.0
		} 
		sorted = append(sorted, element)
	}
	sort.Slice(sorted, func(i, j int) bool {
		primaryLess := 0.0
		secondaryLess := 0.0
		if _, ok := sorted[i][key].(string); ok {
			if _, ok := sorted[j][key].(string); ok {
				primaryLess = float64(strings.Compare(sorted[i][key].(string), sorted[j][key].(string)))
			}
		} else if _, ok := sorted[i][key].(float64); ok {
			if _, ok := sorted[j][key].(float64); ok {
				primaryLess = sorted[i][key].(float64) - sorted[j][key].(float64)
			}
		} 

		if _, ok := sorted[i][secondaryKey].(string); ok {
			if _, ok := sorted[j][secondaryKey].(string); ok {
				secondaryLess = float64(strings.Compare(sorted[i][secondaryKey].(string), sorted[j][secondaryKey].(string)))
			}
		} else if _, ok := sorted[i][secondaryKey].(float64); ok {
			if _, ok := sorted[j][secondaryKey].(float64); ok {
				secondaryLess = sorted[i][secondaryKey].(float64) - sorted[j][secondaryKey].(float64)
			}
		} 
		if (primaryLess < 0) {
			return true;
		} else if (primaryLess == 0) {
			return secondaryLess < 0;
		} else {
			return false;
		}
	})

	return sorted
}




