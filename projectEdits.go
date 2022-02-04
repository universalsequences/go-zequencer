package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type ProjectEditsQuery struct {
	Project string `json:"project"`
}

func HandleProjectEdits(
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

	query := ProjectEditsQuery{}
	json.Unmarshal([]byte(bodyString), &query)
	
	bytes, err :=json.Marshal(runProjectEdits(caches, query))
	w.Write(bytes)
}

func runProjectEdits(caches *Caches, query ProjectEditsQuery) []Project {
	projects := []Project{}

	project := getProject(caches, query.Project)
	if (project.PreviousSequence == "") {
		return projects
	}
	project = getProject(caches, project.PreviousSequence)


	for project.PreviousSequence != "" {
		projects = append(
			projects,
			project)

		project = getProject(caches, project.PreviousSequence)
	}

	if (project.PreviousSequence == "") {
		projects = append(projects, project)
	}
	return projects
}

func getProject(caches *Caches, projectId string) Project {
	queryBuilder := NewQuery(TOKENIZED_SEQUENCES)
	queryBuilder.From(SequenceEdited)
	queryBuilder.Select("previousSequence")
	queryBuilder.Select("newSequence")
	queryBuilder.Select("title")
	queryBuilder.Select("user")
	queryBuilder.WhereIs("newSequence", projectId)

	projects := convertToProjects(queryBuilder.ExecuteQuery(caches))
	if (len(projects) > 0) {
		return projects[0]
	}
	return Project{}
}


