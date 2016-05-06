package runner

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
)

// SetEnv formats the env key / value pair
func SetEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

// GetEnviron overrides and formats the env
func GetEnviron(variables ...string) []string {
	osEnv := os.Environ()
	newEnv := []string{}
	for _, env := range osEnv {
		r := regexp.MustCompile(`(?i)(PATH|DEBUG)\=`)
		if !r.MatchString(env) {
			newEnv = append(newEnv, env)
		}
	}
	return append(newEnv, variables...)
}

// GetPathEnv handles the Path craziness
func GetPathEnv(binPath string) string {
	if runtime.GOOS == "darwin" {
		basePath := "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/opt/X11/bin"
		return SetEnv("PATH", fmt.Sprintf("%s:%s", binPath, basePath))
	} else if runtime.GOOS == "windows" {
		return SetEnv("Path", fmt.Sprintf("%s;%s", os.Getenv("PATH"), binPath))
	}
	return SetEnv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), binPath))
}
