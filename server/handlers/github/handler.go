package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	http2 "github.com/go-git/go-git/v5/plumbing/transport/http"
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

type RemoteFile struct {
	Name string
	Path string
	Data []byte
	Perm os.FileMode
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

	err = common.ExecuteTerraform(getWorkingDir(), vars, common.ApplyAction, repoName+"/infra/github")
	common.CheckIfError(err)

	return repoName
}

func DownloadRepo(repo string) string {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse github environment: %s", err)
	}

	tempDir, err := os.MkdirTemp("", "repo")
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

func AddFilesToRepo(repo string, commitMessage string, files []RemoteFile) {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoDir, err := os.MkdirTemp("", "repo")
	common.CheckIfError(err)
	defer os.RemoveAll(repoDir)

	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:      repo,
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

	err = os.Chdir(repoDir)
	common.CheckIfError(err)

	for _, fl := range files {
		err = os.MkdirAll(fl.Path, fl.Perm)
		common.CheckIfError(err)

		err = os.WriteFile(fl.Path+"/"+fl.Name, fl.Data, fl.Perm)
		common.CheckIfError(err)

		_, err = w.Add(fl.Path + "/" + fl.Name)
		common.CheckIfError(err)
	}

	commit, err := w.Commit(commitMessage, &git.CommitOptions{
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

	err = r.Push(&git.PushOptions{Auth: &http2.BasicAuth{
		Username: githubCreds.GITHUB_USER,
		Password: githubCreds.GITHUB_TOKEN,
	}})
	common.CheckIfError(err)
}

func CreateRepo(appName string, skeletonRepo string, skeletonRepoPath string, data map[string]string) string {
	githubCredsEnv := os.Getenv("GITHUB")

	var githubCreds GithubCreds
	err := json.Unmarshal([]byte(githubCredsEnv), &githubCreds)
	if err != nil {
		log.Fatalf("Can't parse githubToken: %s", err)
	}

	repoName := createRepo(appName)
	repoUrl := githubCreds.GITHUB_BASE + "/" + repoName

	skeletonDir, err := os.MkdirTemp("", "skeleton")
	common.CheckIfError(err)
	defer os.RemoveAll(skeletonDir)

	repoDir, err := os.MkdirTemp("", "repo")
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

func DestroyRepo(name string) {
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

	err = common.ExecuteTerraform(getWorkingDir(), vars, common.DestroyAction, repoName+"/infra/github")
	common.CheckIfError(err)
}
