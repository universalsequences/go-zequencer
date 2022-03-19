package main

import (
	"sort"
)

func getRecentTags(caches *Caches, recentSounds []SampleResult) map[string][]string {
	ids := []interface{}{}
	for _, sound := range recentSounds {
		ids = append(ids, sound.IpfsHash)
	}

	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
			"tag",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "ipfsHash",
				ValueList: ids,
			}},
		FromBlockNumber: 1,
	};

	results := query.ExecuteQuery(caches)

	tags := map[string][]string{}

	for _, result := range results {
		ipfsHash := result["ipfsHash"].(string)
		if _, ok := tags[ipfsHash]; !ok {
			tags[ipfsHash] = []string{result["tag"].(string)}
		} else {
			if (!containsTag(tags[ipfsHash], result["tag"].(string))) {
				tags[ipfsHash] = append(tags[ipfsHash], result["tag"].(string))
			}
		}
	}
	return tags 
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if (t == tag) {
			return true
		}
	}
	return false
}

type SampleTag struct {
	Tag string
	Count int
}

func getAllTags(caches *Caches) []string {
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
			"tag",
		},
		FromBlockNumber: 1,
	};

	results := query.ExecuteQuery(caches)

	tagsMap := map[string]int{}
	for _, result := range results {
		if _, ok := tagsMap[result["tag"].(string)]; !ok {
			tagsMap[ result["tag"].(string)] = 0
		}
		tagsMap[ result["tag"].(string)]++
	}

	sampleTags:= []SampleTag{}
	for tag, count := range tagsMap  {
		sampleTags = append(sampleTags, SampleTag{
			Tag: tag,
			Count: count,
		})
	}

	sort.Sort(BySampleTag(sampleTags))

	tags := []string{}
	for _, sampleTag := range sampleTags {
		tags = append(
			tags,
			sampleTag.Tag)
	}
	return tags
}

func getSoundsWithOrTags(caches *Caches, tags []interface{}, guildIds []interface{}) []interface{}{
	query := Query{
		Address: GUILD_SAMPLES,
		EventLog: SampleTagged,
		SelectStatements: []string{
			"ipfsHash",
		},
		WhereClauses: []WhereClause{
			WhereClause{
				Name: "tag",
				ValueList: tags,
			},
			WhereClause{
				Name: "guildId",
				ValueList: guildIds,
			}},
		FromBlockNumber: 1,
	};

	results := query.ExecuteQuery(caches)

	ids := []interface{}{}
	for _, result := range results {
		ids = append(ids, result["ipfsHash"])
	}
	return ids
}

type BySampleTag []SampleTag
func (a BySampleTag) Len() int           { return len(a) }
func (a BySampleTag) Less(i, j int) bool { return a[i].Count > a[j].Count }
func (a BySampleTag) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

