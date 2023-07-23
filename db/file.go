package db

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"

	"github.com/lordvidex/vpn-gate/utils"
)

const (
	configPath = "~/.config/vpn-gate"
	dbFile     = "db.dat"
)

func createDirIfNotExists(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0755)
		return err
	}
	return nil
}

// GetInstalled returns all the saved configurations in the db file
func GetInstalled() ([]string, error) {
	absDir := utils.ExpandHome(configPath)
	err := createDirIfNotExists(absDir)
	if err != nil {
		return nil, err
	}
	dbpath := filepath.Join(absDir, dbFile)
	file, err := os.OpenFile(dbpath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, errors.Join(errors.New("error opening db file"), err)
	}
	defer file.Close()

	// read file
	scanner := bufio.NewScanner(file)
	var entries []string
	for scanner.Scan() {
		entries = append(entries, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// SetInstalled saves the configurations in the db file
func SetInstalled(configNames []string) error {
	absDir := utils.ExpandHome(configPath)
	err := createDirIfNotExists(absDir)
	if err != nil {
		return err
	}
	dbpath := filepath.Join(absDir, dbFile)
	file, err := os.OpenFile(dbpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Join(errors.New("error opening db file"), err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, name := range configNames {
		_, err := writer.WriteString(name + "\n")
		if err != nil {
			return errors.Join(errors.New("error writing to db file"), err)
		}
	}
	return writer.Flush()
}
