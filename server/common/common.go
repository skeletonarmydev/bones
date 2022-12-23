package common

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type TerraformAction int64

const (
	PlanAction    TerraformAction = 0
	ApplyAction                   = 1
	DestroyAction                 = 2
)

type AWSCreds struct {
	AWS_REGION     string
	AWS_ACCESS_KEY string
	AWS_SECRET_KEY string
}

func getTerraformDir() (execPath string) {

	if GetConfig("SA_LOCAL") == "true" {
		return "/usr/local/bin/terraform"
	} else {
		return "/usr/bin/terraform"
	}

}

func CheckIfError(err error) {
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

// File copies a single file from src to dst
func File(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// Dir copies a whole directory recursively
func Dir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = Dir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = File(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func GetConfig(key string) (value string) {
	return os.Getenv(key)
}

func runApplyTerraform(workingDir string, vars map[string]string, statefileDir string) error {
	execPath := getTerraformDir()

	awsCredsEnv := GetConfig("AWS")

	var awsCreds AWSCreds
	err := json.Unmarshal([]byte(awsCredsEnv), &awsCreds)
	if err != nil {
		fmt.Printf("Can't parse awsCreds: %s", err)
		return err
	}

	os.Remove(workingDir + "/out.plan")

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		fmt.Printf("error running NewTerraform: %s (execPath: %s)", err, execPath)
		return err
	}

	err = tf.Init(context.Background(),
		tfexec.BackendConfig("region=us-east-1"),
		tfexec.BackendConfig("bucket=bones-server"),
		tfexec.BackendConfig("encrypt=true"),
		tfexec.BackendConfig("key="+statefileDir+"/terraform.tfstate"),
		tfexec.BackendConfig("access_key="+awsCreds.AWS_ACCESS_KEY),
		tfexec.BackendConfig("secret_key="+awsCreds.AWS_SECRET_KEY),
	)
	if err != nil {
		fmt.Printf("error running Init: %s", err)
		return err
	}

	os.MkdirAll("/tmp/"+statefileDir, 0777)

	// Convert map to slice of keys.
	var tfvars = []tfexec.PlanOption{tfexec.Out(workingDir + "/out.plan")}
	for key, val := range vars {
		tfvars = append(tfvars, tfexec.Var(key+"="+val))
	}

	pass, err := tf.Plan(context.Background(), tfvars...)
	if err != nil {
		fmt.Printf("error running Plan: %s", err)
		return err
	}

	if pass {
		plan, err := tf.ShowPlanFile(context.Background(), workingDir+"/out.plan")
		if err != nil {
			fmt.Printf("error running fetch plan: %s", err)
			return err
		}

		for _, s := range plan.ResourceChanges {
			fmt.Printf("Change: %s %s\n", s.Change.Actions, s.Name)
		}

		fmt.Println("Applying changes")
		err2 := tf.Apply(context.Background(), tfexec.DirOrPlan(workingDir+"/out.plan"))

		if err2 != nil {
			fmt.Printf("error running apply: %s", err2)
			return err2
		}

		os.Remove(workingDir + "/out.plan")
	}

	return nil
}

func runDestroyTerraform(workingDir string, vars map[string]string, statefileDir string) error {
	execPath := getTerraformDir()

	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		fmt.Printf("error running NewTerraform: %s (execPath: %s)", err, execPath)
		return err
	}

	awsCredsEnv := GetConfig("AWS")

	var awsCreds AWSCreds
	err = json.Unmarshal([]byte(awsCredsEnv), &awsCreds)
	if err != nil {
		fmt.Printf("Can't parse awsCreds: %s", err)
		return err
	}

	err = tf.Init(
		context.Background(),
		tfexec.BackendConfig("region=us-east-1"),
		tfexec.BackendConfig("bucket=bones-server"),
		tfexec.BackendConfig("encrypt=true"),
		tfexec.BackendConfig("key="+statefileDir+"/terraform.tfstate"),
		tfexec.BackendConfig("access_key="+awsCreds.AWS_ACCESS_KEY),
		tfexec.BackendConfig("secret_key="+awsCreds.AWS_SECRET_KEY),
	)
	if err != nil {
		fmt.Printf("error running Init: %s", err)
		return err
	}

	// Convert map to slice of keys.
	var tfvars = []tfexec.DestroyOption{}
	for key, val := range vars {
		tfvars = append(tfvars, tfexec.Var(key+"="+val))
	}

	fmt.Println("Destroying changes")
	err = tf.Destroy(context.Background(), tfvars...)
	if err != nil {
		fmt.Printf("error running destroy: %s", err)
		return err
	}

	return nil
}

func ExecuteTerraform(workingDir string, vars map[string]string, action TerraformAction, statefileDir string) error {
	switch action {
	case PlanAction:
		return nil
	case ApplyAction:
		return runApplyTerraform(workingDir, vars, statefileDir)
	case DestroyAction:
		return runDestroyTerraform(workingDir, vars, statefileDir)
	}

	return nil
}
