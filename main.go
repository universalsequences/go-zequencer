package main
 
import (
	"log"
	"net/http"
	"os"
	"container/list"
	"fmt"
)

func main() {
	caches := LoadAllCaches(os.Args[1])
	cachedQueries := CachedQueries{Queries: make(map[string]CachedQuery), Queue: list.New()}
	
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		HandleQuery(w, r, &caches, &cachedQueries)
	})

	http.HandleFunc("/sample", func(w http.ResponseWriter, r *http.Request) {
		HandleSampleQuery(w, r, &caches)
	})

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Reload called so clearing cache")
		cachedQueries.Clear()
		caches = LoadAllCaches(os.Args[1])
	})
	http.HandleFunc("/textSearch", func(w http.ResponseWriter, r *http.Request) {
		HandleTextQuery(w, r, &caches)
	})
	log.Fatal(http.ListenAndServe(":4567", nil))
}

