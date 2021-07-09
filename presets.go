package main

import (
	"fmt"
	"io/ioutil"
	"encoding/hex"
	"log"
	"encoding/json"
	"net/http"
)

const PRESETS_CONTRACT = "0x45aC8aCbEba84071D4e549d4dCd273E01E5a8daF";

type PresetQuery struct {
	User string `json:"user"`
	SearchTerm string `json:"searchTerm"`
	InstrumentType string `json:"instrumentType"`
	GuildId float64 `json:"guildId"`
	FilterFavorites bool `json:"filterFavorites"`
	ContentHashes []string `json:"contentHashes"`
}

func HandlePresetQuery(
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
	query := PresetQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	results := runPresetQuery(caches, query)

	ret, _ := json.Marshal(results)
	w.Write(ret)
}

func runPresetQuery(caches *Caches, query PresetQuery) []map[string]interface{} {
	instQuery := NewQuery(PRESETS_CONTRACT)
	instQuery.From(PresetInstrumentType)
	instQuery.WhereIs("instrumentType", query.InstrumentType);

	if (len(query.ContentHashes) > 0) {
		// filter by content hashes
		contentHashes := []interface{}{}
		for _, hash := range query.ContentHashes {
			contentHashes = append(contentHashes, hash)
		}
		instQuery.WhereIn("contentHash", contentHashes)
	}

	instResults := []interface{}{}
	for _, result := range instQuery.ExecuteQuery(caches) {
		instResults = append(instResults, result["contentHash"])
	}

	presetsQuery := NewQuery(PRESETS_CONTRACT)
	presetsQuery.From(NewPreset)
	presetsQuery.WhereIn("contentHash", instResults)
	presetsQuery.WhereIs("guildId", query.GuildId)
	
	rawResults := presetsQuery.ExecuteQuery(caches)

	allContentHashes := []interface{}{}
	for _, result := range rawResults {
		allContentHashes = append(allContentHashes, result["contentHash"].(string))
	}

	tags := getTagsForPresets(caches, allContentHashes)
	packs := getPacksForPresets(caches, allContentHashes)
	favorites := getSoundsWithRating(caches, 5, query.User, allContentHashes, "PRESET_RATED")
	
	results := []map[string]interface{}{}
	
	for _, result := range rawResults {
		row := map[string]interface{}{}
		contentHash := result["contentHash"].(string)
		hexTitle := result["encryptedName"].(string)
		decoded, err := hex.DecodeString(hexTitle)
		if err != nil {
			continue
		}
		title := fmt.Sprintf("%s", decoded)
		row["title"] = title
		if t, ok := tags[contentHash]; ok {
			row["tags"] = t
		} else {
			list := [] string{}
			row["tags"] = list
		}
		if _, ok := favorites[contentHash]; ok {
			row["favorited"] = true
		} else if (query.FilterFavorites) {
			// since we are filtering favorites and this is not
			// favorited, we skip
			continue;
		}
		if p, ok := packs[contentHash]; ok {
			row["packs"] = p
		} else {
			list := [] string{}
			row["packs"] = list
		}
		row["blockNumber"] = result["blockNumber"]
		row["contentHash"] = result["contentHash"]
		row["user"] = result["user"]
		row["guildId"] = result["guildId"]
		if (query.GuildId != 0.0) {
			row["encryptedName"] = result["encryptedName"]
			row["publicKey"] = result["publicKey"]
			row["encryptedContentKey"] = result["encryptedContentKey"]
		}
		results = append(results, row)	
	}
	return pruneContent(results, "contentHash")
}

func getTagsForPresets(caches *Caches, contentHashes []interface{}) map[string][]string {
	query := NewQuery(PRESETS_CONTRACT)
	query.From(PresetTagged)
	query.WhereIn("contentHash", contentHashes)
	results := query.ExecuteQuery(caches)
	tags := map[string][]string{}
	for _, result := range results {
		contentHash := result["contentHash"].(string)
		tag := result["tag"].(string)
		if _, ok := tags[contentHash]; !ok {
			list := []string{}
			tags[contentHash] = list
		}
		tags[contentHash] = append(tags[contentHash], tag)
	}
	return tags
}

func getPacksForPresets(caches *Caches, contentHashes []interface{}) map[string][]string {
	query := NewQuery(PACKS_CONTRACT)
	query.From(PackHasContent)
	query.Select("packHash")
	query.WhereIn("contentHash", contentHashes)
	results := query.ExecuteQuery(caches)

	packIds := []interface{}{}
	for _, result := range results {
		packIds = append(
			packIds,
			result["packHash"])
	}

	pruned := pruneOldPacks(caches, packIds)
	packNames := getPackNamesMap(caches, pruned)
	prunedPacks := map[string]bool{}
	for _, pack := range pruned {
		prunedPacks[pack] = true
	}
	
	packs := map[string][]string{}
	for _, result := range results {
		packHash := result["packHash"].(string)
		contentHash := result["contentHash"].(string)
		// first make sure its in the pruned packs
		if _, ok := prunedPacks[packHash]; !ok {
			continue
		}

		packName := packNames[packHash]
		if _, ok := packs[contentHash]; !ok {
			list := []string{}
			packs[contentHash] = list
		}
		packs[contentHash] = append(packs[contentHash], packName)
	}
	return packs
}






