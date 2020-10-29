package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
)
	
type ReloadRequest struct {
	Changes []string `json:"changes"`
}

func HandleReloadRequest(
	w http.ResponseWriter,
	r *http.Request,
	directoryPath string,
	caches *Caches,
	ratingsCache *RatingCache,
	searchCachedQueries *CachedQueries,
	projectsCachedQueries *CachedQueries,
	cachedQueries *CachedQueries,
	streamCachedQueries *CachedQueries,
) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	
	bodyString := string(bodyBytes)
	query := ReloadRequest{}
	json.Unmarshal([]byte(bodyString), &query)

	cachedQueries.Clear()

	for _, contract := range query.Changes {
		fmt.Printf("Reloading contract=%v\n", contract)
		(*caches)[contract] = LoadCache(directoryPath + contract + ".json")
		if (contract == TOKENIZED_SEQUENCES) {
			projectsCachedQueries.Clear()
		} else if (contract == GUILD_SAMPLES) {
			streamCachedQueries.Clear()
			prePopulateCache(caches, ratingsCache, searchCachedQueries)
		}
	}
}
	
