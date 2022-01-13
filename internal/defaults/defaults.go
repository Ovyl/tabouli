package defaults

import (
	"errors"
	"log"
	"tabouli/internal/file_io"

	"gopkg.in/yaml.v2"
)

// YAML Parsing Only
type DefaultFile struct {
	SerialBaudRate int    `yaml:"baud"`
	SerialDataBits int    `yaml:"data_bits"`
	SerialStopBits int    `yaml:"stop_bits"`
	SerialParity   string `yaml:"parity"`
	TXTerminator   string `yaml:"tx_terminator"`
	RXTerminator   string `yaml:"rx_terminator"`
}

// func NewTestFileAutomator() (TestFileAutomator, error) {
func NewDefaultFileHanlder() (DefaultFile, error) {
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

	if defaultFile.RXTerminator == "" || defaultFile.TXTerminator == "" {
		log.Fatal("terminators not found in defaults file")
		return defaultFile, errors.New("error with yaml terminators")
	}

	return defaultFile, nil
}
