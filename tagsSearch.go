package main
import (
	"fmt"
	"strings"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type TagsQuery struct {
	Text string `json:"searchTerm"`
	GuildIds []int `json:"guildIds"`
}

func HandleTagQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	) {

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	query := TagsQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	allTags := getAllTags(caches)
	matchingTags := []string{}

	text := strings.ToLower(query.Text)
	fmt.Printf("TEXT search=%v\n", text)

	for _, tag := range allTags {
		if strings.Contains(strings.ToLower(tag), text) {
			//fmt.Printf("Tag matched %v %v", tag, query.Text)
			matchingTags = append(
				matchingTags,
				tag)
		}
	}
	bytes, err := json.Marshal(matchingTags)
	w.Write(bytes)
}
	


