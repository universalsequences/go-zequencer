package main
import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Query struct {
	Address string `json:"address"`
	EventLog string `json:"eventLog"`
	SelectStatements []string `json:"selectStatements"`
	LimitSize int `json:"limitSize"`
	FromBlockNumber float64 `json:"fromBlockNumber"`
	ToBlockNumber float64 `json:"toBlockNumber"`
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
	query := Query{}
	json.Unmarshal([]byte(bodyString), &query)
	
	results := queryForCache(caches[query.Address], query)
	bytes, err := json.Marshal(results)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func queryForCache(cache Cache, query Query) [] map[string]interface{} {
	rows := cache[query.EventLog]
	
	filteredRows := [] map[string]interface{}{}
	for _, row := range rows {
		if (query.LimitSize != 0 && len(filteredRows) > query.LimitSize) {
			break
		}
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
		valueList:= whereClause.ValueList
		valueList = append(valueList, whereClause.Value)
		rowMatch := false

		for _, desiredValue := range valueList {
			if (rowValue == nil) {
				if (desiredValue == 0.0 || desiredValue == nil) {
					rowMatch = true
					break;
				}
			}
			rowMatch = rowMatch || rowValue == desiredValue
		}
		matches = matches && rowMatch
	}
		
	blockNumber := row["blockNumber"]
	if flt_blockNumber , ok := blockNumber.(float64); ok {
		if (flt_blockNumber >= query.FromBlockNumber &&
			(query.ToBlockNumber == 0 || flt_blockNumber <= query.ToBlockNumber)) {
			return matches;
		}
	}
	return false;
}
	
