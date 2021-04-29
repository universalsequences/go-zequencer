
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

type strSlice []string

func (slice strSlice) pos(value string) int {
    for p, v := range slice {
        if (v == value) {
            return p
        }
    }
    return -1
}

type SearchQuery struct {
	SearchTerm string `json:"searchTerm"`
	GuildIds []float64 `json:"guildIds"`
	GroupBy string `json:"groupBy"`
	Year float64 `json:"year"`
	Start float64 `json:"start"`
	FilterFavorites bool `json:"filterFavorites"`
	ReleaseId float64 `json:"releaseId"`
	VideoId string `json:"videoId"`
	Tag string `json:"tag"`
	Size float64 `json:"size"`
	User string `json:"user"`
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
	Results []SampleResult `json:"results"`
	BlockNumber float64 `json:"blockNumber"`
	Tag string `json:"tag"`
	TotalResults int `json:"totalResults"`
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
	tag := query.Tag
	searchTerm := query.SearchTerm
	groupBy := query.GroupBy
	query.Start = 0
	query.Tag = ""
	isLazy := tag == "" && searchTerm == "" && groupBy == "tag" && query.ReleaseId == 0 && query.VideoId == "";

	// unpaginated queryr
	unpaginatedQueryBytes, _ := json.Marshal(query)
	unpaginatedQueryString := string(unpaginatedQueryBytes)

	if unpaginatedResults, ok := cachedQueries.getQuery(unpaginatedQueryString); ok {
		paginated := []BlockResults{}
		if (tag != "") {
			paginated = filterByTag(tag, unpaginatedResults)
		} else {
			paginated = paginate(start, unpaginatedResults, query.Size)
			if (isLazy) {
				paginated = removeDeepResults(paginated)
			}
		}
		ret, _ := json.Marshal(paginated)
		w.Write(ret)
		return
	} 

	unpaginatedResults := runQuery(caches, ratingsCache, query)
	
	cachedQueries.newQuery(unpaginatedQueryString, unpaginatedResults)

	paginatedResults := paginate(start, unpaginatedResults, query.Size)

	if (isLazy) {
		paginatedResults = removeDeepResults(paginatedResults)
	}
	ret, _ := json.Marshal(paginatedResults)
	w.Write(ret)
}

func prePopulateCache(caches *Caches, ratingsCache *RatingCache, cachedQueries *CachedSearchQueries) {
	bodyString := "{\"searchTerm\":\"\",\"guildId\":0}"
	query := SearchQuery{}
	cachedQueries.newQuery(bodyString, runQuery(caches, ratingsCache, query))
	fmt.Println("Finished pre-populating cache")
}

func removeDeepResults(results []BlockResults) []BlockResults {
	lazyResults := []BlockResults{}
	for _, result := range results {
		lazyResults = append(
			lazyResults,
			BlockResults{
				Tag: result.Tag,
				TotalResults: result.TotalResults,
				BlockNumber: result.BlockNumber,
			})
	}
	return lazyResults
}
func filterByTag(tag string, results [] BlockResults) []BlockResults {
	tagResults := []BlockResults{}
	for _, result := range results {
		if (result.Tag == tag) { 
			tagResults = append(tagResults, result)
		}
	}
	return tagResults
}

