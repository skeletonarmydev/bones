package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	github "github.com/bones/server/handlers/github"
	"log"
	"os"
	"strings"
	"text/template"
)

type CircleCICreds struct {
	TOKEN string
}

func CreateProject(name string, repo string, skeletonRepo string, skeletonRepoPath string, data map[string]string) error {

	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds github.GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	fmt.Printf("Creating CircleCI project for app: %s\n", name)

	skeletonDir := github.DownloadRepo(skeletonRepo)
	defer os.RemoveAll(skeletonDir)

	workingDir := skeletonDir + skeletonRepoPath + "/infra/circleci"

	circleCICredsEnv := common.GetConfig("CIRCLECI")
	githubUser := githubCreds.GITHUB_USER

	var circleCreds CircleCICreds
	err = json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		fmt.Printf("Can't parse circleCreds: %s", err)
		return err
	}

	vars := make(map[string]string)
	vars["project_name"] = data["APP_NAME"]
	vars["github_user"] = githubUser
	vars["circleci_token"] = circleCreds.TOKEN

	err = common.ExecuteTerraform(workingDir, vars, common.ApplyAction, data["APP_NAME"]+"/infra/circleci")
	common.CheckIfError(err)

	//Process template
	tmpl, err := template.ParseFiles(workingDir + "/config.yml")
	common.CheckIfError(err)
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	common.CheckIfError(err)

	var files = []github.RemoteFile{
		github.RemoteFile{
			Name: "config.yml",
			Path: ".circleci",
			Data: buf.Bytes(),
			Perm: 0750,
		},
	}
	github.AddFilesToRepo(repo, "Adding CircleCI Config", files)

	fmt.Printf("Finished creating CircleCI project for app: %s\n", name)

	return err
}

func DestroyProject(name string, repo string) error {

	projectDir := github.DownloadRepo(repo)
	defer os.RemoveAll(projectDir)

	workingDir := projectDir + "/infra/circleci"

	circleCICredsEnv := common.GetConfig("CIRCLECI")

	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds github.GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}
	githubUser := githubCreds.GITHUB_USER

	var circleCreds CircleCICreds
	err = json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		fmt.Printf("Can't parse circleCreds: %s", err)
		return err
	}

	projectName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	vars := make(map[string]string)
	vars["project_name"] = projectName
	vars["github_user"] = githubUser
	vars["circleci_token"] = circleCreds.TOKEN

	err = common.ExecuteTerraform(workingDir, vars, common.DestroyAction, projectName+"/infra/circleci")
	return err
}
