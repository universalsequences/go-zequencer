package main

import (
	"strings"
)

type Release struct {
	ReleaseId float64
	ArtistName string
	ReleaseName string
	ReleaseType string
	CoverArtHash string
}
const ARTISTS_CONTRACT = "0x0C8aB6a2ED347F076ed69f9eBf05967a8A889400"

func getRecentReleases(caches *Caches, recentDiscogs []SampleResult) map[string]Release {
	discogsIds := []interface{}{}
	for _, row := range recentDiscogs {
		discogsIds = append(discogsIds, row.DiscogsId)
	}
	releases := getReleases(caches, discogsIds)
	results := map[string]Release{}
	for _, result := range recentDiscogs {
		if release, ok := releases[result.DiscogsId]; ok {
			results[result.IpfsHash] = release
		}
	}

	return results
}

func getReleases(caches *Caches, discogsIds []interface{}) map[float64]Release {
		query := Query{
			Address: ARTISTS_CONTRACT,
			EventLog: ReleaseInfo,
			SelectStatements: []string{
				"artistName",
				"coverArtHash",
				"releaseName",
				"releaseType",
				"releaseId",
			},
			FromBlockNumber: 1,
			WhereClauses: []WhereClause{
				WhereClause{
					Name: "releaseId",
					ValueList: discogsIds,
				}},
		};
	
	cache := (*caches)[ARTISTS_CONTRACT] 
	results := queryForCache(cache, query)
	ret := map[float64]Release{}
	for _, result := range results {
		releaseId := result["releaseId"].(float64)
		release := Release{
			ReleaseType: result["releaseType"].(string),
			ReleaseId: result["releaseId"].(float64),
			CoverArtHash: result["coverArtHash"].(string),
			ReleaseName: strings.TrimPrefix(strings.TrimPrefix(result["releaseName"].(string), "RECORD"), "SAMPLE_PACK"),
			ArtistName: result["artistName"].(string),
		}
		ret[releaseId] = release
	}
	return ret
}

func getRecentArtists(caches *Caches, sounds []SampleResult) map[string]string {
	ids := []interface{}{}
	for _, row := range sounds {
		ids = append(ids, row.IpfsHash)
	}
	query := NewQuery(ARTISTS_CONTRACT)
	query.From(SampleByArtist)
	query.Select("artistName")
	query.Select("ipfsHash")
	query.WhereIn("ipfsHash", ids)
	results := queryForCache((*caches)[ARTISTS_CONTRACT], query)

	artists := map[string]string{}
	for _, result := range results {
		artists[result["ipfsHash"].(string)] = result["artistName"].(string)
	}
	return artists

}

func getSamplesFromRelease(caches *Caches, releaseId float64, guildIds []interface{}) []interface{} {
	query := NewQuery(GUILD_SAMPLES)
	query.From(NewDiscogsSample)
	query.Select("sampleHash")
	query.WhereIs("discogsId", releaseId)
	query.WhereIn("guildId", guildIds)
	results := queryForCache((*caches)[GUILD_SAMPLES], query)

	soundIds := [] interface {}{}
	for _, result := range results {
		soundIds = append(
			soundIds,
			result["sampleHash"])
	}

	return soundIds 
}

