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
	filterFavorites bool,
	releaseId float64,
	videoId string,
	pack string,
	user string) []SampleResult {
	filterByTitle := false
	soundIds := []interface{}{}

	guildList := []interface{}{}
	for _, guildId := range guildIds {
		guildList = append(guildList, guildId);
	}
	if (len(guildList) == 0) {
		guildList = append(guildList, 0.0)
	}

	if (searchTerm != "") {
		// then we need to search for tags
		matchingTags := findMatchingTags(caches, searchTerm);
		if (len(matchingTags) == 0) {
			filterByTitle = true
		} else {
			soundIds = getSoundsWithOrTags(caches, matchingTags, guildList)
		}
	}
	if (releaseId != 0.0) {
		soundIds = getSamplesFromRelease(caches, releaseId, guildList)
	}

	if (videoId != "") {
		soundIds = getSamplesFromVideo(caches, videoId, guildList)
	}

	if (year != 0.0) {
		soundIds = getSoundsWithYear(caches, year, guildList);
	}

	if (pack != "") {
		soundIds = getSoundsWithPack(caches, pack, guildList);
	}

	if (filterFavorites) {
		// NEEDS TO INCLUDE USER OR THIS DOESNT MAKE SENSE
		favoritedSounds := getSoundsWithRating(caches, 5, user, soundIds, "SAMPLE_RATED")
		sounds := []interface{}{}
		for id, _ := range favoritedSounds {
			sounds = append(
				sounds,
				id)
		}
		soundIds = sounds
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleCreated)
	query.Select("guildId")
	query.Select("ipfsHash")
	query.Select("user")
	query.Select("title")

	if (len(soundIds) > 0) {
		query.WhereIn("ipfsHash",soundIds)
	}
	if (searchTerm == "" || year != 0 || videoId != "" || releaseId != 0.0) {
		query.WhereIn("guildId",guildList)
	}

	results := query.ExecuteQuery(caches)
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
		converted.GuildId = result["guildId"].(float64)
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

func getSoundsWithYear(caches *Caches, year float64, guildIds []interface{}) []interface{}{
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
			},
			WhereClause{
				Name: "guildId",
				ValueList: guildIds,
			}},
		FromBlockNumber: 1,
	};

	results := query.ExecuteQuery(caches)

	ids := []interface{}{}
	for _, result := range results {
		ids = append(ids, result["ipfsHash"])
	}
	return ids
}

func filterByList(a []interface{}, bMap map[string]bool) []interface{} {
	filtered := []interface{}{}
	for _, val := range a {
		if _, ok := bMap[val.(string)]; ok {
			filtered = append(filtered, val)
		}
	}
	return filtered
}
