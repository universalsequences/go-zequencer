package main

const PACKS_CONTRACT = "0xF7bd2ada59c4ab5AD0f6BFbE94EB4f8eCa18eEDd";

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
	query.WhereIn("previousHash", packHashes)
	results := query.ExecuteQuery(caches)
	oldHashes := map[string]bool{}
	for _, result := range results {
		if (result["packHash"] != result["previousHash"] &&
			result["previousHash"] != nil) {
			oldHashes[result["packHash"].(string)] = true
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
