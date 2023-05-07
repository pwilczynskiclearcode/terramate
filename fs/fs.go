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
	"path/filepath"
	"strings"

	"github.com/mineiros-io/terramate/errors"
	"github.com/rs/zerolog/log"
)

// ListTerramateFiles returns a list of terramate related files from the
// directory dir.
func ListTerramateFiles(dir string) ([]string, error) {
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

	filenames, err := f.Readdirnames(0)
	if err != nil {
		return nil, errors.E(err, "reading dir to list Terramate files")
	}

	logger.Trace().Msg("looking for Terramate files")

	files := []string{}

	for _, filename := range filenames {
		logger := logger.With().
			Str("entryName", filename).
			Logger()

		if !isTerramateFile(filename) {
			logger.Trace().Msg("ignoring file")
			continue
		}

		abspath := filepath.Join(dir, filename)
		st, err := os.Lstat(abspath)
		if err != nil {
			return nil, err
		}

		if st.IsDir() {
			logger.Trace().Msg("ignoring dir")
			continue
		}

		logger.Trace().Msg("Found Terramate file")
		files = append(files, filename)
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

	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.E(err, "reading dir to list Terramate dirs")
	}

	logger.Trace().Msg("looking for Terramate directories")

	dirs := []string{}

	for _, dirEntry := range dirEntries {
		logger := logger.With().
			Str("entryName", dirEntry.Name()).
			Logger()

		if !dirEntry.IsDir() {
			logger.Trace().Msg("ignoring non-dir")
			continue
		}

		if strings.HasPrefix(dirEntry.Name(), ".") {
			logger.Trace().Msg("ignoring dotdir")
			continue
		}

		dirs = append(dirs, dirEntry.Name())
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
