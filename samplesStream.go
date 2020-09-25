package main

import (
	"encoding/json"
	"time"
	"io/ioutil"
	"fmt"
	"log"
	"net/http"
	"sort"
	"math/rand"
	"math"
)
	
type SamplesStreamQuery struct {
	AndTags []string `json:"andTags"`
	OrTags []string `json:"orTags"`
	GuildIds []int `json:"guildIds"`
}

type SamplesStreamResults struct {
	Ids []string `json:"ids"`
}

type RatedSound struct {
	Id string
	Rating int
}

func HandleSamplesStreamQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	ratingsCache *RatingCache,
	streamCache *CachedQueries) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)

	query := SamplesStreamQuery{}
	json.Unmarshal([]byte(bodyString), &query)
	fmt.Println(query)
	
	bytes, err := json.Marshal(runSamplesStreamQuery(caches, ratingsCache, query, bodyString, streamCache))
	w.Write(bytes)
}

func runSamplesStreamQuery(
	caches *Caches,
	ratingsCache *RatingCache,
	query SamplesStreamQuery,
	bodyString string,
	streamCache *CachedQueries) SamplesStreamResults {

	// first check if we have the sorted results in the cache
	// then we just need to shuffle them
	if val, ok := streamCache.getQuery(bodyString); ok {
		sortedSounds := []string{}
		json.Unmarshal(val, &sortedSounds)
		return SamplesStreamResults{
			Ids: shuffleInParts(sortedSounds),
		}
	}

	sounds := []string{}
	if (len(query.OrTags) == 0) {
		sounds = getSamplesWithAndTags(caches, query.AndTags)
	} else {
		sounds = getSamplesWithOrTags(caches, query.OrTags)
	}

	projectCount := getProjectCountForSamples(caches, sounds)

	// now sort by rating
	ratedSounds := []RatedSound{}
	for _, id := range sounds {
		// for every sequence a sound is used in, increase the rating by 1
		count := 0
		if val, ok := projectCount[id]; ok {
			count = val
		}

		ratedSounds = append(
			ratedSounds,
			RatedSound{
				Id: id,
				Rating: int(math.Pow(float64((*ratingsCache)[id]), 3)) + (count),
			})
	}
	sort.Sort(ByRating(ratedSounds))

	sortedSounds := []string{}
	for _, ratedSound := range ratedSounds {
		sortedSounds = append(
			sortedSounds,
			ratedSound.Id,
		)
	}
	
	bytes, _  := json.Marshal(sortedSounds)
	streamCache.newQuery(bodyString, bytes)

	return SamplesStreamResults{
		Ids: shuffleInParts(sortedSounds),
	}
}

func getSamplesWithAndTags(caches *Caches, tags []string) []string {
	sampleMap := map[string]bool{}

	for i, tag := range tags {
		tagSamples := getSamplesWithOrTags(caches, []string{tag})

		if i == 0 {
			// on first go around we add to map
			for _, sample := range tagSamples {
				sampleMap[sample] = true
			}
		} else {
			tagSamplesMap := map[string]bool{}
			for _, sample := range tagSamples {
				tagSamplesMap[sample] = true
			}

			// on subsequent go arounds we make sure every object in map is contained in
			// tagSamplesMap
			for id, _ := range sampleMap {
				if _, ok := tagSamplesMap[id]; !ok {
					// not inside the map so flip to false
					sampleMap[id] = false
				} 
			}
		}
	}

	samples := []string{}
	for id, val := range sampleMap {
		if val {
			samples = append(
				samples,
				id)
		}
	}
	return samples
}

func getSamplesWithOrTags(caches *Caches, tags []string) []string {
	tagsToFilter := []interface{}{}
	for _, tag := range tags {
		tagsToFilter = append(
			tagsToFilter,
			tag)
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleTagged)
	query.Select("ipfsHash")
	query.WhereIn("tag", tagsToFilter)
	query.WhereIs("guildId", 0.0)

	results := query.ExecuteQuery(caches)
	soundMap := map[string]bool{}
	for _, result := range results {
		soundMap[result["ipfsHash"].(string)] = true
	}

	sounds := []string{}
	for id, _ := range soundMap {
		sounds = append(
			sounds,
			id)
	}

	return sounds
}

type ByRating []RatedSound
func (a ByRating) Len() int           { return len(a) }
func (a ByRating) Less(i, j int) bool { return a[i].Rating > a[j].Rating }
func (a ByRating) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getProjectCountForSamples(caches *Caches, ids []string) map[string]int{
	query := NewQuery(TOKENIZED_SEQUENCES)
	query.From(SampleInSequence)
	query.Select("sampleHash")
	query.Select("sequenceHash")

	sampleHashes := []interface{}{}
	for _, id := range ids {
		sampleHashes = append(
			sampleHashes,
			id)
	}
	query.WhereIn("sampleHash", sampleHashes)

	results := query.ExecuteQuery(caches)
	projectToCount := map[string]int{}

	for _, result := range results {
		sampleHash := result["sampleHash"].(string)
		if _, ok := projectToCount[sampleHash]; !ok {
			projectToCount[sampleHash] = 0
		}
		projectToCount[sampleHash]++
	}
	return projectToCount
}

func shuffleInParts(ids []string) [] string {
	partSize := 15

	rand.Seed(time.Now().UnixNano())

	parts := [][]string{}
	current := []string{}
	for i, id := range ids {
		if (i % partSize  == partSize - 1) {
			rand.Shuffle(len(current), func(i, j int) { current[i], current[j] = current[j], current[i] })
			parts = append(
				parts,
				current)
			current = []string{}
		}
		current = append(
			current,
			id)
	}

	// add the current to the list
	parts = append(
		parts,
		current)

	shuffled := []string{}
	for _, a := range parts {
		for _, b := range a {
			shuffled = append(
				shuffled,
				b)
		}
	}
	return shuffled
}
