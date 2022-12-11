package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/terraform-exec/tfexec"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type GithubCreds struct {
	GITHUB_TOKEN string
}

func getTerraformDir() (execPath string, workingDir string) {
	/*
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("No caller information")
		}
	*/

	workingDir = "/go/terraform"
	execPath = "/usr/bin/terraform"
	/*
		workingDir = path.Dir(filename) + "/terraform"

		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion("1.0.6")),
		}

		execPath, err := installer.Install(context.Background())
		if err != nil {
			log.Fatalf("error installing Terraform: %s", err)
		} else {
			fmt.Printf("Terraform Installation successful (%s)\n", execPath)
		}
	*/

	return execPath, workingDir
}

func createRepo(name string) string {
	githubUser := os.Getenv("GITHUB_USER")
	githubToken := os.Getenv("GITHUB_TOKEN")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubToken), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	execPath, workingDir := getTerraformDir()

	fmt.Printf("Create app: %s in repo %s\n", name, repoName)

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
		tfexec.Var("repo_name="+repoName),
		tfexec.Var("github_user="+githubUser),
		tfexec.Var("github_token="+githubCreds.GITHUB_TOKEN),
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

	return repoName
}

func destroyRepo(name string) {
	githubUser := os.Getenv("GITHUB_USER")
	githubToken := os.Getenv("GITHUB_TOKEN")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubToken), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	execPath, workingDir := getTerraformDir()

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	fmt.Println("Destroying changes: " + repoName)
	err = tf.Destroy(context.Background(),
		tfexec.Var("repo_name="+repoName),
		tfexec.Var("github_user="+githubUser),
		tfexec.Var("github_token="+githubCreds.GITHUB_TOKEN),
	)
	if err != nil {
		log.Fatalf("error running destroy: %s", err)
	}
}

func CreateRepo(appName string, skeletonRepo string, skeletonRepoPath string) string {
	githubUser := common.GetConfig("GITHUB_USER")
	githubEmail := common.GetConfig("GITHUB_EMAIL")
	githubToken := common.GetConfig("GITHUB_TOKEN")
	githubBase := common.GetConfig("GITHUB_BASE")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubToken), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoName := createRepo(appName)
	repoUrl := githubBase + "/" + repoName

	skeletonDir, err := ioutil.TempDir("", "skeleton")
	common.CheckIfError(err)
	defer os.RemoveAll(skeletonDir)

	repoDir, err := ioutil.TempDir("", "repo")
	common.CheckIfError(err)
	defer os.RemoveAll(repoDir)

	_, err = git.PlainClone(skeletonDir, false, &git.CloneOptions{
		URL:      skeletonRepo,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	//repoName := strings.ReplaceAll(strings.ToLower(appName), " ", "-")

	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	r, err := git.PlainOpen(repoDir)
	common.CheckIfError(err)

	w, err := r.Worktree()
	common.CheckIfError(err)

	err = common.Dir(skeletonDir+skeletonRepoPath, repoDir)
	common.CheckIfError(err)

	curDir, err := os.Getwd()
	common.CheckIfError(err)

	err = os.Chdir(repoDir)
	common.CheckIfError(err)

	err = filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasPrefix(path, ".git") {
				fmt.Println("Adding ", path)
				_, err = w.Add(path)
				common.CheckIfError(err)
			}

			return nil
		})

	commit, err := w.Commit("Initial Commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  githubUser,
			Email: githubEmail,
			When:  time.Now(),
		},
	})
	common.CheckIfError(err)
	obj, err := r.CommitObject(commit)
	common.CheckIfError(err)

	fmt.Println(obj)

	err = r.Push(&git.PushOptions{Auth: &http2.BasicAuth{Username: githubUser, Password: githubCreds.GITHUB_TOKEN}})
	common.CheckIfError(err)

	os.Chdir(curDir)

	return repoUrl
}

func DestroyRepo(appName string) {
	destroyRepo(appName)
}
