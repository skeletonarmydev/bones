package main

import (
	"encoding/json"
	"flag"
	"fmt"
	github "github.com/bones/server/handlers/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var banner = `
 ____                        
|  _ \                       
| |_) | ___  _ __   ___  ___ 
|  _ < / _ \| '_ \ / _ \/ __|
| |_) | (_) | | | |  __/\__ \
|____/ \___/|_| |_|\___||___/
                             
Scaffolding service
`

type Project struct {
	Id   string `json:"Id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
	Repo string `json:"repo"`
}

type ProjectCreateRequest struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type ProjectDeleteRequest struct {
	Id string `json:"Id"`
}

var Projects = make(map[string]Project)

func returnAllProjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllProjects")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Projects)
}

func createNewProject(w http.ResponseWriter, r *http.Request) {

	var projectRequest ProjectCreateRequest
	err := json.NewDecoder(r.Body).Decode(&projectRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var project Project

	id := uuid.New()
	project.Id = id.String()
	project.Name = projectRequest.Name
	project.Type = projectRequest.Type
	project.Desc = projectRequest.Desc

	go func() {
		repoUrl := github.CreateRepo(project.Name, "https://github.com/ascii27/skeletons")

		project.Repo = repoUrl

		Projects[id.String()] = project
	}()

	json.NewEncoder(w).Encode(project)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {

	var projectRequest ProjectDeleteRequest
	err := json.NewDecoder(r.Body).Decode(&projectRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project := Projects[projectRequest.Id]

	go github.DestroyRepo(project.Name)

	delete(Projects, projectRequest.Id)

	json.NewEncoder(w).Encode(project)
}

func handleRequests() {

	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/all", returnAllProjects)
	myRouter.HandleFunc("/project", createNewProject).Methods("POST")
	myRouter.HandleFunc("/project", deleteProject).Methods("DELETE")

	fmt.Println("Now online and ready")
	log.Fatal(http.ListenAndServe(":8080", myRouter))

}

func main() {
	fmt.Printf(banner)

	var help = flag.Bool("help", false, "Show help")
	var createAction = flag.Bool("create", false, "create")
	var destroyAction = flag.Bool("destroy", false, "destroy")
	var repoFlag = ""
	var appName = ""

	flag.StringVar(&repoFlag, "repo", "", "The repo of the scaffold template")
	flag.StringVar(&appName, "name", "", "The name of your new app")

	// Parse the flag
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *createAction {
		github.CreateRepo(appName, repoFlag)
	} else if *destroyAction {
		github.DestroyRepo(appName)
	}

	handleRequests()
}
