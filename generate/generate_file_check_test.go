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

package generate_test

import (
	"testing"

	"github.com/madlambda/spells/assert"
	"github.com/mineiros-io/terramate/generate"
	"github.com/mineiros-io/terramate/test/sandbox"
)

func TestCheckReturnsOutdatedStackFilenamesForGeneratedFile(t *testing.T) {
	s := sandbox.New(t)

	stackEntry := s.CreateStack("stacks/stack")
	stack := stackEntry.Load()

	assertOutdatedFiles := func(want []string) {
		t.Helper()

		got, err := generate.CheckStack(s.RootDir(), stack)
		assert.NoError(t, err)
		assertEqualStringList(t, got, want)
	}

	// Checking detection when there is no config generated yet
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("test.txt"),
				strAttr("content", "test"),
			),
		).String())
	assertOutdatedFiles([]string{"test.txt"})

	s.Generate()

	assertOutdatedFiles([]string{})

	// Now checking when we have code + it gets outdated.
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("test.txt"),
				strAttr("content", "changed"),
			),
		).String())

	assertOutdatedFiles([]string{"test.txt"})

	s.Generate()

	// Changing generated filenames will NOT trigger detection for the old file
	// since there is no way to automatically track the files for now
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("testnew.txt"),
				strAttr("content", "changed"),
			),
		).String())

	assertOutdatedFiles([]string{"testnew.txt"})

	// Adding new filename to generation trigger detection
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("testnew.txt"),
				strAttr("content", "changed"),
			),
			generateFile(
				labels("another.txt"),
				strAttr("content", "changed"),
			),
		).String())

	assertOutdatedFiles([]string{"another.txt", "testnew.txt"})

	s.Generate()

	assertOutdatedFiles([]string{})

	// Removed configurations will not be detected by default since there
	// is no way to track the files for now.
	stackEntry.CreateConfig(stackConfig().String())

	assertOutdatedFiles([]string{})
}

func TestCheckOutdatedIgnoresEmptyGenerateFileContent(t *testing.T) {
	s := sandbox.New(t)

	stackEntry := s.CreateStack("stacks/stack")
	stack := stackEntry.Load()

	assertOutdatedFiles := func(want []string) {
		t.Helper()

		got, err := generate.CheckStack(s.RootDir(), stack)
		assert.NoError(t, err)
		assertEqualStringList(t, got, want)
	}

	// Checking detection when the config is empty at first
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("test.txt"),
				strAttr("content", ""),
			),
		).String())

	assertOutdatedFiles([]string{})

	// Checking detection when the config isnt empty, is generated and then becomes empty
	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("test.txt"),
				strAttr("content", "test"),
			),
		).String())

	assertOutdatedFiles([]string{"test.txt"})

	s.Generate()

	assertOutdatedFiles([]string{})

	stackEntry.CreateConfig(
		stackConfig(
			generateFile(
				labels("test.txt"),
				strAttr("content", ""),
			),
		).String())

	assertOutdatedFiles([]string{"test.txt"})

	s.Generate()

	assertOutdatedFiles([]string{})
}