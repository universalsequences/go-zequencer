package main

const SEQUENCE_METADATA = "0x03172e863dbF7EEb3994AF5a6608c470eB6E79fC"

func GetStarredProjects(caches *Caches) map[string]bool {
	starredQuery := NewQuery(SEQUENCE_METADATA)
	starredQuery.From(SequenceStarred)
	starredResults := starredQuery.ExecuteQuery(caches)

	unstarredQuery := NewQuery(SEQUENCE_METADATA)
	unstarredQuery.From(SequenceUnstarred)
	unstarredResults := unstarredQuery.ExecuteQuery(caches)

	counts := map[string]int{}

	for _, starredResult := range starredResults {
		id := starredResult["ipfsHash"].(string)
		if _, ok := counts[id]; !ok {
			counts[id] = 0
		}
		counts[id]++
	}
	for _, unstarredResult := range unstarredResults {
		id := unstarredResult["ipfsHash"].(string)
		if _, ok := counts[id]; !ok {
			counts[id] = 0
		}
		counts[id]--
	}

	starredProjects := map[string]bool{}
	for project, count := range counts {
		if (count > 0) {
			starredProjects[project] = true
		}
	}
	return starredProjects
}

func GetFavoritedProjects(caches *Caches, user string) map[string]bool {
	favoriteQuery := NewQuery(SEQUENCE_METADATA)
	favoriteQuery.From(SequenceFavorited)
	favoriteQuery.Select("ipfsHash")
	favoriteQuery.WhereIs("user", user)
	favoriteResults := favoriteQuery.ExecuteQuery(caches)

	counts := map[string]int{}

	for _, result := range favoriteResults {
		id := result["ipfsHash"].(string)
		if _, ok := counts[id]; !ok {
			counts[id] = 0
		}
		counts[id]++
	}


	unfavoriteQuery := NewQuery(SEQUENCE_METADATA)
	unfavoriteQuery.From(SequenceUnfavorited)
	unfavoriteQuery.Select("ipfsHash")
	unfavoriteQuery.WhereIs("user", user)

	unfavoriteResults := unfavoriteQuery.ExecuteQuery(caches)
	for _, result:= range unfavoriteResults {
		id := result["ipfsHash"].(string)
		if _, ok := counts[id]; !ok {
			counts[id] = 0
		}
		counts[id]--
	}

	favorited := map[string]bool{}
	for project, count := range counts {
		if (count > 0) {
			favorited[project] = true
		}
	}
	return favorited 
}

