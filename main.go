package main
 
import (
	"log"
	"net/http"
	"os"
)

func main() {
	caches := LoadAllCaches(os.Args[1])
	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		HandleQuery(w, r, caches)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

