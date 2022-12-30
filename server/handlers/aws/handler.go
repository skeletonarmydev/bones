package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	github "github.com/bones/server/handlers/github"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type AWSCreds struct {
	AWS_REGION     string
	AWS_ACCESS_KEY string
	AWS_SECRET_KEY string
}

func CreateAWSInfra(name string, repo string, skeletonRepo string, skeletonRepoPath string, data map[string]string) error {

	fmt.Printf("Creating AWS Infra for app: %s\n", name)

	skeletonDir := github.DownloadRepo(skeletonRepo)
	defer os.RemoveAll(skeletonDir)

	workingDir := skeletonDir + skeletonRepoPath + "/infra/aws-ecs"

	awsCredsEnv := common.GetConfig("AWS")

	var awsCreds AWSCreds
	err := json.Unmarshal([]byte(awsCredsEnv), &awsCreds)
	if err != nil {
		fmt.Printf("Can't parse awsCreds: %s", err)
		return err
	}

	fmt.Printf("Create AWS Infra: %s\n", data["APP_NAME"])

	vars := make(map[string]string)
	vars["vpc_id"] = "vpc-c92c8baf"
	vars["aws_region"] = awsCreds.AWS_REGION
	vars["aws_access_key"] = awsCreds.AWS_ACCESS_KEY
	vars["aws_secret_key"] = awsCreds.AWS_SECRET_KEY

	//Process template
	var files = []github.RemoteFile{}

	err = filepath.Walk(workingDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				tmpl, err := template.ParseFiles(path)
				common.CheckIfError(err)
				buf := &bytes.Buffer{}
				err = tmpl.Execute(buf, data)

				files = append(files, github.RemoteFile{
					Name: info.Name(),
					Path: "infra/aws-ecs",
					Data: buf.Bytes(),
					Perm: 0750,
				})

				os.WriteFile(path, buf.Bytes(), 0750)
			}

			return nil
		})

	github.AddFilesToRepo(repo, "Process AWS Terraform file", files)

	err = common.ExecuteTerraform(workingDir, vars, common.ApplyAction, data["APP_NAME"]+"/infra/aws-ecs")

	fmt.Printf("Finished creating AWS Infra for app: %s\n", name)

	return err
}

func DestroyAWSInfra(name string, repo string) error {

	projectDir := github.DownloadRepo(repo)
	defer os.RemoveAll(projectDir)

	workingDir := projectDir + "/infra/aws-ecs"

	awsCredsEnv := common.GetConfig("AWS")

	var awsCreds AWSCreds
	err := json.Unmarshal([]byte(awsCredsEnv), &awsCreds)
	if err != nil {
		fmt.Printf("Can't parse awsCreds: %s", err)
		return err
	}
	appName := strings.ReplaceAll(strings.ToLower(name), " ", "-")

	fmt.Printf("Destroy AWS Infra: %s\n", appName)

	vars := make(map[string]string)
	vars["vpc_id"] = "vpc-c92c8baf"
	//vars["app_name"] = appName
	vars["aws_region"] = awsCreds.AWS_REGION
	vars["aws_access_key"] = awsCreds.AWS_ACCESS_KEY
	vars["aws_secret_key"] = awsCreds.AWS_SECRET_KEY

	err = common.ExecuteTerraform(workingDir, vars, common.DestroyAction, appName+"/infra/aws-ecs")
	return err
}
