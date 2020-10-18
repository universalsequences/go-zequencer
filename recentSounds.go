package main

import (
	"strings"
	"sort"
)

func getRecentSounds(caches *Caches, searchTerm string) []SampleResult {
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
	whereClauses := []WhereClause{
			WhereClause{
				Name: "guildId",
				Value: 0.0,
			}}
	if (len(soundIds) > 0) {
		whereClauses = append(
			whereClauses,
			WhereClause{
				Name: "ipfsHash",
				ValueList: soundIds,
			})
	}
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleCreated,
		SelectStatements: []string{
			"ipfsHash",
			"user",
			"title",
		},
		WhereClauses: whereClauses,
		FromBlockNumber: 1,
	};

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
		return convertResults(filtered);
	}
	return convertResults(results)
}

func convertResults(results []map[string]interface{}) []SampleResult {
	convertedResults := []SampleResult{}

	for _, result := range results {
		converted := SampleResult{}
		converted.Title = result["title"].(string)
		converted.BlockNumber = result["blockNumber"].(float64)
		converted.IpfsHash = result["ipfsHash"].(string)
		convertedResults = append(convertedResults, converted)
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
