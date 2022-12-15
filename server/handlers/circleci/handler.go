package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	"github.com/go-git/go-git/v5"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-exec/tfexec"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type CircleCICreds struct {
	token string
}

type GithubCreds struct {
	GITHUB_TOKEN string
}

func getTerraformDir() (execPath string) {
	execPath = "/usr/bin/terraform"

	return execPath
}

func createProject(name string, workingDir string) string {
	circleCICredsEnv := os.Getenv("CIRCLECI")
	githubUser := "ascii27"

	var circleCreds CircleCICreds
	err := json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		log.Fatalf("Can't parse circleCreds: %s", err)
	}

	projectName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	execPath := getTerraformDir()

	fmt.Printf("Create circleci project %s for app: %s\n", projectName, name)

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s (execPath: %s)", err, execPath)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	pass, err := tf.Plan(context.Background(),
		tfexec.Out(workingDir+"/out.plan"),
		tfexec.Var("project_name="+projectName),
		tfexec.Var("github_user="+githubUser),
		tfexec.Var("circleci_token="+circleCreds.token),
	)
	if err != nil {
		log.Fatalf("error running Plan: %s", err)
	}

	if pass {
		plan, err := tf.ShowPlanFile(context.Background(), workingDir+"/out.plan")
		if err != nil {
			log.Fatalf("error running fetch plan: %s", err)
		}

		for _, s := range plan.ResourceChanges {
			fmt.Printf("Change: %s %s\n", s.Change.Actions, s.Name)
		}

		fmt.Println("Applying changes")
		err2 := tf.Apply(context.Background(), tfexec.DirOrPlan(workingDir+"/out.plan"))

		if err2 != nil {
			log.Fatalf("error running apply: %s", err2)
		}

		os.Remove(workingDir + "/out.plan")

	}

	return projectName
}

func destroyProject(name string, workingDir string) {
	circleCICredsEnv := os.Getenv("CIRCLECI")
	githubUser := "ascii27"

	var circleCreds CircleCICreds
	err := json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		log.Fatalf("Can't parse circleCreds: %s", err)
	}

	projectName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	execPath := getTerraformDir()

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	fmt.Println("Destroying changes: " + projectName)
	err = tf.Destroy(context.Background(),
		tfexec.Var("project_name="+projectName),
		tfexec.Var("github_user="+githubUser),
		tfexec.Var("circleci_token="+circleCreds.token),
	)
	if err != nil {
		log.Fatalf("error running destroy: %s", err)
	}
}

func CreateProject(appName string, skeletonRepo string, skeletonRepoPath string) string {
	githubUser := common.GetConfig("GITHUB_USER")
	githubToken := common.GetConfig("GITHUB_TOKEN")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubToken), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	circleCICredsEnv := os.Getenv("CIRCLECI")

	var circleCreds CircleCICreds
	err = json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		log.Fatalf("Can't parse circleCreds: %s", err)
	}

	skeletonDir, err := ioutil.TempDir("", "skeleton")
	common.CheckIfError(err)
	defer os.RemoveAll(skeletonDir)

	_, err = git.PlainClone(skeletonDir, false, &git.CloneOptions{
		URL:      skeletonRepo,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	projectName := createProject(appName, skeletonDir+skeletonRepoPath+"/infra/circleci")

	return projectName
}

func DestroyProject(appName string, skeletonRepo string, skeletonRepoPath string) {
	githubUser := common.GetConfig("GITHUB_USER")
	githubToken := common.GetConfig("GITHUB_TOKEN")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubToken), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	circleCICredsEnv := os.Getenv("CIRCLECI")

	var circleCreds CircleCICreds
	err = json.Unmarshal([]byte(circleCICredsEnv), &circleCreds)
	if err != nil {
		log.Fatalf("Can't parse circleCreds: %s", err)
	}

	skeletonDir, err := ioutil.TempDir("", "skeleton")
	common.CheckIfError(err)
	defer os.RemoveAll(skeletonDir)

	_, err = git.PlainClone(skeletonDir, false, &git.CloneOptions{
		URL:      skeletonRepo,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	destroyProject(appName, skeletonDir+skeletonRepoPath+"/infra/circleci")
}
