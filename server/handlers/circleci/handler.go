package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	github "github.com/bones/server/handlers/github"
	"os"
	"strings"
)

type CircleCICreds struct {
	TOKEN string
}

type GithubCreds struct {
	GITHUB_TOKEN string
}

func CreateProject(name string, skeletonRepo string, skeletonRepoPath string) error {

	skeletonDir := github.DownloadRepo(skeletonRepo)
	defer os.RemoveAll(skeletonDir)

	workingDir := skeletonDir + skeletonRepoPath + "/infra/circleci"

	circleCICredsEnv := common.GetConfig("CIRCLECI")
	githubUser := "ascii27"

	var circleCreds CircleCICreds
	err := json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		fmt.Printf("Can't parse circleCreds: %s", err)
		return err
	}

	projectName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	vars := make(map[string]string)
	vars["project_name"] = projectName
	vars["github_user"] = githubUser
	vars["circleci_token"] = circleCreds.TOKEN

	err = common.ExecuteTerraform(workingDir, vars, common.ApplyAction)
	return err
}

func DestroyProject(name string, skeletonRepo string, skeletonRepoPath string) error {

	skeletonDir := github.DownloadRepo(skeletonRepo)
	defer os.RemoveAll(skeletonDir)

	workingDir := skeletonDir + skeletonRepoPath + "/infra/circleci"

	circleCICredsEnv := common.GetConfig("CIRCLECI")
	githubUser := "ascii27"

	var circleCreds CircleCICreds
	err := json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		fmt.Printf("Can't parse circleCreds: %s", err)
		return err
	}

	projectName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	vars := make(map[string]string)
	vars["project_name"] = projectName
	vars["github_user"] = githubUser
	vars["circleci_token"] = circleCreds.TOKEN

	err = common.ExecuteTerraform(workingDir, vars, common.DestroyAction)
	return err
}
