package main

func FilterSoundsByYears(caches *Caches, startYear float64, endYear float64, sounds []string) []string {
	sampleMap := map[string]bool{}
	for i, _ := range sounds {
		sampleMap[sounds[i]] = true
	}

	years := make([]interface{}, int(endYear - startYear))

	for i := 0; int(startYear) + i < int(endYear); i++ {
		years[i] = startYear + float64(i)
	}

	query := NewQuery(GUILD_SAMPLES)
	query.From(SampleYear)
	query.WhereIn("year", years)

	results := query.ExecuteQuery(caches);

	
	soundYear := map[string]bool{}

	for _, result := range results {
		if _, ok := sampleMap[result["ipfsHash"].(string)]; ok {
			soundYear[result["ipfsHash"].(string)] = true
		}
	}

	resultingSounds := []string{}
	for sound, _ := range soundYear {
		resultingSounds = append(resultingSounds, sound)
	}
	
	return resultingSounds
}
