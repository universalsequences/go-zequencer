package main

func pruneContent(results []map[string]interface{}, key string) []map[string]interface{} {
	contentsFound := map[string]bool{}
	pruned := []map[string]interface{}{}
	for _, row := range results {
		if _, ok := contentsFound[row[key].(string)]; !ok {
			pruned = append(pruned, row)
			contentsFound[row[key].(string)] = true
		}
	}
	return pruned
}
