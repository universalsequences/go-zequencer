package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"log"
)

type EventCache []map[string]interface{}
type Indices map[string]EventCache
type Cache map[string]Indices
type Caches map[string]Cache

func LoadAllCaches(directoryPath string) Caches {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	caches := make(Caches)
	for _, file := range files {
		var contractAddress = file.Name()[0:len(file.Name())-len(".json")]
		caches[contractAddress] = LoadCache(directoryPath + file.Name())
	}
	return caches;
}

func LoadCache(fileName string) Cache{
	fmt.Println("Loading %s", fileName)
	jsonFile, err := os.Open(fileName)
	if (err != nil) {
		log.Fatal(err)
	}
	bytes, err := ioutil.ReadAll(jsonFile)
	if (err != nil) {
		log.Fatal(err)
	}

	var cache map[string]EventCache
	json.Unmarshal(bytes, &cache)
	for _, eventCache := range cache {
		ReverseSlice(eventCache)
	}
	return SortByIndices(cache);
}

func ReverseSlice(s EventCache) EventCache {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
    s[i], s[j] = s[j], s[i]
	}
	return s
}


func PrintCache(cache Cache) {
	for eventType, eventCache := range cache {
		fmt.Println(eventType)

		for _, row := range eventCache {
			fmt.Println(row)
		}
	}
}



