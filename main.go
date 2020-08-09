package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/stevenroose/gonfig"
)

// Config is the configuration for the application
type Config struct {
	Repo         string `short:"r" desc:"Repo to apply terraform from"`
	TerraformDir string `short:"t" desc:"Directory to run the terraform files from"`
	ConfigFile   string `id:"config" short:"C"`
}

var config = Config{
	Repo:         "https://github.com/unicornpoop-cloud/gitops-terraform-demo",
	TerraformDir: "/tmp/workdir",
}

func main() {
	err := gonfig.Load(&config, gonfig.Conf{
		ConfigFileVariable:  "config",
		FileDefaultFilename: "kynes-config.yaml",
		FileDecoder:         gonfig.DecoderYAML,

		EnvPrefix: "KYNES_",
	})
	checkIfError(err)

	logInfo("git clone " + config.Repo)
	var r *git.Repository

	r, err = git.PlainClone(config.TerraformDir, false, &git.CloneOptions{
		URL:      config.Repo,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		logInfo("repo already exists - pull latest")
		r, err = git.PlainOpen(config.TerraformDir)
		checkIfError(err)

		w, err := r.Worktree()
		checkIfError(err)

		err = w.Pull(&git.PullOptions{
			RemoteName: "origin",
			Progress:   os.Stdout,
		})

		if err == git.NoErrAlreadyUpToDate {
			logInfo("already up to date - let's continue")
		} else {
			checkIfError(err)
		}
	}

	ref, err := r.Head()
	checkIfError(err)

	commit, err := r.CommitObject(ref.Hash())
	checkIfError(err)

	fmt.Printf("Current commit: %s\n", commit)

	logInfo("Running init")
	_, err = exec.Command("terraform", "init", config.TerraformDir).Output()
	checkIfError(err)

	planFileName := config.TerraformDir + "/tfplan.out"
	stateFileName := config.TerraformDir + "/terraform.tfstate"

	tfPlanCmd := exec.Command("terraform", "plan", "-detailed-exitcode", "-state="+stateFileName, "-out="+planFileName, config.TerraformDir)
	var outPlan bytes.Buffer
	tfPlanCmd.Stdout = &outPlan
	logInfo("Running plan")
	err = tfPlanCmd.Run()

	planExitCode := tfPlanCmd.ProcessState.ExitCode()

	if planExitCode == 0 {
		logInfo("No apply needed")
		tfOutputCmd, err := exec.Command("terraform", "output", "-state="+stateFileName).Output()
		checkIfError(err)
		fmt.Printf("%s", tfOutputCmd)
	}

	if planExitCode == 1 {
		fmt.Printf("Plan had errors: %s", outPlan.String())
	}

	if planExitCode == 2 {
		fmt.Printf("Changes detected: \n%s\nRunning apply.", outPlan.String())
		tfApplyCmd, err := exec.Command("terraform", "apply", "-state="+stateFileName, "-auto-approve", planFileName).Output()
		checkIfError(err)
		fmt.Printf("%s", tfApplyCmd)
	}
}

func logInfo(msg string) {
	log.Print(msg)
}

func checkIfError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
