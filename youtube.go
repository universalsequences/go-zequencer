package main

func getRecentYoutubeSamples(caches *Caches, recentSounds []SampleResult) []SampleResult {
	ids := []interface{}{}
	for _, sound := range recentSounds {
		ids = append(ids, sound.IpfsHash)
	}

	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleYoutube,
		SelectStatements: []string{
			"ipfsHash",
			"videoId",
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

	convertedResults := []SampleResult{}

	for _, result := range results {
		convertedResult := SampleResult{
			IpfsHash: result["ipfsHash"].(string),
			VideoId: result["videoId"].(string),
		}
		convertedResults = append(convertedResults, convertedResult)
	}
	return convertedResults
}


func getSamplesFromVideo(caches *Caches, videoId string, guildIds []interface{}) []interface{} {
	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleYoutube)
	query.Select("ipfsHash")
	query.WhereIs("videoId", videoId)
	query.WhereIn("guildId", guildIds)
	results := queryForCache((*caches)[GUILD_SAMPLES], query)

	soundIds := [] interface {}{}
	for _, result := range results {
		soundIds = append(
			soundIds,
			result["ipfsHash"])
	}

	return soundIds 
}

