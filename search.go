package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"log"
	"net/http"
	"strings"
	"fmt"
)
	
const blocksPerDay = 6275

type SearchQuery struct {
	SearchTerm string `json:"searchTerm"`
}

type QueryResults struct {
	Unsourced map[string][]SampleResult `json:"unsourced"`
	Youtube map[string][]SampleResult `json:"youtube"`
	Discogs map[string][]SampleResult `json:"discogs"`
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
}

func HandleSearchQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	ratingsCache *RatingCache,
	cachedQueries *CachedQueries) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)

	if val, ok := cachedQueries.getQuery(bodyString); ok {
		w.Write(val)
		return
	}

	query := SearchQuery{}
	json.Unmarshal([]byte(bodyString), &query)
	bytes, err := json.Marshal(runQuery(caches, ratingsCache, query))
	cachedQueries.newQuery(bodyString, bytes)
	w.Write(bytes)
}

func prePopulateCache(caches *Caches, ratingsCache *RatingCache, cachedQueries *CachedQueries) {
	bodyString := "{\"searchTerm\":\"\"}"
	query := SearchQuery{}
	bytes, _ := json.Marshal(runQuery(caches, ratingsCache, query))
	cachedQueries.newQuery(bodyString, bytes)
	fmt.Println("Finished pre-populating cache")
}

func runQuery(caches *Caches, ratingsCache *RatingCache, query SearchQuery) []BlockResults {
	recentSounds := getRecentSounds(caches, query.SearchTerm)
	recentDiscogs := getRecentDiscogs(caches, recentSounds)
	recentReleases := getRecentReleases(caches, recentDiscogs)
	recentYoutubes := getRecentYoutubeSamples(caches, recentSounds)
	recentTags := getRecentTags(caches, recentSounds)
	ratings := getRatings(ratingsCache, getSampleIds(recentSounds))
	sounds := combineAll(recentSounds, recentDiscogs, recentYoutubes, recentTags, ratings, recentReleases);
	byDay := getByDay(sounds);
	return partitionBySource(byDay, query.SearchTerm);
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
	recentReleases map[string]Release) []SampleResult {

	combined := map[string]*SampleResult{}

	for _, result := range sampleCreated {
		combined[result.IpfsHash] = &SampleResult{
			IpfsHash: result.IpfsHash,
			Title: result.Title,
			BlockNumber: result.BlockNumber,
			Tags: tags[result.IpfsHash],
			Rating: ratings[result.IpfsHash],
		};
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


