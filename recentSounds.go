package main

import (
	"strings"
	"sort"
)

func getRecentSounds(
	caches *Caches,
	searchTerm string,
	guildIds []float64,
	year float64,
	filterFavorites bool) []SampleResult {
	filterByTitle := false
	soundIds := []interface{}{}

	if (searchTerm != "") {
		// then we need to search for tags
		matchingTags := findMatchingTags(caches, searchTerm);
		if (len(matchingTags) == 0) {
			filterByTitle = true
		} else {
			soundIds = getSoundsWithOrTags(caches, matchingTags)
		}
	}
	if (year != 0) {
		soundIds = getSoundsWithYear(caches, year);
	}

	if (filterFavorites) {
		favoritedSounds := getSoundsWithRating(caches, 5)
		if (len(soundIds) == 0) {
			for id, _ := range favoritedSounds {
				soundIds = append(
					soundIds,
					id)
			}
		} else {
		}
	}

	guildList := []interface{}{}
	for _, guildId := range guildIds {
		guildList = append(guildList, guildId);
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleCreated)
	query.Select("ipfsHash")
	query.Select("user")
	query.Select("title")

	
	if (!filterFavorites) {
		query.WhereIn("guildId", guildList)
	}

	if (len(soundIds) > 0) {
		query.WhereIn("ipfsHash",soundIds)
	}

	cache := (*caches)[GUILD_SAMPLES] 
	results := queryForCache(cache, query)
	if (filterByTitle) {
		filtered := []map[string]interface{}{}
		searchTerm = strings.ToLower(searchTerm)
		for _, result := range results {
			title := result["title"].(string)
			if (strings.Contains(strings.ToLower(title), searchTerm)) {
				filtered = append(
					filtered,
					result)
			}
		}
		x := convertResults(filtered);
		return x
	}
	x := convertResults(results)
	return x
}

func convertResults(results []map[string]interface{}) []SampleResult {
	uniqueResults := map[string]SampleResult{}

	for _, result := range results {
		converted := SampleResult{}
		converted.Title = result["title"].(string)
		converted.BlockNumber = result["blockNumber"].(float64)
		converted.IpfsHash = result["ipfsHash"].(string)
		converted.User = result["user"].(string)
		uniqueResults[converted.IpfsHash] = converted
	}
	convertedResults := []SampleResult{}
	for _, result := range uniqueResults {
		convertedResults = append(convertedResults, result)
	}
	sort.Sort(ByBlock(convertedResults))
	return convertedResults
}

type ByBlock []SampleResult
func (a ByBlock) Len() int           { return len(a) }
func (a ByBlock) Less(i, j int) bool { return a[i].BlockNumber > a[j].BlockNumber }
func (a ByBlock) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }


func findMatchingTags(caches *Caches, searchTerm string) []interface{} {
	searchTerm = strings.ToLower(searchTerm)
	tags := []interface{}{}
	allTags := getAllTags(caches)
	for _, tag := range allTags {
		if strings.Contains(strings.ToLower(tag), searchTerm) {
			tags = append(tags, tag)
		}
	}
	return tags
}

func getSoundsWithYear(caches *Caches, year float64) []interface{}{
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleYear,
		SelectStatements: []string{
			"ipfsHash",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "year",
				Value: year,
			}},
		FromBlockNumber: 1,
	};

	cache := (*caches)[GUILD_SAMPLES] 
	results := queryForCache(cache, query)

	ids := []interface{}{}
	for _, result := range results {
		ids = append(ids, result["ipfsHash"])
	}
	return ids
}
