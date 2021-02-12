package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"sort"
	"log"
	"net/http"
	"strings"
	"fmt"
)
	
const blocksPerDay = 6275
//const blocksPerDay = 6275*31;

const pageSize = 15;

type SearchQuery struct {
	SearchTerm string `json:"searchTerm"`
	GuildIds []float64 `json:"guildIds"`
	GroupBy string `json:"groupBy"`
	Year float64 `json:"year"`
	Start float64 `json:"start"`
	FilterFavorites bool `json:"filterFavorites"`
}

type QueryResults struct {
	Unsourced map[string][]SampleResult `json:"unsourced"`
	Youtube map[string][]SampleResult `json:"youtube"`
	Discogs map[string][]SampleResult `json:"discogs"`
	Resampled map[string][]SampleResult `json:"resampled"`
	BlockNumber float64 `json:"blockNumber"`
	Results []SampleResult `json:"results"`
};

type SampleResult struct {
	Title string `json:"title"`
	IpfsHash string `json:"ipfsHash"`
	Tags []string `json:"tags"`
	VideoId string `json:"videoId"`
	Rating int `json:"rating"`
	CoverArtHash string `json:"coverArtHash"`
	ReleaseName string `json:"releaseName"`
	ArtistName string `json:"artistName"`
	ReleaseId float64 `json:"releaseId"`
	DiscogsId float64 `json:"discogsId"`
	GuildId float64 `json:"guildId"`
	User string `json:"user"`
	BlockNumber float64 `json:"blockNumber"`
}

type BlockResults struct {
	Unsourced map[string][]SampleResult `json:"unsourced"`
	Youtube map[string][]SampleResult `json:"youtube"`
	Discogs map[string][]SampleResult `json:"discogs"`
	BlockNumber float64 `json:"blockNumber"`
	Results []SampleResult `json:"results"`
	Tag string `json:"tag"`
}

func HandleSearchQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	ratingsCache *RatingCache,
	cachedQueries *CachedSearchQueries) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)
	query := SearchQuery{}
	json.Unmarshal([]byte(bodyString), &query)
	start := query.Start
	query.Start = 0

	// unpaginated queryr
	unpaginatedQueryBytes, _ := json.Marshal(query)
	unpaginatedQueryString := string(unpaginatedQueryBytes)

	if unpaginatedResults, ok := cachedQueries.getQuery(unpaginatedQueryString); ok {
		ret, _ := json.Marshal(paginate(start, unpaginatedResults))
		w.Write(ret)
		return
	}

	unpaginatedResults := runQuery(caches, ratingsCache, query)
	
	cachedQueries.newQuery(unpaginatedQueryString, unpaginatedResults)

	paginatedResults := paginate(start, unpaginatedResults)
	ret, _ := json.Marshal(paginatedResults)
	w.Write(ret)
}

func prePopulateCache(caches *Caches, ratingsCache *RatingCache, cachedQueries *CachedSearchQueries) {
	bodyString := "{\"searchTerm\":\"\",\"guildId\":0}"
	query := SearchQuery{}
	cachedQueries.newQuery(bodyString, runQuery(caches, ratingsCache, query))
	fmt.Println("Finished pre-populating cache")
}

func paginate(start float64, results [] BlockResults) []BlockResults {
	if (int(start) >= len(results)) {
		return []BlockResults{}
	}
	if (int(start) + pageSize > len(results)) {
		return results[int64(start):len(results)]
	} else {
		return results[int64(start):int64(start+pageSize)]
	}
}

func runQuery(caches *Caches, ratingsCache *RatingCache, query SearchQuery) []BlockResults {
	recentSounds := getRecentSounds(
		caches,
		query.SearchTerm,
		query.GuildIds,
		query.Year,
		query.FilterFavorites)
	recentDiscogs := getRecentDiscogs(caches, recentSounds)
	recentArtists := getRecentArtists(caches, recentSounds)
	recentReleases := getRecentReleases(caches, recentDiscogs)
	recentYoutubes := getRecentYoutubeSamples(caches, recentSounds)
	recentTags := getRecentTags(caches, recentSounds)
	ratings := getRatings(ratingsCache, getSampleIds(recentSounds))
	// usedIn := getUsedIn(caches, recentSounds);
	sounds := combineAll(recentSounds, recentDiscogs, recentYoutubes, recentTags, ratings, recentReleases, recentArtists);
	var results []BlockResults
	if (query.GroupBy == "tag") {
		results = partitionBySource(getByTag(sounds, query.SearchTerm), query.SearchTerm)
	} else {
		results = partitionBySource(getByDay(sounds), query.SearchTerm)
	}
	return results;
}

func getSampleIds(recentSounds []SampleResult) []string {
	ids := []string{}
	for _, sound := range recentSounds {
		ids = append(ids, sound.IpfsHash)
	}
	return ids
}


