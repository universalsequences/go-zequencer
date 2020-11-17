package main
import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

func HandleQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	cachedQueries *CachedQueries) {

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)
	if val, ok := cachedQueries.getQuery(bodyString); ok {
		w.Write(val)
		return
	}

	query := Query{}
	json.Unmarshal([]byte(bodyString), &query)
	
	results := queryForCache((*caches)[query.Address], query)
	bytes, err := json.Marshal(results)
	cachedQueries.newQuery(bodyString, bytes)

	w.Write(bytes)
}

func queryForCache(cache Cache, query Query) [] map[string]interface{} {
	results := [] map[string]interface{}{}
	empty := true
	for _, whereClause := range query.WhereClauses {
		// each of these where clauses get AND'd together

		// for each of these where clauses we do a binary search giving us
		// results matching that where clause

		// for example: guildId = 1 AND tag = "breaks"
		// this would need to do 2 searches and then do an intersectiono

		// ideally though the indices would do sub-sorting so
		// within the guildId = 0 the tags would be sorted by tags

		// so once its gone through the first key, it'll search through the second

		valueList:= whereClause.ValueList
		valueList = append(valueList, whereClause.Value)
		if _, ok := cache[query.EventLog][whereClause.Name]; ok {
			// search by Name in the 
			if (empty) {
				results = cache[query.EventLog][whereClause.Name]
				empty = false
			}
			results = searchByKey(results, whereClause.Name, valueList)
		} else {
			if (empty) {
				results = cache[query.EventLog]["blockNumber"]
				empty = false
			}
			results = naiveSearch(results, query)
		}
	}
	if (empty) {
		for _, values := range cache[query.EventLog] {
			for _, value := range values {
				results = append(results, value)
			}
			break
		}
	}
	// sort by block number
	sort.Slice(results, func (i, j int) bool {
		return results[i]["blockNumber"].(float64) > results[j]["blockNumber"].(float64)
	});
	
	return results 
}

func searchByKey(rows []map[string]interface{}, name string, valueList []interface{}) [] map[string]interface{} {
	results := []map[string]interface{}{}
	for _, value := range valueList {
		x := sort.Search(len(rows), func (i int) bool {
			if _, ok := value.(string); ok {
				return strings.Compare(rows[i][name].(string), value.(string)) >= 0;
			} else if _, ok := value.(float64); ok {
				return rows[i][name].(float64) >= value.(float64)
			} else {
				return false;
			}
		});
		for i := x; i < len(rows); i++ {
			if rows[i][name] == value {
				results = append(results, rows[i])
			} else {
				break
			}
		}
	}
	return results
}

func naiveSearch(rows []map[string]interface{}, query Query) [] map[string]interface{} {
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
	
