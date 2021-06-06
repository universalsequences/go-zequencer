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
	Ids []string `json:"ids"`
	User string `json:"user"`
}

type SampleQueryResults struct {
	Id string `json:"id"`
	Title string `json:"title"`
	Tags []string `json:"tags"`
	VideoId string `json:"videoId"`
	CoverArtHash string `json:"coverArtHash"`
	DiscogsId float64 `json:"discogsId"`
	GuildId float64 `json:"guildId"`
	Year float64 `json:"year"`
	User string `json:"user"`
	BlockNumber float64 `json:"blockNumber"`
	Packs []string `json:"packs"`
	Rating int `json:"rating"`
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
	if (len(query.Ids) > 0) {
		data := [] SampleQueryResults{}
		for _, id := range query.Ids {
			data = append(
				data, GetSampleInformation(caches, id, query.User))
		}
		bytes, _ := json.Marshal(data)
		w.Write(bytes)
	} else {
		sampleData := GetSampleInformation(caches, query.Id, query.User)
		bytes, _ := json.Marshal(sampleData)
		w.Write(bytes)
	}
}

func GetSampleInformation(caches *Caches, id string, user string) SampleQueryResults {
	whereClause := WhereClause{Name: "ipfsHash", Value: id};
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
		SelectStatements: []string{"coverArtHash", "discogsId"},
		WhereClauses: []WhereClause{WhereClause{Name: "sampleHash", Value: id}},
		LimitSize: 1}

	titleQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "SampleCreated(address,bytes32,string,uint32)",
		SelectStatements: []string{"title", "guildId", "user"},
		WhereClauses: []WhereClause{whereClause},
		LimitSize: 1}

	yearQuery := Query{
		Address: GUILD_SAMPLES,
		EventLog: "SampleYear(bytes32,int16,uint32)",
		SelectStatements: []string{"year"},
		WhereClauses: []WhereClause{whereClause},
		LimitSize: 1}

	sampleData := SampleQueryResults{}

	tagResults := tagsQuery.ExecuteQuery(caches)
	titleResults := titleQuery.ExecuteQuery(caches)
	coverArtResults := coverArtQuery.ExecuteQuery(caches)
	videoResults := videoQuery.ExecuteQuery(caches)
	yearResults := yearQuery.ExecuteQuery(caches)
	packResults := getPacksWithSound(caches, id)

	if (len(tagResults) >= 1) {
		for _, s := range tagResults {
			sampleData.Tags = append(sampleData.Tags,  s["tag"].(string));
		}
	}

	sampleData.Id = id

	if (len(titleResults) >= 1) {
		sampleData.Title = titleResults[0]["title"].(string);
		sampleData.User = titleResults[0]["user"].(string);
		sampleData.BlockNumber = titleResults[0]["blockNumber"].(float64);
		sampleData.Packs = packResults

		if guildId, ok := titleResults[0]["guildId"].(float64); ok {
			sampleData.GuildId = guildId
		} else {
			sampleData.GuildId = 0
		}
	}

	if (len(videoResults) >= 1) {
		sampleData.VideoId = videoResults[0]["videoId"].(string);
	}

	if (len(yearResults) >= 1) {
		sampleData.Year = yearResults[0]["year"].(float64);
	}

	if (len(coverArtResults) >= 1) {
		if (coverArtResults[0]["coverArtHash"] != nil) {
			sampleData.CoverArtHash = coverArtResults[0]["coverArtHash"].(string);
		}
		if (coverArtResults[0]["discogsId"] != nil) {
			sampleData.DiscogsId = coverArtResults[0]["discogsId"].(float64);
		}

		sampleData.Rating = getRatingForSound(caches, user, id)
	}

	return sampleData
}

	
