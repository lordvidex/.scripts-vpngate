// Package osascripts contains all the AppleScript functions that are executed from the command line
// to interact with Tunnelblick application
package osascripts

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/lordvidex/vpn-gate/db"
	"github.com/lordvidex/vpn-gate/utils"
)

// DeleteInstalledConfigs only deletes configurations that were downloaded with vpn-gate
func DeleteInstalledConfigs(except ...string) error {
	// configs in this map should not be deleted
	excluded := make(map[string]struct{})
	for _, s := range except {
		s = utils.RemoveExtension(s)
		excluded[s] = struct{}{}
	}

	// configs to be deleted
	stored, err := db.GetInstalled()
	if err != nil {
		return err
	}
	if len(stored) == 0 {
		return nil
	}
	shouldDelete := make(map[string]bool)
	for _, it := range stored {
		shouldDelete[it] = true
	}

	// checking tunnelblick to delete configs
	filesInPath := TunnelblickFiles()

	errs := make([]error, 0)
	for _, config := range filesInPath {
		name := utils.RemoveExtension(path.Base(config))
		if _, ok := excluded[name]; ok {
			shouldDelete[name] = false
			continue
		}
		if _, ok := shouldDelete[name]; ok {
			err = os.RemoveAll(config)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}
	stored = stored[:0]
	for k, deleted := range shouldDelete {
		if !deleted {
			stored = append(stored, k)
		}
	}
	err = db.SetInstalled(stored)
	if err != nil {
		return err
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func InstallConfigs(configs []string) error {
	installed, err := db.GetInstalled()
	if err != nil {
		return err
	}
	for i := 0; i < len(configs); i++ {
		err := exec.Command("open", "-a", "Tunnelblick", configs[i]).Run()
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(configs); i++ {
		installed = append(installed, utils.RemoveExtension(configs[i]))
	}
	return db.SetInstalled(installed)
}

// GetConfigs returns all the current configurations either installed by vpn-gate or not
func GetConfigs() ([]string, error) {
	cmd := exec.Command("osascript", "-e", `tell application "Tunnelblick"
	get name of configurations
	end tell`)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	arr := strings.Split(string(output), ",")
	for i := 0; i < len(arr); i++ {
		arr[i] = strings.TrimSpace(arr[i])
	}
	return arr, nil
}
