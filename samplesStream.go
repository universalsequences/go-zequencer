package main

import (
	"math/rand"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"math"
)
	
type SamplesStreamQuery struct {
	AndTags []string `json:"andTags"`
	OrTags []string `json:"orTags"`
	GuildIds []float64 `json:"guildIds"`
	IsFavorited bool `json:"isFavorited"`
	User string `json:"user"`
}

type SamplesStreamResults struct {
	Ids []string `json:"ids"`
}

type RatedSound struct {
	Id string
	Rating float64 
}

const PART_RATIO = 0.65

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
		sounds = getSamplesWithAndTags(caches, query.AndTags, query.GuildIds)
		if (len(sounds) < 3) {
			moreSounds := getSamplesWithOrTags(caches, query.AndTags, query.GuildIds)
			for _, sound := range moreSounds {
				sounds = append(
					sounds, sound)
			}
		} 
	} else {
		sounds = getSamplesWithOrTags(caches, query.OrTags, query.GuildIds)
	}

	if (query.IsFavorited) {
		soundsToRate := []interface{}{}
		for _, id := range sounds {
			soundsToRate = append(soundsToRate, id)
		}
		favoritedSounds := getSoundsWithRating(caches, 5, query.User, soundsToRate, "SAMPLE_RATED")
		_sounds := []string{}
		for id, _ := range favoritedSounds {
			_sounds = append(
				_sounds,
				id)
		}
		sounds = _sounds
	}

	projectCount := getProjectCountForSamples(caches, sounds)
	resampledSounds := getResampledSamples(caches, sounds)
	blockNumbers := getBlockNumbers(caches, sounds)
	minBlock := getMinBlock(blockNumbers)

	// now sort by rating
	ratedSounds := []RatedSound{}
	for _, id := range sounds {
		// for every sequence a sound is used in, increase the rating by 1
		count := 0
		if val, ok := projectCount[id]; ok {
			count = val
		}

		blockDist := blockNumbers[id] - minBlock
		if (blockDist < 0) {
			blockDist = 100
		}

		rating := 
			math.Pow((float64((*ratingsCache)[id]) + 1.0) , 2) *
				math.Pow(float64(count+1), 0.5)

		// i want new sounds to be valued in proportion to how new they are
		rating = math.Pow(blockDist, 0.5) * math.Pow(rating, 0.40)

		if _, ok := resampledSounds[id]; ok {
			rating = math.Pow(rating, 1 / 4.0)
		}

		ratedSounds = append(
			ratedSounds,
			RatedSound{
				Id: id,
				Rating: rating,
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

func getSamplesWithAndTags(caches *Caches, tags []string, guildIds []float64) []string {
	sampleMap := map[string]bool{}

	for i, tag := range tags {
		tagSamples := getSamplesWithOrTags(caches, []string{tag}, guildIds)
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

func getMinBlock(blockNumbers map[string]float64) float64{
	min:= 10000000000.0
	for _, blockNumber := range blockNumbers {
		if (blockNumber < min) {
			min = blockNumber
		}
	}
	return min 
}

func getBlockNumbers(caches *Caches, ids []string) map[string]float64 {
	idsToFilter := []interface{}{}
	for _, id := range ids {
		idsToFilter = append(idsToFilter, id);
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleCreated)
	query.Select("ipfsHash")
	query.WhereIn("ipfsHash", idsToFilter)

	results := query.ExecuteQuery(caches)
	blockMap := map[string]float64{}
	for _, result := range results {
		blockMap[result["ipfsHash"].(string)] = result["blockNumber"].(float64);
	}
	return blockMap
}

func getSamplesWithOrTags(caches *Caches, tags []string, guildIds [] float64) []string {
	tagsToFilter := []interface{}{}
	for _, tag := range tags {
		tagsToFilter = append(
			tagsToFilter,
			tag)
	}

	guildIdsToFilter := []interface{}{}
	for _, guildId := range guildIds {
		guildIdsToFilter = append(guildIdsToFilter, guildId)
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleTagged)
	query.Select("ipfsHash")
	query.WhereIn("tag", tagsToFilter)
	query.WhereIn("guildId", guildIdsToFilter)

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

func getResampledSamples(caches *Caches, ids []string) map[string]bool {
	idsToFilter := []interface{}{}
	for _, id := range ids {
		idsToFilter = append(idsToFilter, id);
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleTagged)
	query.Select("ipfsHash")
	query.WhereIs("tag", "Resampled")

	results := query.ExecuteQuery(caches)
	blockMap := map[string]bool{}
	for _, result := range results {
		blockMap[result["ipfsHash"].(string)] = true
	}
	return blockMap
}

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
	rand.Seed(time.Now().UnixNano())

	partSize := int(math.Floor(math.Pow(float64(len(ids)) , PART_RATIO)))

	parts := [][]string{}
	current := []string{}
	for i, id := range ids {
		if (i % partSize == partSize - 1) {
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
