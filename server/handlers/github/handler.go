package handlers

import (
	"context"
	"fmt"
	"github.com/bones/server/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func getTerraformDir() (execPath string, workingDir string) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}

	workingDir = path.Dir(filename) + "/terraform"

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.0.6")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		log.Fatalf("error installing Terraform: %s", err)
	}

	return execPath, workingDir
}

func createRepo(name string) {
	githubUser := os.Getenv("GITHUB_USER")
	githubToken := os.Getenv("GITHUB_TOKEN")
	repoName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	execPath, workingDir := getTerraformDir()

	fmt.Printf("Create app: %s in repo %s\n", name, repoName)

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background())
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	pass, err := tf.Plan(context.Background(),
		tfexec.Out(workingDir+"/out.plan"),
		tfexec.Var("repo_name="+repoName),
		tfexec.Var("github_user="+githubUser),
		tfexec.Var("github_token="+githubToken),
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
}

func destroyRepo(name string) {
	githubUser := os.Getenv("GITHUB_USER")
	githubToken := os.Getenv("GITHUB_TOKEN")
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
		tfexec.Var("github_token="+githubToken),
	)
	if err != nil {
		log.Fatalf("error running destroy: %s", err)
	}
}

func CreateRepo(appName string, repoFlag string) {
	githubUser := common.GetConfig("GITHUB_USER")
	githubEmail := common.GetConfig("GITHUB_EMAIL")
	githubToken := common.GetConfig("GITHUB_TOKEN")
	githubBase := common.GetConfig("GITHUB_BASE")

	createRepo(appName)
	skeletonDir, err := ioutil.TempDir("", "skeleton")
	common.CheckIfError(err)
	defer os.RemoveAll(skeletonDir)

	repoDir, err := ioutil.TempDir("", "repo")
	common.CheckIfError(err)
	defer os.RemoveAll(repoDir)

	_, err = git.PlainClone(skeletonDir, false, &git.CloneOptions{
		URL:      repoFlag,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubToken,
		},
	})
	common.CheckIfError(err)

	repoName := strings.ReplaceAll(strings.ToLower(appName), " ", "-")

	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      githubBase + "/" + repoName,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubUser,
			Password: githubToken,
		},
	})
	common.CheckIfError(err)

	r, err := git.PlainOpen(repoDir)
	common.CheckIfError(err)

	w, err := r.Worktree()
	common.CheckIfError(err)

	err = common.Dir(skeletonDir+"/go-app", repoDir)
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

	err = r.Push(&git.PushOptions{Auth: &http2.BasicAuth{Username: githubUser, Password: githubToken}})
	common.CheckIfError(err)

	os.Chdir(curDir)
}

func DestroyRepo(appName string) {
	destroyRepo(appName)
}
