package main

import (
	"strings"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

const OLD_TOKENIZED_SEQUENCES = "0x606f760c228cd5f11c6f79de64d3b299b11f1ed1"
const TOKENIZED_SEQUENCES = "0x27Fd050dF7c0c603A407f1BC4fd0ED634824270E";

type ProjectsQuery struct {
	SearchTerm string `json:"searchTerm"`
	User       string `json:"user"`
	Starred bool `json:"starred"`
	Favorited string `json:"favorited"`
	FilterMine bool `json:"filterMine"`
	SearchTag string`json:"searchTag"`
	GuildIds []float64 `json:"guildIds"`
	Old bool `json:"old"`
}

type Project struct {
	NewSequence string `json:"newSequence"`
	PreviousSequence string `json:"previousSequence"`
	Title string `json:"title"`
	User string `json:"user"`
	Edits int `json:"edits"`
	BlockNumber float64 `json:"blockNumber"`
	PublicKey string `json:"publicKey"`
	EncryptedName string `json:"encryptedName"`
	EncryptedContentKey string `json:"encryptedContentKey"`
	GuildId float64 `json:"guildId"`
	Collaborators []string `json:"collaborators"`
}


func HandleProjectsQuery(
	w http.ResponseWriter,
	r *http.Request,
	caches *Caches,
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

	query := ProjectsQuery{}
	json.Unmarshal([]byte(bodyString), &query)

	bytes, err :=json.Marshal(runProjectsQuery(caches, query))
	cachedQueries.newQuery(bodyString, bytes)
	w.Write(bytes)
}

func runProjectsQuery(caches *Caches, query ProjectsQuery) []Project {
	contract := TOKENIZED_SEQUENCES;
	if (query.Old) {
		contract = OLD_TOKENIZED_SEQUENCES;
	}
	queryBuilder := NewQuery(contract)
	if (query.Old) {
		queryBuilder.From(SequenceEditedOld)
	} else {
		queryBuilder.From(SequenceEdited)
	}
	queryBuilder.Select("previousSequence")
	queryBuilder.Select("newSequence")
	queryBuilder.Select("title")
	queryBuilder.Select("user")

	if (query.FilterMine || query.User != "") {
		queryBuilder.WhereIs("user", query.User)
	}
	results := []map[string]interface{}{}
	for _, guildId := range query.GuildIds {
		if (guildId == 0) {
			results = pruneContent(queryBuilder.ExecuteQuery(caches), "newSequence")
			break
		}
	}
	projectResults := convertToProjects(results)
	guildResults := getGuildSequences(caches, query.GuildIds, query.FilterMine, query.User)
	for _, result := range guildResults {
		projectResults = append(projectResults, result)
	}

	filtered := []Project{}
	starred := GetStarredProjects(caches)
	favorited := GetFavoritedProjects(caches, query.Favorited)
	projectTags := GetProjectTags(caches)

	for _, result := range projectResults {
		id := result.NewSequence
		if (query.Favorited != "" && query.Starred) {
			if _, ok := starred[id]; !ok {
				if _, ok2 := favorited[id]; !ok2 {
					continue
				}
			}
		} else if (query.Favorited != "") {
				if _, ok := favorited[id]; !ok {
					continue
				}
		} else if (query.Starred) {
				if _, ok := starred[id]; !ok {
					continue
				}
		}

		matchedTag := false
		if (query.SearchTag != "") {
			if tags, ok := projectTags[id]; ok {
				for _, tag := range tags {
					if (strings.Contains(strings.ToLower(tag), strings.ToLower(query.SearchTag))) {
						matchedTag = true
					}
				}
			}
			if (query.SearchTerm == "" && !matchedTag) {
				continue
			}
		} else {
			matchedTag = true
		}

		if (query.SearchTerm != "") {
			matchedTag = false
			if tags, ok := projectTags[id]; ok {
				for _, tag := range tags {
					if (strings.Contains(strings.ToLower(tag), strings.ToLower(query.SearchTerm))) {
						matchedTag = true
					}
				}
			}

			if (!matchedTag && !strings.Contains(strings.ToLower(result.Title), strings.ToLower(query.SearchTerm))) {
				continue;
			}
		}

		filtered = append(
			filtered,
			result)
	}

	return collapseProjects(filtered) //collapseProjects(convertToProjects(filtered))
}

func convertToProjects(
	results []map[string]interface{}) []Project {
	projects := []Project{}
	for _, result := range results {
		previousSequence, ok := result["previousSequence"].(string)
		if (!ok) {
			previousSequence = ""
		}

		id := result["newSequence"].(string)

		projects = append(
			projects,
			Project{
				User: result["user"].(string),
				NewSequence: id,
				PreviousSequence: previousSequence,
				Title: result["title"].(string),
				BlockNumber: result["blockNumber"].(float64),
			})
	}
	return projects
}

func collapseProjects(projects []Project) []Project {
	idToProject := map[string]Project{}
	projectToNext := map[string]string{}

	for _, project := range projects {
		idToProject[project.NewSequence] = project
	}

	for _, project := range projects {
		if _, ok := idToProject[project.PreviousSequence]; ok {
			// this project has a previous sequence so map
			// previous -> next
			projectToNext[project.PreviousSequence] = project.NewSequence
		}
	}

	collapsed := []Project{}
	for _, project := range projects {
		if _, ok := projectToNext[project.NewSequence]; !ok {
			// has no next so its the top
			allEdits := getAllEdits(project, idToProject)
			edits := len(allEdits)
			project.Collaborators = getCollaborators(allEdits)
			project.Edits = edits - 1
			collapsed = append(
				collapsed,
				project)
		}
	}
	return collapsed
}

func getAllEdits(project Project, idToProject map[string]Project) []Project {
	if previousProject, ok := idToProject[project.PreviousSequence]; ok {
		// then we have a previous project that exists
		if ok && previousProject.NewSequence == project.NewSequence {
			return []Project{project}
		}
		
		return append(
			getAllEdits(previousProject, idToProject),
			project)
	}
	
	// no previous projects so just return singleton list
	return []Project{project}
}

func getCollaborators(projects []Project) []string {
	collabMap := map[string]bool{}

	for _, project := range projects {
		collabMap[project.User] = true
	}

	collaborators := []string{}
	for collaborator, _ := range collabMap {
		collaborators = append(
			collaborators,
			collaborator)
	}
	return collaborators
}
