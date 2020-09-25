package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
)

const TOKENIZED_SEQUENCES = "0x606f760c228cd5f11c6f79de64d3b299b11f1ed1"

type ProjectsQuery struct {
	SearchTerm string `json:"searchTerm"`
	User       string `json:"user"`
}

type Project struct {
	NewSequence string `json:"newSequence"`
	PreviousSequence string `json:"previousSequence"`
	Title string `json:"title"`
	User string `json:"user"`
	Edits int `json:"edits"`
	BlockNumber float64 `json:"blockNumber"`
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
	fmt.Println("runProjectsQuery")

	queryBuilder := NewQuery(TOKENIZED_SEQUENCES)
	queryBuilder.From(SequenceEdited)
	queryBuilder.Select("previousSequence")
	queryBuilder.Select("newSequence")
	queryBuilder.Select("title")
	queryBuilder.Select("user")

	if (query.User != "") {
		fmt.Printf("Filtering by user=%s\n", query.User)
		queryBuilder.WhereIs("user", query.User)
	}

	results := queryBuilder.ExecuteQuery(caches)
	return collapseProjects(convertToProjects(results))
}

func convertToProjects(results []map[string]interface{}) []Project {
	projects := []Project{}
	for _, result := range results {
		previousSequence, ok := result["previousSequence"].(string)
		if (!ok) {
			previousSequence = ""
		}


		projects = append(
			projects,
			Project{
				User: result["user"].(string),
				NewSequence: result["newSequence"].(string),
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
			edits := len(getAllEdits(project, idToProject))
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
		if previousProject == project {
			return []Project{project}
		}
		
		return append(
			getAllEdits(previousProject, idToProject),
			project)
	}
	
	// no previous projects so just return singleton list
	return []Project{project}
}
