package main

import (
	"encoding/json"
	"flag"
	"fmt"
	aws "github.com/bones/server/handlers/aws"
	circleci "github.com/bones/server/handlers/circleci"
	github "github.com/bones/server/handlers/github"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
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
	Id   string            `json:"Id"`
	Name string            `json:"name"`
	Type string            `json:"type"`
	Desc string            `json:"desc"`
	Repo string            `json:"repo"`
	Data map[string]string `json:"data""`
}

type ProjectType struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Repo string `json:"repo"`
	Path string `json:"path"`
}

type ProjectCreateRequest struct {
	Type string            `json:"type"`
	Name string            `json:"name"`
	Desc string            `json:"desc"`
	Data map[string]string `json:"data""`
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

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.

type GenerateStep struct {
	Name    string
	Handler string
	Path    string
	Cmd     string
}

type DestroyStep struct {
	Name    string
	Handler string
	Path    string
	Cmd     string
}

type SkeletonYaml struct {
	Generate struct {
		Steps []GenerateStep
	}
	Destroy struct {
		Steps []DestroyStep
	}
}

// Globals
var Projects = make(map[string]*Project)
var ProjectTypes = make(map[string]ProjectType)
var SkeletonYAML = SkeletonYaml{}

func returnAllProjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllProjects")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Projects)
}

func processGenerateSteps(step GenerateStep, project *Project, projectType ProjectType) error {
	fmt.Printf("Running step: %s\n", step.Name)

	switch step.Handler {
	case "github":
		project.Repo = github.CreateRepo(project.Name, projectType.Repo, projectType.Path, project.Data)
		return nil
	case "aws":
		return aws.CreateAWSInfra(project.Name, project.Repo, projectType.Repo, projectType.Path, project.Data)
	case "circleci":
		return circleci.CreateProject(project.Name, project.Repo, projectType.Repo, projectType.Path, project.Data)
	}

	return nil
}

func processDestroySteps(step DestroyStep, project *Project) error {
	fmt.Printf("Running step: %s\n", step.Name)

	switch step.Handler {
	case "github":
		github.DestroyRepo(project.Repo)
		return nil
	case "aws":
		return aws.DestroyAWSInfra(project.Name, project.Repo)
	case "circleci":
		return circleci.DestroyProject(project.Name, project.Repo)
	}

	return nil
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
	project.Data = projectRequest.Data

	go func() {

		skeletonDir := github.DownloadRepo(projectType.Repo)
		defer os.RemoveAll(skeletonDir)

		//get bones manifest
		skeletonyaml, err := os.ReadFile(skeletonDir + projectType.Path + "/.skeleton/skeleton.yaml")
		if err != nil {
			http.Error(w, "Project configuration not found (skeleton.yaml missing!)", http.StatusFailedDependency)
			log.Print(err)
			return
		}

		err = yaml.Unmarshal(skeletonyaml, &SkeletonYAML)
		if err != nil {
			http.Error(w, "Project configuration not formatted correctly (skeleton.yaml corrupted!)", http.StatusBadRequest)
			log.Print(err)
			return
		}

		//Setting standard values
		slug := strings.ReplaceAll(strings.ToLower(project.Name), " ", "-")
		project.Data["APP_NAME"] = slug
		project.Data["SERVICE_NAME"] = slug + "-service"

		for _, s := range SkeletonYAML.Generate.Steps {
			processGenerateSteps(s, &project, projectType)
		}

	}()

	Projects[id.String()] = &project

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

	go func() {
		projectDir := github.DownloadRepo(project.Repo)
		defer os.RemoveAll(projectDir)

		//get bones manifest
		skeletonyaml, err := os.ReadFile(projectDir + "/.skeleton/skeleton.yaml")
		if err != nil {
			http.Error(w, "Project configuration not found (skeleton.yaml missing!)", http.StatusFailedDependency)
			log.Print(err)
			return
		}

		err = yaml.Unmarshal(skeletonyaml, &SkeletonYAML)
		if err != nil {
			http.Error(w, "Project configuration not formatted correctly (skeleton.yaml corrupted!)", http.StatusBadRequest)
			log.Print(err)
			return
		}

		for _, s := range SkeletonYAML.Destroy.Steps {
			processDestroySteps(s, project)
		}

	}()

	delete(Projects, projectRequest.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func returnAllProjectTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ProjectTypes)
}

func returnHealth(w http.ResponseWriter, r *http.Request) {
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