func combineAll(
	sampleCreated []SampleResult,
	newDiscogs []SampleResult,
	sampleYoutube []SampleResult,
	tags map[string][]string,
	ratings map[string]int,
	recentReleases map[string]Release,
	recentArtists map[string]string) []SampleResult {

	combined := map[string]*SampleResult{}

	for _, result := range sampleCreated {
		combined[result.IpfsHash] = &SampleResult{
			IpfsHash: result.IpfsHash,
			Title: result.Title,
			BlockNumber: result.BlockNumber,
			Tags: tags[result.IpfsHash],
			Rating: ratings[result.IpfsHash],
			User: result.User,
		};
	}

	for ipfsHash, artistName := range recentArtists {
		combined[ipfsHash].ArtistName = artistName
	}

	for ipfsHash, result := range recentReleases {
		combined[ipfsHash].CoverArtHash = result.CoverArtHash
		combined[ipfsHash].ReleaseName = result.ReleaseName
		combined[ipfsHash].ArtistName = result.ArtistName
		combined[ipfsHash].ReleaseId = result.ReleaseId
	}

	for _, result := range newDiscogs {
		combined[result.IpfsHash].DiscogsId = result.DiscogsId
	}

	for _, result := range sampleYoutube {
		combined[result.IpfsHash].VideoId = result.VideoId
	}

	results := []SampleResult{}
	for _, result := range sampleCreated {
		results = append(results, *combined[result.IpfsHash])
	}
	return results
}

type TempBlock struct {
	BlockNumber float64
	Results *[]SampleResult
};

func getByTag(results []SampleResult, searchTerm string) []BlockResults {
	tagToResults := map[string][]SampleResult{}
	for _, result := range results {
		for _, tag := range result.Tags {
			if (len(result.Tags) > 1 && tag == searchTerm) {
				continue
			}
			trimmed := strings.TrimSpace(tag)
			tagToResults[trimmed] = append(
				tagToResults[trimmed], result)
		}
	}

	blockResults := []BlockResults{}
	samplesProcessed := map[string]bool{}
	for tag, results := range tagToResults {

		if (strings.HasPrefix(tag, "0")) {
			continue
		}
		resultsToProcess := []SampleResult{}
		for _, result := range results {
			if _, ok := samplesProcessed[result.IpfsHash]; ok {
				continue
			}
			samplesProcessed[result.IpfsHash] = true
			resultsToProcess = append(
				resultsToProcess,
				result)
		}

		if (len(resultsToProcess) == 0) {
			continue
		}
		blockResults = append(
			blockResults,
			BlockResults{
				Tag: tag,
				Results: resultsToProcess,
			})
	}

	sort.Sort(ByTag(blockResults))

	return blockResults
}

type ByTag []BlockResults
func (a ByTag) Len() int           { return len(a) }
func (a ByTag) Less(i, j int) bool { return strings.Compare(strings.ToLower(a[i].Tag),strings.ToLower(a[j].Tag)) < 0 }
func (a ByTag) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }



func getByDay(results []SampleResult) []BlockResults{
	// first try to find blocks that are within a certain range
	blocks := []*TempBlock{}

	for _, result := range results {
		if (len(blocks) > 0) {
			// check if last block
			lastBlock := blocks[len(blocks) - 1]
			lastBlockNumber := lastBlock.BlockNumber
			if (math.Abs(lastBlockNumber - result.BlockNumber) < blocksPerDay) { 
				fuck := append(
					*lastBlock.Results,
					result)
				lastBlock.Results = &fuck
				continue
			}
		}
		// otherwise we wanna just append it to this one
		block := TempBlock{
			BlockNumber: result.BlockNumber,
			Results: &([]SampleResult{result}),
		}
		blocks = append(blocks, &block)
	}

	realBlocks := []BlockResults{}
	for _, block := range blocks {
		realBlocks = append(
			realBlocks,
			BlockResults{
				BlockNumber: block.BlockNumber,
				Results: *block.Results,
			})
	}
	return realBlocks;
};

func partitionBySource(byDay [] BlockResults, searchTerm string) []BlockResults {
	partitioned := []BlockResults{}
	for _, toPartition := range byDay {
		results := toPartition.Results;
		dayResult := BlockResults{
			BlockNumber: toPartition.BlockNumber,
			Tag: toPartition.Tag,
			Results: toPartition.Results,
			Discogs: map[string][]SampleResult{},
			Youtube: map[string][]SampleResult{},
			Unsourced: map[string][]SampleResult{},
		}
		for _, result := range results {
			if (result.DiscogsId != 0) {
				if _, ok := dayResult.Discogs[result.CoverArtHash]; !ok {
					dayResult.Discogs[result.CoverArtHash] = []SampleResult{}
				} 
				dayResult.Discogs[result.CoverArtHash] = append(
					dayResult.Discogs[result.CoverArtHash],
					result);
			} else if (result.VideoId != "") {
				if _, ok := dayResult.Youtube[result.VideoId]; !ok {
					dayResult.Youtube[result.VideoId] = []SampleResult{}
				}
				dayResult.Youtube[result.VideoId] = append(
					dayResult.Youtube[result.VideoId],
					result);
			} else {
				tag := "";
				if (len(result.Tags) > 0) {
					tag = result.Tags[0]
				}

				found := findInList(result.Tags, strings.ToLower(searchTerm))
				if (searchTerm != "" && found != "") {
					tag = found
				}
				if _, ok := dayResult.Unsourced[tag]; !ok {
					dayResult.Unsourced[tag] = []SampleResult{};
				}
				dayResult.Unsourced[tag] = append(
					dayResult.Unsourced[tag],
					result);
			}
		}
		partitioned = append(partitioned, dayResult)
	}
	return partitioned;
};

func findInList(tags [] string, searchTerm string) string {
	if searchTerm == "" {
		if (len(tags) == 0) {
			return ""
		} else {
			return tags[0]
		}
	}

	for _, tag := range tags {
	if !strings.Contains(strings.ToLower(tag), searchTerm) &&
		tag != "Resampled" {
			return tag
		}
	}
	return ""
}


