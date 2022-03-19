package main
 
import (
	"log"
	"net/http"
	"os"
	"container/list"
)

func main() {
	directoryPath := os.Args[1]
	caches := LoadAllCaches(directoryPath)
	cachedQueries := CachedQueries{Queries: make(map[string]CachedQuery), Queue: list.New()}
	searchCachedQueries := CachedSearchQueries{Queries: make(map[string]CachedSearchQuery), Queue: list.New()}
	projectsCachedQueries := CachedQueries{Queries: make(map[string]CachedQuery), Queue: list.New()}
	streamCachedQueries := CachedQueries{Queries: make(map[string]CachedQuery), Queue: list.New()}
	ratings := LoadRatings(caches)
	prePopulateCache(&caches, &ratings, &searchCachedQueries)
	
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		HandleQuery(w, r, &caches, &cachedQueries)
	})

	http.HandleFunc("/tagSearch", func(w http.ResponseWriter, r *http.Request) {
		HandleTagQuery(w, r, &caches)
	})

	http.HandleFunc("/ratings", func(w http.ResponseWriter, r *http.Request) {
		HandleRatingsQuery(w, r, &ratings)
	})

	http.HandleFunc("/sample", func(w http.ResponseWriter, r *http.Request) {
		HandleSampleQuery(w, r, &caches)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		HandleSearchQuery(w, r, &caches, &ratings, &searchCachedQueries)
	})

	http.HandleFunc("/samplesStream", func(w http.ResponseWriter, r *http.Request) {
		HandleSamplesStreamQuery(w, r, &caches, &ratings, &streamCachedQueries)
	})

	http.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		HandleProjectsQuery(w, r, &caches, &projectsCachedQueries)
	})

	http.HandleFunc("/projectEdits", func(w http.ResponseWriter, r *http.Request) {
		HandleProjectEdits(w, r, &caches)
	})

	http.HandleFunc("/packSearch", func(w http.ResponseWriter, r *http.Request) {
		HandlePackQuery(w, r, &caches)
	})

	http.HandleFunc("/presetSearch", func(w http.ResponseWriter, r *http.Request) {
		HandlePresetQuery(w, r, &caches)
	})

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		HandleReloadRequest(
			w,
			r,
			directoryPath,
			&caches,
			&ratings,
			&searchCachedQueries,
			&projectsCachedQueries,
			&cachedQueries,
			&streamCachedQueries);
	})
	http.HandleFunc("/textSearch", func(w http.ResponseWriter, r *http.Request) {
		HandleTextQuery(w, r, &caches)
	})
	log.Fatal(http.ListenAndServe(":4567", nil))
	runRepl(&caches)
}

