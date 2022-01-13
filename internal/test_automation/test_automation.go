package test_automation

import (
	"errors"
	"log"
	"tabouli/internal/file_io"

	"gopkg.in/yaml.v2"
)

/*
*	Call NewTestFileAutomator to create a new automator
*	Then it should expose a list of FileNames and bool flag for handling if a test is in progress
*	Each TestFile has a name and list of commands as parsed from the YAML file
 */
type TestFile struct {
	TestFileName string
	CommandList  []string
}

type TestFileAutomator struct {
	TestFiles      []TestFile
	TestInProgress bool
}

// YAML Parsing Only
type CmdFile struct {
	Cmds []string `yaml:"commands"`
}

func NewTestFileAutomator() (TestFileAutomator, error) {
	var files = file_io.FindFilesWithPattern(".", "test_*.yaml")
	automator := TestFileAutomator{}
	if files == nil {
		return automator, errors.New("no test files found")
	}
	for _, fileName := range files {
		var fileContents = file_io.GetFileContents(fileName)

		// Parse into YAML
		cmdFile := CmdFile{}
		cmdFile = getTestFileContentsAsCmdFile(fileContents)

		// Put into public struct
		testFile := TestFile{}
		testFile.TestFileName = fileName
		testFile.CommandList = append(testFile.CommandList, cmdFile.Cmds...)

		// Add the list list of test files
		automator.TestFiles = append(automator.TestFiles, testFile)
	}
	automator.TestInProgress = false
	return automator, nil
}

func getTestFileContentsAsCmdFile(fileContents string) CmdFile {
	cmdFile := CmdFile{}
	err := yaml.Unmarshal([]byte(fileContents), &cmdFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return cmdFile

}
