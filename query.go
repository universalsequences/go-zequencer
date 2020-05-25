package main
import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"html"
	"log"
	"net/http"
)

type Query struct {
	Address string `json:"address"`
	EventLog string `json:"eventLog"`
	SelectStatements []string `json:"selectStatements"`
	LimitSize int `json:"limitSize"`
	FromBlockNumber int `json:"fromBlockNumber"`
	ToBlockNumber int `json:"toBlockNumber"`
	WhereClauses []WhereClause `json:"whereClauses"`
}

type WhereClause struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
	ValueList []interface{}`json:"valueList"`
}

func HandleQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches Caches) {

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString);
	query := Query{}
	json.Unmarshal([]byte(bodyString), &query)
	
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(query.Address))
}
	
