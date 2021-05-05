package main
import (
	"fmt")

const PACKS_CONTRACT = "0xF7bd2ada59c4ab5AD0f6BFbE94EB4f8eCa18eEDd";

func getSoundsWithPack(caches *Caches, pack string, guildIds []interface{}) []interface{} {
	query := NewQuery(PACKS_CONTRACT)
	query.From(PackHasContent)
	query.Select("contentHash")
	query.WhereIs("packHash", pack)
	//query.WhereIn("guildId", guildIds)
	results := queryForCache((*caches)[PACKS_CONTRACT], query)

	soundIds := [] interface {}{}
	for _, result := range results {
		soundIds = append(
			soundIds,
			result["contentHash"])
	}

	fmt.Printf("Finding results for pack=%v results=%v\n",
		pack, soundIds)
	return soundIds 
}
