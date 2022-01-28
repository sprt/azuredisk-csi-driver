<<<<<<< HEAD
<<<<<<< HEAD
//go:build windows
=======
>>>>>>> upgrade to k8s 1.23 lib
=======
//go:build windows
>>>>>>> chore: Merge changes from upstream as of 2022-01-26 (#351)
// +build windows

package cobra

import (
	"fmt"
	"os"
	"time"

	"github.com/inconshreveable/mousetrap"
)

var preExecHookFn = preExecHook

func preExecHook(c *Command) {
	if MousetrapHelpText != "" && mousetrap.StartedByExplorer() {
		c.Print(MousetrapHelpText)
		if MousetrapDisplayDuration > 0 {
			time.Sleep(MousetrapDisplayDuration)
		} else {
			c.Println("Press return to continue...")
			fmt.Scanln()
		}
		os.Exit(1)
	}
}
