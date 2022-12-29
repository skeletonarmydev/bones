package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/bones/server/common"
	github "github.com/bones/server/handlers/github"
	"os"
	"strings"
)

type AWSCreds struct {
	AWS_REGION     string
	AWS_ACCESS_KEY string
	AWS_SECRET_KEY string
}

func CreateAWSInfra(name string, skeletonRepo string, skeletonRepoPath string) error {

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

	appName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	fmt.Printf("Create AWS Infra: %s\n", appName)

	vars := make(map[string]string)
	vars["vpc_id"] = "vpc-c92c8baf"
	//vars["app_name"] = appName
	vars["aws_region"] = awsCreds.AWS_REGION
	vars["aws_access_key"] = awsCreds.AWS_ACCESS_KEY
	vars["aws_secret_key"] = awsCreds.AWS_SECRET_KEY

	err = common.ExecuteTerraform(workingDir, vars, common.ApplyAction, appName+"/infra/aws-ecs")

	fmt.Printf("Finished creating AWS Infra for app: %s\n", name)

	return err
}

func DestroyAWSInfra(name string, skeletonRepo string, skeletonRepoPath string) error {

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
