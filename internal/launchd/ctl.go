package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func uid() string {
	return strconv.Itoa(os.Getuid())
}

func Bootstrap(plistPath string) error {
	target := fmt.Sprintf("gui/%s", uid())
	cmd := exec.Command("launchctl", "bootstrap", target, plistPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl bootstrap: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func Bootout(name string) error {
	target := fmt.Sprintf("gui/%s/%s", uid(), Label(name))
	cmd := exec.Command("launchctl", "bootout", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		if strings.Contains(outStr, "No such process") || strings.Contains(outStr, "Could not find specified service") {
			return nil
		}
		return fmt.Errorf("launchctl bootout: %s (%w)", outStr, err)
	}
	return nil
}

func IsLoaded(name string) bool {
	target := fmt.Sprintf("gui/%s/%s", uid(), Label(name))
	cmd := exec.Command("launchctl", "print", target)
	return cmd.Run() == nil
}

func InstallPlist(data []byte, name string) error {
	plistPath, err := PlistPath(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}
	return os.WriteFile(plistPath, data, 0644)
}

func RemovePlist(name string) error {
	plistPath, err := PlistPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove plist: %w", err)
	}
	return nil
}
