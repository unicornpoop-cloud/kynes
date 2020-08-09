package kynes

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
	Repo string `short:"r" desc:"Repo to apply terraform from"`

	ConfigFile string `id:"config" short:"C"`
}

var config = Config{
	Repo: "https://github.com/unicornpoop-cloud/gitops-terraform-demo",
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
	_, err = git.PlainClone("./terraform", false, &git.CloneOptions{
		URL:      config.Repo,
		Progress: os.Stdout,
	})

	logInfo("Running init")
	_, err = exec.Command("terraform", "init", "./terraform").Output()
	checkIfError(err)

	tfPlanCmd := exec.Command("terraform", "plan", "-detailed-exitcode", "-out=tfplan.out", "./terraform")
	var outPlan bytes.Buffer
	tfPlanCmd.Stdout = &outPlan
	logInfo("Running plan")
	err = tfPlanCmd.Run()
	if err != nil {
		fmt.Printf("Error planning: %s\n", outPlan.String())
	}
	planExitCode := tfPlanCmd.ProcessState.ExitCode()

	if planExitCode == 0 {
		logInfo("No apply needed")
	}

	if planExitCode == 1 {
		fmt.Printf("Plan had errors: %s", outPlan.String())
	}

	if planExitCode == 2 {
		logInfo("Running apply")
		tfApplyCmd, err := exec.Command("terraform", "apply", "-auto-approve", "tfplan.out").Output()
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
