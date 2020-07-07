package main

import (
	"io/ioutil"
	"net/http"
	"log"
	"encoding/json"
)

const GUILD_SAMPLES = "0xc77d4e72dF7D0Bf96488eF543253af537fEb8737";

type SampleQuery struct {
	Id string `json:"id"`
}

type SampleQueryResults struct {
	Title string `json:"title"`
	Tags []string `json:"tags"`
	VideoId string `json:"videoId"`
	CoverArtHash string `json:"coverArtHash"`
	GuildId float64 `json:"guildId"`
}

func HandleSampleQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")

	bodyString := string(bodyBytes)
	query := SampleQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	whereClause := WhereClause{Name: "ipfsHash", Value: query.Id};

	tagsQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "SampleTagged(bytes32,bytes32,uint32)",
		SelectStatements: []string{"tag"},
		WhereClauses: []WhereClause{whereClause}}


	videoQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "SampledYoutube(bytes32,bytes32,uint32)",
		SelectStatements: []string{"videoId"},
		WhereClauses: []WhereClause{whereClause},
		LimitSize: 1}

	coverArtQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "NewDiscogsSample(bytes32,uint256,bytes32,uint32)",
		SelectStatements: []string{"coverArtHash"},
		WhereClauses: []WhereClause{WhereClause{Name: "sampleHash", Value: query.Id}},
		LimitSize: 1}

	titleQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "SampleCreated(address,bytes32,string,uint32)",
		SelectStatements: []string{"title", "guildId"},
		WhereClauses: []WhereClause{whereClause},
		LimitSize: 1}

	sampleData := SampleQueryResults{}

	tagResults := queryForCache((*caches)[GUILD_SAMPLES], tagsQuery)
	titleResults := queryForCache((*caches)[GUILD_SAMPLES], titleQuery)
	coverArtResults := queryForCache((*caches)[GUILD_SAMPLES], coverArtQuery)
	videoResults := queryForCache((*caches)[GUILD_SAMPLES], videoQuery)

	if (len(tagResults) >= 1) {
		for _, s := range tagResults {
			sampleData.Tags = append(sampleData.Tags,  s["tag"].(string));
		}
	}

	if (len(titleResults) >= 1) {
		sampleData.Title = titleResults[0]["title"].(string);
		if guildId, ok := titleResults[0]["guildId"].(float64); ok {
			sampleData.GuildId = guildId
		} else {
			sampleData.GuildId = 0
		}
	}

	if (len(videoResults) >= 1) {
		sampleData.VideoId = videoResults[0]["videoId"].(string);
	}

	if (len(coverArtResults) >= 1) {
		sampleData.CoverArtHash = coverArtResults[0]["coverArtHash"].(string);
	}

	bytes, err := json.Marshal(sampleData)
	w.Write(bytes)
}

	
