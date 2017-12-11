package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aquiladev/2leveldb/source"
	"github.com/aquiladev/2leveldb/util"
	"github.com/pkg/errors"
)

type config struct {
	LevelDBPath string
	Source      *source.Config

	LogDir string `long:"logdir" description:"Directory to log output."`
}

const (
	defaultConfigFilename = "2leveldb.conf"
	defaultLogDirname     = "logs"
	defaultLogFilename    = "2leveldb.log"
)

var (
	defaultHomeDir    = util.AppDataDir("2leveldb", false)
	defaultConfigFile = filepath.Join(defaultHomeDir, defaultConfigFilename)
	defaultLogDir     = filepath.Join(defaultHomeDir, defaultLogDirname)
)

// cleanAndExpandPath expands environment variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// Expand initial ~ to OS specific home directory.
	if strings.HasPrefix(path, "~") {
		homeDir := filepath.Dir(defaultHomeDir)
		path = strings.Replace(path, "~", homeDir, 1)
	}

	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%,
	// but they variables can still be expanded via POSIX-style $VARIABLE.
	return filepath.Clean(os.ExpandEnv(path))
}

func loadConfig() (*config, error) {
	// Default config.
	cfg := &config{
		LogDir: defaultLogDir,
	}

	bytes, err := ioutil.ReadFile(defaultConfigFile)
	if err != nil {
		err = errors.Wrap(err, "Unable to read file")
		return cfg, err
	}

	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		err = errors.Wrap(err, "Error while parsing config file\n"+string(bytes))

		return cfg, err
	}

	// Create the home directory if it doesn't already exist.
	funcName := "loadConfig"
	err = os.MkdirAll(defaultHomeDir, 0700)
	if err != nil {
		// Show a nicer error message if it's because a symlink is
		// linked to a directory that does not exist (probably because
		// it's not mounted).
		if e, ok := err.(*os.PathError); ok && os.IsExist(err) {
			if link, lerr := os.Readlink(e.Path); lerr == nil {
				err = fmt.Errorf("is symlink %s -> %s mounted? ", e.Path, link)
			}
		}

		err := fmt.Errorf("%s: Failed to create home directory: %v", funcName, err)
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	// Append the network type to the log directory so it is "namespaced"
	// per network in the same fashion as the data directory.
	cfg.LogDir = cleanAndExpandPath(cfg.LogDir)

	// Initialize log rotation.  After log rotation has been initialized, the
	// logger variables may be used.
	initLogRotator(filepath.Join(cfg.LogDir, defaultLogFilename))

	return cfg, nil
}
