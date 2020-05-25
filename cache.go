package main

import (
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"log"
)

type EventCache []map[string]interface{}
type Cache map[string]EventCache
type Caches map[string]Cache

func LoadAllCaches(directoryPath string) Caches {
	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	caches := make(Caches)
	for _, file := range files {
		caches[file.Name()] = LoadCache(directoryPath + file.Name())
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

	var cache Cache 
	json.Unmarshal(bytes, &cache)
	return cache;
}


func PrintCache(cache Cache) {
	for eventType, eventCache := range cache {
		fmt.Println(eventType)

		for _, row := range eventCache {
			fmt.Println(row)
		}
	}
}



