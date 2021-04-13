package main

const GUILD_SEQUENCES = "0x15Da2ef01D9F881f694Be962Db28b76110fa195C"

func getGuildSequences(caches *Caches, guildIds []float64, filterMine bool, user string) []Project {
	results := []Project{}
	guilds := []interface{}{}
	for _, guildId := range guildIds {
		guilds = append(guilds, guildId)
	}
	queryBuilder := NewQuery(GUILD_SEQUENCES)
	queryBuilder.From(NewGuildSequence)
	queryBuilder.Select("contentHash")
	queryBuilder.Select("previousHash")
	queryBuilder.Select("encryptedName")
	queryBuilder.Select("encryptedContentKey")
	queryBuilder.Select("publicKey")
	queryBuilder.Select("guildId")
	queryBuilder.WhereIn("guildId", guilds)

	if (filterMine || user != "") {
		queryBuilder.WhereIs("user", user)
	}
	
	guildResults := queryBuilder.ExecuteQuery(caches)
	for _, result := range guildResults {
		project := Project{
			EncryptedName: result["encryptedName"].(string),
			GuildId: result["guildId"].(float64),
			NewSequence: result["contentHash"].(string),
			User: result["user"].(string),
			EncryptedContentKey: result["encryptedContentKey"].(string),
			PublicKey: result["publicKey"].(string),
			BlockNumber: result["blockNumber"].(float64),
		}

		if previousHash, ok := result["previousHash"].(string); ok {
			project.PreviousSequence = previousHash
		}
		results = append(results, project)
	}
	return results
}

