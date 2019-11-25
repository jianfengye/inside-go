// +build windows

package cobra

import (
	"fmt"
	"os"
	"time"

	"github.com/inconshreveable/mousetrap"
)

var preExecHookFn = preExecHook

// windows下只允许在命令行打开（powershell）,而不允许双击打开
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
