package main

import (
	"fmt"
)
	
const SEQUENCE_METADATA = "0x03172e863dbF7EEb3994AF5a6608c470eB6E79fC"
const XANADU = "0x305306F68D9C230B59d5B6869AEd1723365C9290"
const PROJECT_METADATA = "0xD3d9ee5d5467c1C9e4Ae3dA5e225882F46Bd45aA"


type ProjectMetadata struct {
	PatternsCount float64 `json:"patternsCount"`
	Visuals string `json:"thumbnailHash"`
	Presets []string `json:"presets"`
	BPM float64 `json:"BPM"`
}

func GetProjectsMetadata(caches *Caches, projectHashes []string ) map[string]ProjectMetadata {
	metadatas := map[string]ProjectMetadata{}

	query := NewQuery(PROJECT_METADATA)
	query.From(ProjectPatternsCount)

	_projectHashes := []interface{}{}
	for _, result := range(projectHashes) {
		_projectHashes = append(_projectHashes, result)
	}

	query.WhereIn("projectHash", _projectHashes)
	results := query.ExecuteQuery(caches)
	patterns := map[string]float64{}
	for _, result := range(results) {
		projectHash := result["projectHash"].(string)
		patternsCount := result["patterns"].(float64)
		patterns[projectHash] = patternsCount
	}

	query = NewQuery(PROJECT_METADATA)
	query.From(ProjectVisuals)

	query.WhereIn("projectHash", _projectHashes)
	results = query.ExecuteQuery(caches)
	visuals := map[string]string{}

	for _, result := range(results) {
		projectHash := result["projectHash"].(string)
		thumbnail := result["thumbnailhash"].(string)
		visuals[projectHash] = thumbnail
	}

	query = NewQuery(PROJECT_METADATA)
	query.From(ProjectBPM)
	results = query.ExecuteQuery(caches)
	bpms := map[string]float64{}
	for _, result := range(results) {
		projectHash := result["projectHash"].(string)
		bpm := result["bpm"].(float64)
		bpms[projectHash] = bpm
	}

	for id, patternsCount := range(patterns) {

		if (patternsCount > 0) {
			fmt.Printf("Project %v had %v many patterns\n", id, patternsCount)
		}
		if thumbnail, ok := visuals[id]; ok {
			metadatas[id] = ProjectMetadata{
				PatternsCount: patternsCount,
				Visuals: thumbnail,
				BPM: bpms[id],
			}
		} else {
			metadatas[id] = ProjectMetadata{
				PatternsCount: patternsCount,
				BPM: bpms[id],
			}
		}
	}
	
	return metadatas;
}

func GetProjectTags(caches *Caches) map[string][]string {
	projectTags := map[string][]string{}
	query := NewQuery(XANADU)
	query.From(NewAnnotation)
	query.Select("data") // ipfs hash of project
	query.Select("annotationData") // the tag
	query.WhereIs("annotationType", SEQUENCE_TAG)
	results := query.ExecuteQuery(caches)

	for _, result := range results {
		if _, ok := result["annotationData"].(string); !ok {
			continue;
		}
		tag := result["annotationData"].(string)
		projectHash := result["data"].(string)
		if _, ok := projectTags[projectHash]; !ok {
			projectTags[projectHash] = []string{
				tag,
			}
		} else {
			projectTags[projectHash] = append(
				projectTags[projectHash],
				tag,
			)
		}
	}

	return projectTags
}

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

