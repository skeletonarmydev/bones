package main

import (
	"encoding/json"
	"flag"
	"fmt"
	circleci "github.com/bones/server/handlers/circleci"
	github "github.com/bones/server/handlers/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
)

var banner = `
 ____                        
|  _ \                       
| |_) | ___  _ __   ___  ___ 
|  _ < / _ \| '_ \ / _ \/ __|
| |_) | (_) | | | |  __/\__ \
|____/ \___/|_| |_|\___||___/
                             
Your skeleton army scaffolding service
`

type Project struct {
	Id   string `json:"Id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Desc string `json:"desc"`
	Repo string `json:"repo"`
}

type ProjectType struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Repo string `json:"repo"`
	Path string `json:"path"`
}

type ProjectCreateRequest struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

type ProjectDeleteRequest struct {
	Id string `json:"Id"`
}

type ProjectTypeCreateRequest struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	Repo string `json:"repo"`
	Path string `json:"path"`
}

type ProjectTypeDeleteRequest struct {
	Slug string `json:"slug"`
}

//Globals
var Projects = make(map[string]Project)
var ProjectTypes = make(map[string]ProjectType)

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

	projectType, ok := ProjectTypes[projectRequest.Type]
	if !ok {
		http.Error(w, "Project Type Not Found", http.StatusNotFound)
		return
	}

	var project Project

	id := uuid.New()
	project.Id = id.String()
	project.Name = projectRequest.Name
	project.Type = projectRequest.Type
	project.Desc = projectRequest.Desc

	go func() {
		repoUrl := github.CreateRepo(project.Name, projectType.Repo, projectType.Path)

		project.Repo = repoUrl
		Projects[id.String()] = project
	}()

	go circleci.CreateProject(project.Name, projectType.Repo, projectType.Path)

	Projects[id.String()] = project

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {

	var projectRequest ProjectDeleteRequest
	err := json.NewDecoder(r.Body).Decode(&projectRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, ok := Projects[projectRequest.Id]
	if !ok {
		http.Error(w, "Project Not Found", http.StatusNotFound)
		return
	}

	projectType, ok := ProjectTypes[project.Type]
	if !ok {
		http.Error(w, "Project Type Not Found", http.StatusNotFound)
		return
	}

	go github.DestroyRepo(project.Name)
	go circleci.DestroyProject(project.Name, projectType.Repo, projectType.Path)

	delete(Projects, projectRequest.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func returnAllProjectTypes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllProjectTypes")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectTypes)
}

func returnHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnHealth")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Alive")
}

func createNewProjectType(w http.ResponseWriter, r *http.Request) {
	var projectType ProjectType

	var projectTypeRequest ProjectTypeCreateRequest
	err := json.NewDecoder(r.Body).Decode(&projectTypeRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	projectType.Name = projectTypeRequest.Name
	projectType.Desc = projectTypeRequest.Desc
	projectType.Repo = projectTypeRequest.Repo
	projectType.Path = projectTypeRequest.Path
	projectType.Slug = strings.ReplaceAll(strings.ToLower(projectType.Name), " ", "-")

	ProjectTypes[projectType.Slug] = projectType

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectType)
}

func deleteProjectType(w http.ResponseWriter, r *http.Request) {

	var projectTypeRequest ProjectTypeDeleteRequest
	err := json.NewDecoder(r.Body).Decode(&projectTypeRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	projectType, ok := ProjectTypes[projectTypeRequest.Slug]
	if !ok {
		http.Error(w, "Project Type Not Found", http.StatusNotFound)
		return
	}

	delete(ProjectTypes, projectTypeRequest.Slug)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projectType)
}

func handleRequests() {

	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/", returnHealth)
	myRouter.HandleFunc("/project", returnAllProjects).Methods("GET")
	myRouter.HandleFunc("/project", createNewProject).Methods("POST")
	myRouter.HandleFunc("/project", deleteProject).Methods("DELETE")

	myRouter.HandleFunc("/type", returnAllProjectTypes).Methods("GET")
	myRouter.HandleFunc("/type", createNewProjectType).Methods("POST")
	myRouter.HandleFunc("/type", deleteProjectType).Methods("DELETE")

	fmt.Println("Now online and ready")
	log.Fatal(http.ListenAndServe(":8080", myRouter))

}

func main() {
	fmt.Printf(banner)

	var help = flag.Bool("help", false, "Show help")

	// Parse the flag
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	handleRequests()
}
