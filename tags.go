package main

func getRecentTags(caches *Caches, recentSounds []SampleResult) map[string][]string {
	ids := []interface{}{}
	for _, sound := range recentSounds {
		ids = append(ids, sound.IpfsHash)
	}

	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
			"tag",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "ipfsHash",
				ValueList: ids,
			}},
		FromBlockNumber: 1,
	};

	cache := (*caches)[GUILD_SAMPLES] 
	results := queryForCache(cache, query)

	tags := map[string][]string{}

	for _, result := range results {
		ipfsHash := result["ipfsHash"].(string)
		if _, ok := tags[ipfsHash]; !ok {
			tags[ipfsHash] = []string{result["tag"].(string)}
		} else {
			if (!containsTag(tags[ipfsHash], result["tag"].(string))) {
				tags[ipfsHash] = append(tags[ipfsHash], result["tag"].(string))
			}
		}
	}
	return tags 
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if (t == tag) {
			return true
		}
	}
	return false
}

func getAllTags(caches *Caches) []string {
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
			"tag",
		},
		FromBlockNumber: 1,
	};

	cache := (*caches)[GUILD_SAMPLES] 
	results := queryForCache(cache, query)

	tagsMap := map[string]bool{}
	for _, result := range results {
		tagsMap[ result["tag"].(string)] = true
	}
	tags := []string{}
	for tag, _ := range tagsMap  {
		tags = append(tags, tag)
	}

	return tags
}

func getSoundsWithOrTags(caches *Caches, tags []interface{}) []interface{}{
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "tag",
				ValueList: tags,
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
