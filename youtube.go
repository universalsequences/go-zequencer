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


