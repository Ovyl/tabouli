package defaults

import (
	"errors"
	"log"
	"tabouli/internal/file_io"

	"gopkg.in/yaml.v2"
)

// YAML Parsing Only
type DefaultFile struct {
	CLIBaudRate      int    `yaml:"cli_baud"`
	CLIDataBits      int    `yaml:"cli_data_bits"`
	CLIStopBits      int    `yaml:"cli_stop_bits"`
	CLIParity        string `yaml:"cli_parity"`
	CLITXTerminator  string `yaml:"cli_tx_terminator"`
	CLIRXTerminator  string `yaml:"cli_rx_terminator"`
	LogsBaudRate     int    `yaml:"logs_baud"`
	LogsDataBits     int    `yaml:"logs_data_bits"`
	LogsStopBits     int    `yaml:"logs_stop_bits"`
	LogsParity       string `yaml:"logs_parity"`
	LogsTXTerminator string `yaml:"logs_tx_terminator"`
	LogsRXTerminator string `yaml:"logs_rx_terminator"`
}

// func NewTestFileAutomator() (TestFileAutomator, error) {
func NewDefaultFileHanlder(cli bool, logs bool) (DefaultFile, error) {
	var file = file_io.FindFilesWithPattern(".", "defaults.yaml")
	defaultFile := DefaultFile{}
	if file == nil {
		return defaultFile, errors.New("default file not found")
	}

	var defaultFileContent = file_io.GetFileContents(file[0])

	err := yaml.Unmarshal([]byte(defaultFileContent), &defaultFile)
	if err != nil {
		log.Fatalf("error: %v", err)
		return defaultFile, errors.New("error parsing yaml")
	}

	if cli {
		if defaultFile.CLIRXTerminator == "" || defaultFile.CLITXTerminator == "" {
			log.Fatal("cli terminators not found in defaults file")
			return defaultFile, errors.New("error with yaml terminators")
		}
	}

	if logs {
		if defaultFile.LogsRXTerminator == "" || defaultFile.LogsTXTerminator == "" {
			log.Fatal("logs terminators not found in defaults file")
			return defaultFile, errors.New("error with yaml terminators")
		}
	}

	return defaultFile, nil
}
