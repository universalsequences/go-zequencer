package main

func getRecentDiscogs(caches *Caches, recentSounds []SampleResult) []SampleResult {
	ids := []interface{}{}
	for _, sound := range recentSounds {
		ids = append(ids, sound.IpfsHash)
	}

	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: NewDiscogsSample,
		SelectStatements: []string{
			"sampleHash",
			"coverArtHash",
			"discogsId",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "sampleHash",
				ValueList: ids,
			}},
		FromBlockNumber: 1,
	};

	cache := (*caches)[GUILD_SAMPLES] 
	results := queryForCache(cache, query)
	convertedResults := []SampleResult{}
	for _, result := range results {
		coverArtHash, err := result["coverArtHash"].(string)
		convertedResult := SampleResult{
			IpfsHash: result["sampleHash"].(string),
			DiscogsId: result["discogsId"].(float64),
		}
		if !err {
			convertedResult.CoverArtHash = coverArtHash
		}
		convertedResults = append(convertedResults, convertedResult)
	}
	return convertedResults
}



	

