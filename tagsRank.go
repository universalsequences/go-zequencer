package main
import (
	"math"
	"sort"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type TagsRankQuery struct {
	Tags []string `json:"tags"`
	GuildIds []float64 `json:"guildIds"`
}

func HandleTagsRankQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
	) {

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	query := TagsRankQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	matchingTags := runTagsRankQuery(query, caches)

	bytes, err := json.Marshal(matchingTags)
	w.Write(bytes)
}
	

func runTagsRankQuery(
	query TagsRankQuery,
	caches *Caches) []string {

	tags := make([]interface{}, len(query.Tags))
	for i := range query.Tags {
		tags[i] = query.Tags[i]
	}

	guildIds := make([]interface{}, len(query.GuildIds))
	for i := range query.GuildIds {
		guildIds[i] = query.GuildIds[i]
	}

	_sounds := GetSamplesWithAndTags(caches, query.Tags, query.GuildIds)
	sounds := make([]interface{}, len(_sounds))

	for i, _ := range _sounds {
		sounds[i] = _sounds[i]
	}

	q := NewQuery(GUILD_SAMPLES)
	q.From(SampleTagged)
	q.WhereIn("ipfsHash", sounds)

	results := q.ExecuteQuery(caches)

	tagsMap := map[string]int{}

	for _, result := range results {
		if _, ok := tagsMap[result["tag"].(string)]; !ok {
			tagsMap[result["tag"].(string)] = 0
		}
		tagsMap[ result["tag"].(string)]++
	}

	
	sampleTags := []SampleTag{}
	for tag, count := range tagsMap {
		if (tag == "Resampled") {
			continue
		}
		t := make([]interface{}, 1)
		t[0] = tag
		count2 := len(getSoundsWithOrTags(caches, t, guildIds))
		count3 := math.Pow(float64(count2), 3.0)
		sampleTags = append(sampleTags, SampleTag{
			Tag: tag,
			Count: count*int(count3),
		})
	}
	
	sort.Sort(BySampleTag(sampleTags))

	rawTags := make([]string, len(sampleTags))
	for i, _ := range sampleTags {
		rawTags[i] = sampleTags[i].Tag
	}
	return rawTags
}


