package tui

import "os"

// execReplacedProcess replaces the current process with bin (Unix exec).
// Tests override this so a successful rebuild cannot replace the test binary.
var execReplacedProcess = execReplacedProcessImpl

func relaunchPath(built string) string {
	if running := runningExecutable(); running != "" {
		if _, err := os.Stat(running); err == nil {
			return running
		}
	}
	return built
}

func tryRelaunch(builtPath string) bool {
	if builtPath == "" {
		return false
	}
	return execReplacedProcess(relaunchPath(builtPath)) == nil
}
