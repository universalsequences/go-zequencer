package main
import (
	"fmt"
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
	empty := true

	// First the where clauses must be ordered by the index of this
	// event type
	sortWhereClausesByIndex(query)

	names := [] string {}
	for _, whereClause := range query.WhereClauses {
		names = append(names, whereClause.Name)
	}
	if (query.Debug) {
		fmt.Printf("event=%s where clauses %v\n", query.EventLog, names)
	}

	lastResults := map[string][] map[string]interface{}{}
	
	for _, whereClause := range query.WhereClauses {
		// loop through each where clause (which are sorted by how important their key
		// is)
		if (query.Debug) {
			fmt.Printf("Looping through where clause =%+v\n", whereClause)
		}
		valueList:= whereClause.ValueList
		valueList = append(valueList, whereClause.Value)
		currentResults := map[string][] map[string]interface{}{}
		if _, ok := cache[query.EventLog][whereClause.Name]; ok {
			// this where clause is indexed
			if (empty) {
				lastResults["first"] = cache[query.EventLog][whereClause.Name]
				empty = false
			}
			for _, valueResults := range lastResults {
				if (query.Debug) {
					// fmt.Printf("Search by key=%v for where.name=%v valueList=%v lenResults=%v\n", key, whereClause.Name, valueList, len(valueResults))
				}
				currentResults = searchByKey(valueResults, whereClause.Name, valueList, query.Debug)
			}
		} else {
			// this where clause is NOT indexed
			if (empty) {
				lastResults["first"] = cache[query.EventLog]["blockNumber"]
				empty = false
			}
			for name, valueResults := range lastResults {
				currentResults[name] = naiveSearch(valueResults, query)
			}
		}

		if (query.Debug) {
			fmt.Printf("Partitioned results for where %v\n", whereClause.Name)
		}
		lastResults = currentResults
	}

	results := [] map[string]interface{}{}
	// now union all the last results
	for _, r := range lastResults {
		for _, result := range r {
			results = append(
				results,
				result);
		}
	}
	if (query.Debug) {
		fmt.Printf("Unioning the results %v\n", lastResults)
		fmt.Printf("Union of the results %v\n", results)
	}

	if (empty) {
		values := cache[query.EventLog]["blockNumber"]
		for _, value := range values {
				results = append(results, value)
		}
	}
	// sort by block number
	sort.Slice(results, func (i, j int) bool {
		return results[i]["blockNumber"].(float64) > results[j]["blockNumber"].(float64)
	});
	
	return results 
}


func sortWhereClausesByIndex(query Query) {
	sort.Slice(query.WhereClauses, func(i, j int) bool {
		nameA := query.WhereClauses[i].Name
		nameB := query.WhereClauses[j].Name
		indexOrder := TableIndices[query.EventLog]
		for _, indexEvent := range indexOrder {
			if (indexEvent == nameA) {
				return true;
			}
			if (indexEvent == nameB) {
				return false;
			}
		}
		// if its not indexed, it should go at the end
		return true
	});
}

// return results partitioned by the values in the value list
// Ex: for where guildId in [1,2,3] return results for partitioned by 1, 2, 3
func searchByKey(rows []map[string]interface{}, name string, valueList []interface{}, debug bool) map[string][]map[string]interface{} {
	results := map[string][]map[string]interface{}{}
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
		valueResults := []map[string]interface{}{}
		if (debug && len(rows) > 0) {
			//fmt.Printf("search found x=%v\n", x)
		}
		for i := x; i < len(rows); i++ {
			if (debug) {
				//fmt.Printf("i=%v row[%v][%v] = %+v\n", i, i, name, rows[i])
			}
			if rows[i][name] == value {
				valueResults = append(valueResults, rows[i])
			} else {
				break
			}
		}
		if _, ok := value.(string); ok {
			results[value.(string)] = valueResults
		} else if _, ok := value.(float64); ok {
			results[fmt.Sprintf("%.6f", value.(float64))] = valueResults
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
	
