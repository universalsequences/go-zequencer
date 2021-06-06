package main

import (
	"strings"
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
	searchCachedQueries *CachedSearchQueries,
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
		(*caches)[strings.ToLower(contract)] = LoadCache(directoryPath + contract + ".json")

		// certain contracts getting updated should result in clearing the cache
		// since the results may have changed
		if (contract == TOKENIZED_SEQUENCES ||
			contract == XANADU ||
			contract == GUILD_SEQUENCES) {
			projectsCachedQueries.Clear()
		} else if (contract == GUILD_SAMPLES || contract == ARTISTS_CONTRACT) {
			streamCachedQueries.Clear()
			searchCachedQueries.Clear()
			prePopulateCache(caches, ratingsCache, searchCachedQueries)
		}
	}
}
	
