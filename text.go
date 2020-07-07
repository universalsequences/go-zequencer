package main
import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type TextQuery struct {
	Text string `json:"textQuery"`
	GuildIds []int `json:"guildIds"`
}

func HandleTextQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches) {

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString);
	query := TextQuery{}
	json.Unmarshal([]byte(bodyString), &query)
}
	

