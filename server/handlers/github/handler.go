package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type GithubCreds struct {
	GITHUB_USER  string
	GITHUB_TOKEN string
	GITHUB_EMAIL string
	GITHUB_BASE  string
}

func getWorkingDir() string {

	if common.GetConfig("SA_LOCAL") == "true" {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("No caller information")
		}

		return path.Dir(filename) + "/terraform"
	} else {
		return "/go/terraform"
	}
}

func createRepo(name string) string {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}
	repoName := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	vars := make(map[string]string)
	vars["repo_name"] = repoName
	vars["github_user"] = githubCreds.GITHUB_USER
	vars["github_token"] = githubCreds.GITHUB_TOKEN

	err = common.ExecuteTerraform(getWorkingDir(), vars, common.ApplyAction)
	common.CheckIfError(err)

	return repoName
}

func destroyRepo(name string) {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse github environment: %s", err)
	}

	repoName := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	vars := make(map[string]string)
	vars["repo_name"] = repoName
	vars["github_user"] = githubCreds.GITHUB_USER
	vars["github_token"] = githubCreds.GITHUB_TOKEN

	err = common.ExecuteTerraform(getWorkingDir(), vars, common.DestroyAction)
	common.CheckIfError(err)
}

func DownloadRepo(repo string) string {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse github environment: %s", err)
	}

	tempDir, err := ioutil.TempDir("", "repo")
	common.CheckIfError(err)

	_, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubCreds.GITHUB_USER,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	return tempDir
}

func CreateRepo(appName string, skeletonRepo string, skeletonRepoPath string) string {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoName := createRepo(appName)
	repoUrl := githubCreds.GITHUB_BASE + "/" + repoName

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
			Username: githubCreds.GITHUB_USER,
			Password: githubCreds.GITHUB_TOKEN,
		},
	})
	common.CheckIfError(err)

	//repoName := strings.ReplaceAll(strings.ToLower(appName), " ", "-")

	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		Auth: &http2.BasicAuth{
			Username: githubCreds.GITHUB_USER,
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
			Name:  githubCreds.GITHUB_USER,
			Email: githubCreds.GITHUB_EMAIL,
			When:  time.Now(),
		},
	})
	common.CheckIfError(err)
	obj, err := r.CommitObject(commit)
	common.CheckIfError(err)

	fmt.Println(obj)

	err = r.Push(&git.PushOptions{Auth: &http2.BasicAuth{Username: githubCreds.GITHUB_USER, Password: githubCreds.GITHUB_TOKEN}})
	common.CheckIfError(err)

	os.Chdir(curDir)

	return repoUrl
}

func DestroyRepo(appName string) {
	destroyRepo(appName)
}
