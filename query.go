package main
import (
	"fmt"
	"encoding/json"
	"strings"
	"io/ioutil"
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
	
	results := queryForCache(caches[query.Address], query)
	bytes, err := json.Marshal(results)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func queryForCache(cache Cache, query Query) [] map[string]interface{} {
	fmt.Println(query.EventLog)
	rows := cache[query.EventLog]

	filteredRows := [] map[string]interface{}{}

	for _, row := range rows {
		if (resultMatches(row, query)) {
			filteredRows = append(filteredRows, row)
		}
	}
	return filteredRows
}

func resultMatches(row map[string]interface{}, query Query) bool {
	// go through the whereclause
	matches := true
	for _, whereClause := range query.WhereClauses {
		rowValue := row[whereClause.Name]
		desiredValue := whereClause.Value

		// is it a float or string?
		switch desiredValue.(type) {
		case float64:
			matches = matches && rowValue == desiredValue
		case string:
			matches = matches && strings.Contains(
				strings.ToLower(fmt.Sprintf("%v", rowValue)),
				strings.ToLower(fmt.Sprintf("%v", desiredValue)))
		}
	}
	return matches
}
	