func paginate(start float64, results [] BlockResults, pageSize float64) []BlockResults {
	if (int(start) >= len(results)) {
		return []BlockResults{}
	}
	if (int(start) + int(pageSize) > len(results)) {
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
		query.FilterFavorites,
		query.ReleaseId,
		query.VideoId,
		query.User)
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
		results = partitionBySource(
			getByTag(
				sounds,
				query.SearchTerm,
				query.ReleaseId,
				query.VideoId),
			query.SearchTerm,
			query.VideoId != "" || query.ReleaseId != 0.0)
	} else {
		results = partitionBySource(getByDay(sounds), query.SearchTerm, false)
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

func getByTag(results []SampleResult, searchTerm string, releaseId float64, videoId string) []BlockResults {
	tagToResults := map[string][]SampleResult{}
	for _, result := range results {
		for _, tag := range result.Tags {
			if (searchTerm != "" && strings.Contains(tag, strings.ToLower(searchTerm)) && len(result.Tags) > 1) {
				continue
			}
			trimmed := strings.TrimSpace(tag)
			tagToResults[trimmed] = append(
				tagToResults[trimmed], result)
		}
	}

	tags := strSlice {}
	for tag, _ := range tagToResults {
		tags = append(tags, tag)
	}

	sort.Slice(tags, func (i, j int) bool {
		return len(tagToResults[tags[i]]) > len(tagToResults[tags[j]]);
	});

	blockResults := []BlockResults{}
	samplesProcessed := map[string]bool{}
	for _, tag := range tags {
		results := tagToResults[tag];
		if (strings.HasPrefix(tag, "0")) {
			continue
		}
		resultsToProcess := []SampleResult{}
		for _, result := range results {
			// TODO: in order to properly break stuff up
			// If searchTerm=KORG and tags=[KORG, KORG M1 Kick],
			// choose the second most popular tag (i.e. KORG M1 Kick).
			// If tag !== this second most popular tag, then skip it
			// It will be caught when that tag is processed
			if (searchTerm != "") {
				secondMostPopular := getSecondMostPopularTag(tags, result.Tags)
				if (strings.Contains(strings.ToLower(tag), strings.ToLower(searchTerm)) &&
					secondMostPopular != "" && secondMostPopular != tag) {
					continue
				}
			}

			if (searchTerm != "" || releaseId != 0.0 || videoId != "") {
				if _, ok := samplesProcessed[result.IpfsHash]; ok {
					continue
				}
			}
			samplesProcessed[result.IpfsHash] = true
			resultsToProcess = append(
				resultsToProcess,
				result)
		}

		if (len(resultsToProcess) == 0) {
			continue
		}
		// if every result to process contains another tag that
		// is ranked higher than dont do append
		skip := true
		if (searchTerm == "") {
			position := tags.pos(tag)
			for _, r := range resultsToProcess {
				min := 1000000000
				for _, t := range r.Tags {
					_position := tags.pos(t)
					if (_position < min) {
						min = _position
					}
				}
				if (min < position) {
					// contains a tag that is ranked higher
				} else {
					skip = false
					break
				}
			}
		} else {
			skip = false
		}
		
		if (!skip) {
			blockResults = append(
				blockResults,
				BlockResults{
					Tag: tag,
					Results: resultsToProcess,
				})
		}
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

func partitionBySource(byDay [] BlockResults, searchTerm string, keepTags bool) []BlockResults {
	partitioned := []BlockResults{}
	for _, toPartition := range byDay {
		results := toPartition.Results;
		dayResult := BlockResults{
			BlockNumber: toPartition.BlockNumber,
			Tag: toPartition.Tag,
			TotalResults: len(toPartition.Results),
			Results: toPartition.Results,
			Discogs: map[string][]SampleResult{},
			Youtube: map[string][]SampleResult{},
			Unsourced: map[string][]SampleResult{},
		}
		for _, result := range results {
			if (result.DiscogsId != 0 && !keepTags) {
				if _, ok := dayResult.Discogs[result.CoverArtHash]; !ok {
					dayResult.Discogs[result.CoverArtHash] = []SampleResult{}
				} 
				dayResult.Discogs[result.CoverArtHash] = append(
					dayResult.Discogs[result.CoverArtHash],
					result);
			} else if (result.VideoId != "" && !keepTags) {
				if _, ok := dayResult.Youtube[result.VideoId]; !ok {
					dayResult.Youtube[result.VideoId] = []SampleResult{}
				}
				dayResult.Youtube[result.VideoId] = append(
					dayResult.Youtube[result.VideoId],
					result);
			} else {
				tag := getMostPopularTag(result.Tags, dayResult.Unsourced, strings.ToLower(toPartition.Tag), strings.ToLower(searchTerm))
				found := tag //findInList(result.Tags, strings.ToLower(searchTerm))
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


func getMostPopularTag(tags [] string, unsourced map[string][]SampleResult, a string, b string) string {
	max := -1
	maxTag := ""
	for _, tag := range tags {
		if (strings.Contains(strings.ToLower(tag), a)) {
			continue
		}
		if (strings.Contains(strings.ToLower(tag), b)) {
			continue
		}
		if (strings.Contains(strings.ToLower(tag), "Resampled")) {
			continue
		}
		tagLength := len(unsourced[tag])
		if (tagLength > max) {
			maxTag = tag
			max = tagLength
		}
	}
	if (maxTag == "" && len(tags) > 0) {
		return tags[0];
	}
	return maxTag
}
	
func getSecondMostPopularTag(orderedTags strSlice, tags []string) string {
	t := []string{}
	for _, tag := range tags {
		t = append(t, tag);
	}

	sort.Slice(t, func (i, j int) bool {
		return orderedTags.pos(t[i]) > orderedTags.pos(t[j])
	});

	if (len(t) >= 2) {
		return t[1]
	} else {
		return ""
	} 
}
