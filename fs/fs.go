// Copyright 2022 Mineiros GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fs

import (
	"os"
	"strings"

	"github.com/mineiros-io/terramate/errors"
	"github.com/rs/zerolog/log"
)

// ListTerramateFiles returns a list of terramate related files from the
// directory dir.
func ListTerramateFiles(dir string) (filenames []string, err error) {
	logger := log.With().
		Str("action", "fs.listTerramateFiles()").
		Str("dir", dir).
		Logger()

	logger.Trace().Msg("listing files")

	f, err := os.Open(dir)
	if err != nil {
		return nil, errors.E(err, "opening directory %s for reading file entries", dir)
	}

	defer func() {
		err = errors.L(err, f.Close()).AsError()
	}()

	dirEntries, err := f.ReadDir(-1)
	if err != nil {
		return nil, errors.E(err, "reading dir to list Terramate files")
	}

	logger.Trace().Msg("looking for Terramate files")

	files := []string{}

	for _, entry := range dirEntries {
		fname := entry.Name()

		logger := logger.With().
			Str("entryName", fname).
			Logger()

		if entry.IsDir() || !isTerramateFile(fname) {
			logger.Trace().Msg("ignoring file")
			continue
		}

		logger.Trace().Msg("Found Terramate file")
		files = append(files, fname)
	}

	return files, nil
}

// ListTerramateDirs lists Terramate dirs, which are any dirs
// except ones starting with ".".
func ListTerramateDirs(dir string) ([]string, error) {
	logger := log.With().
		Str("action", "fs.ListTerramateDirs()").
		Str("dir", dir).
		Logger()

	logger.Trace().Msg("listing dirs")

	f, err := os.Open(dir)
	if err != nil {
		return nil, errors.E(err, "opening directory %s for reading file entries", dir)
	}

	defer func() {
		err = errors.L(err, f.Close()).AsError()
	}()

	dirEntries, err := f.ReadDir(-1)
	if err != nil {
		return nil, errors.E(err, "reading dir to list Terramate dirs")
	}

	logger.Trace().Msg("looking for Terramate directories")

	dirs := []string{}

	for _, dirEntry := range dirEntries {
		fname := dirEntry.Name()

		logger := logger.With().
			Str("entryName", fname).
			Logger()

		if fname[0] == '.' || !dirEntry.IsDir() {
			logger.Trace().Msg("ignoring file")
			continue
		}

		dirs = append(dirs, fname)
	}

	return dirs, nil
}

func isTerramateFile(filename string) bool {
	if len(filename) <= 3 || filename[0] == '.' {
		return false
	}
	switch filename[len(filename)-1] {
	default:
		return false
	case 'l':
		return strings.HasSuffix(filename, ".tm.hcl")
	case 'm':
		return strings.HasSuffix(filename, ".tm")
	}
}
