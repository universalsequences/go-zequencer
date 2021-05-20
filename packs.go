package main

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"strings"
	"encoding/json"
	"net/http"
)

const PRESETS_CONTRACT = "0x45aC8aCbEba84071D4e549d4dCd273E01E5a8daF";
const PACKS_CONTRACT = "0xF7bd2ada59c4ab5AD0f6BFbE94EB4f8eCa18eEDd";

type PackQuery struct {
	SearchTerm string `json:"searchTerm"`
	IncludeContent bool `json:"includeContent"`
	ContentType string `json:"contentType"`
}

type PackMatchResult struct {
	Matches bool
	MatchingContent []string
}

func HandlePackQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)
	query := PackQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	results := runPackQuery(caches, query.SearchTerm, query.IncludeContent, query.ContentType)

	ret, _ := json.Marshal(results)
	w.Write(ret)
}

func runPackQuery(caches *Caches, searchTerm string, includeContent bool, contentType string) []interface{} {
	// first get all the packs
	query := NewQuery(PACKS_CONTRACT)
	query.From(NewPack)
	allPacks := query.ExecuteQuery(caches)
	hashes := [] interface{}{}
	for _, pack := range allPacks {
		hashes = append(hashes, pack["packHash"])
	}

	packHashes := pruneOldPacks(caches, hashes)

	packMap := map[string]bool{}
	for _, hash := range packHashes {
		packMap[hash] = true
	}

	searchTerm = strings.ToLower(searchTerm)

	results := []interface{}{}
	for _, result := range allPacks {
		packHash := result["packHash"].(string)
		if _, ok := packMap[packHash]; ok {
			matchData := packMatchesSearch(caches, result, searchTerm, includeContent, contentType)
			result["matchingContent"] = matchData.MatchingContent
			if (matchData.Matches) {
				results = append(results, result)
			}
		}
	}
	return results
}

func packMatchesSearch(
	caches *Caches,
	pack map[string]interface{},
	searchTerm string,
	includeContent bool,
	contentType string) PackMatchResult {

	title := pack["title"].(string)
	guildIds := []interface{}{}

	matchingContent := []string{}
	allContent := []string{}

	// go through the content of each pack
	contentHashes := getSoundsWithPack(caches, pack["packHash"].(string), guildIds)
	for _, contentHash := range contentHashes {
		// figure out if its a sound or preset
		sampleInfo := GetSampleInformation(caches, contentHash.(string), "")
		tags := []string{}
		title := ""
		if (sampleInfo.Title != "") {
			// then this is a sample!
			if (contentType == "PRESET") {
				continue
			}
			tags = sampleInfo.Tags
			title = sampleInfo.Title
		} else {
			if (contentType == "SOUND") {
				continue
			}
			// then its a preset so get the title of the preset
			// and the tags
			query := NewQuery(PRESETS_CONTRACT)
			query.From(NewPreset)
			query.WhereIs("contentHash", contentHash)
			results := query.ExecuteQuery(caches)
			if (len(results) == 0) {
				continue
			}
			hexTitle := results[0]["encryptedName"].(string)
			decoded, err := hex.DecodeString(hexTitle)
			if err == nil {
				title = fmt.Sprintf("%s", decoded)
			}
			query = NewQuery(PRESETS_CONTRACT)
			query.From(PresetTagged)
			query.WhereIs("contentHash", contentHash)
			results = query.ExecuteQuery(caches)
			for _, result := range results {
				tags = append(tags, result["tag"].(string))
			}
		}

		allContent = append(
			allContent, contentHash.(string))

		// check if the title matches
		if (strings.Contains(strings.ToLower(title), searchTerm)) {
			matchingContent = append(
				matchingContent, contentHash.(string))
			continue
		}
		// check if the tags match
		for _, tag := range tags {
			if (strings.Contains(strings.ToLower(tag), searchTerm)) {
				matchingContent = append(
					matchingContent, contentHash.(string))
				continue
			}
		}
	}
	titleMatches := strings.Contains(strings.ToLower(title), searchTerm)
	if (titleMatches && !includeContent) {
		matchingContent = [] string{}
	}
		
	matches := len(matchingContent) > 0 || titleMatches
	if (includeContent && contentType != "" && (len(matchingContent) == 0 || titleMatches)) {
		if ((searchTerm == "" || !titleMatches) && len(matchingContent) == 0) {
			matches = false
		} else {
			matchingContent = allContent
		}
	}
	match := PackMatchResult{
		Matches: matches,
		MatchingContent: matchingContent,
	}
	return match
}

func getSoundsWithPack(caches *Caches, pack string, guildIds []interface{}) []interface{} {
	query := NewQuery(PACKS_CONTRACT)
	query.From(PackHasContent)
	query.Select("contentHash")
	query.WhereIs("packHash", pack)
	//query.WhereIn("guildId", guildIds)
	results := queryForCache((*caches)[PACKS_CONTRACT], query)

	soundIds := [] interface {}{}
	for _, result := range results {
		soundIds = append(
			soundIds,
			result["contentHash"])
	}

	return soundIds 
}

func getPacksWithSound(caches *Caches, soundId string) []string {
	query := NewQuery(PACKS_CONTRACT)
	query.From(PackHasContent)
	query.Select("packHash")
	query.WhereIs("contentHash", soundId)
	results := queryForCache((*caches)[PACKS_CONTRACT], query)

	packIds := []interface{}{}
	for _, result := range results {
		packIds = append(
			packIds,
			result["packHash"])
	}

	return getPackNames(caches, pruneOldPacks(caches, packIds))
}

func pruneOldPacks(caches *Caches, packHashes[] interface{}) []string {
	query := NewQuery(PACKS_CONTRACT)
	query.From(NewPack)
	query.Select("packHash")
	query.WhereIn("packHash", packHashes)
	results := query.ExecuteQuery(caches)
	oldHashes := map[string]bool{}
	for _, result := range results {
		if (result["packHash"] != result["previousHash"] &&
			result["previousHash"] != nil) {
			oldHashes[result["previousHash"].(string)] = true
		}
	}

	pruned := []string{}
	for _, hash := range packHashes {
		if _, ok := oldHashes[hash.(string)]; !ok {
			pruned = append(
				pruned,
				hash.(string))
		}
	}

	return pruned
}


func getPackNames(caches *Caches, packHashes[] string) []string {
	packs := []interface{}{}
	for _, pack := range packHashes {
		packs = append(
			packs, pack)
	}
	query := NewQuery(PACKS_CONTRACT)
	query.From(NewPack)
	query.Select("title")
	query.WhereIn("packHash", packs)
	results := query.ExecuteQuery(caches)
	titles := []string{}
	for _, result := range results {
		titles = append(
			titles, result["title"].(string))
	}
	return titles
}


