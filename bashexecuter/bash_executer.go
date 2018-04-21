package bashexecuter

import (
	"os/exec"

	"git.workshop21.ch/go/abraxas/logging"
)

// Execute a given command in bash, returns error
func Execute(cmdStr string) (string, error) {
	out, err := exec.Command("/bin/bash", "-c", cmdStr).Output()
	if err != nil {
		logging.WithError("BA-BashExecuter-piu3240h897q340h87q3", err)
	}
	return string(out), err
}
