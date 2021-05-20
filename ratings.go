package main

import (
	"io/ioutil"
	"strings"
	"net/http"
	"log"
	"encoding/json"
	"strconv"
)

type RatingCache map[string]int

type RatingQuery struct {
	SampleIds []string `json:"sampleIds"`
}

type RatingResults struct {
	Ratings map[string]int `json:"ratings"`
}

const xanaduContract = "0x305306F68D9C230B59d5B6869AEd1723365C9290";
const NewAnnotationn = "NewAnnotation(bytes32,bytes32,bytes32,address)"
const annotationType = "SAMPLE_RATED"

func HandleRatingsQuery(
	w http.ResponseWriter,
	r *http.Request,
	cache *RatingCache) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)
	query := RatingQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	results := RatingResults{}
	results.Ratings = make(map[string]int)
	for _, id := range query.SampleIds {
		if _, ok := (*cache)[id]; ok {
			results.Ratings[id] = (*cache)[id]
			} else {
				results.Ratings[id] = 0
			}
	}

	bytes, err := json.Marshal(results)
	w.Write(bytes)
}
	
func LoadRatings(caches Caches) RatingCache {
	sum := make(map[string]int)
	count := make(map[string]int)
	cache := make(RatingCache)

	rows := caches[xanaduContract][NewAnnotation]["blockNumber"]
	for _, row := range rows {
		if _, ok := row["annotationType"].(string); ok {
			if (row["annotationType"].(string) == annotationType) {
				rating, _ := strconv.Atoi(row["annotationData"].(string))
				if _, ok :=  sum[row["data"].(string)]; ok {
					sum[row["data"].(string)] += rating
					count[row["data"].(string)] += 1
				} else {
					sum[row["data"].(string)] = rating 
					count[row["data"].(string)] += 1
				}
			}
		}
	}
	for key, _ := range sum {
		cache[key] = sum[key] / count[key]
	}
	
	return cache
}

func getRatings(cache *RatingCache, ids []string) map[string]int {
	ratings := map[string]int{}
	for _, id := range ids {
		if _, ok := (*cache)[id]; ok {
			ratings[id] = (*cache)[id]
		} else {
			ratings[id] = 0
		}
	}
	return ratings
}

func getRatingForSound(caches *Caches, user string, soundId interface{}) int {
	query := NewQuery(XANADU)
	query.From(NewAnnotation)
	query.Select("data")
	query.Select("annotationData")
	query.WhereIs("annotationType", "SAMPLE_RATED")
	if (user != "") {
		query.WhereIs("address", strings.ToLower(user))
	}
	query.WhereIs("data", soundId)
	results := query.ExecuteQuery(caches)
	if (len(results) > 0) {
		row := results[0]
		if rating, err := strconv.Atoi(row["annotationData"].(string)); err == nil {
			return rating
		}
	}
	return 0
}

func getSoundsWithRating(caches *Caches, rating int, user string, soundIds []interface{}) map[string]bool {
	query := NewQuery(XANADU)
	query.From(NewAnnotation)
	query.Select("data")
	query.Select("annotationData")
	query.WhereIs("annotationType", "SAMPLE_RATED")
	query.WhereIs("annotationData", "5")
	if (user != "") {
		query.WhereIs("address", strings.ToLower(user))
	}
	if (len(soundIds) > 0) {
		query.WhereIn("data", soundIds)
	}

	//query.Debug = true
	cache := (*caches)[XANADU] 
	results := queryForCache(cache, query)

	ids := map[string]bool{}
	for _, result := range results {
		ids[result["data"].(string)] = true
	}

	return ids
}

