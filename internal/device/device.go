package device

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/goburrow/serial"
)

type crlfReader struct {
	r *bufio.Reader
}

type Cmd struct {
	CmdText     string
	Description string
}

type Terminators struct {
	TX string
	RX string
}

type Device struct {
	config        serial.Config
	p             serial.Port
	Commands      []Cmd
	reader        *crlfReader
	IsOpen        bool
	tx_terminator string
	rx_terminator string
}

func NewDevice(config serial.Config, terminators Terminators) Device {
	return Device{
		config:        config,
		tx_terminator: terminators.TX,
		rx_terminator: terminators.RX,
	}
}

func (device *Device) Open() (err error) {
	device.IsOpen = false
	device.p, err = serial.Open(&device.config)
	if err != nil {
		return err
	}
	device.IsOpen = true
	device.reader = &crlfReader{
		r: bufio.NewReader(device.p),
	}
	return err
}

func (device *Device) Close() error {
	return device.p.Close()
}

func (device *Device) Write(cmd string) (int, error) {
	if device.IsOpen {
		return fmt.Fprintf(device.p, "%s%s", cmd, device.tx_terminator)
	}
	return 0, errors.New("device is not open, unable to send")
}

func (device *Device) Read(b []byte) (n int, err error) {
	// Loop until we find the terminator
	for !strings.HasSuffix(string(b), device.rx_terminator) {
		var c byte
		c, err = device.reader.r.ReadByte()
		if err != nil {
			break
		}
		b[n] = c
		n++
	}
	return n, err
}

// This is a single function to send a command and wait for a response
func (device *Device) TXcmdRXresponse(cmd string) (response string, err error) {
	if !device.IsOpen {
		return "", errors.New("port is not open, unable to send data")
	}
	// Send command
	fmt.Fprintf(device.p, "%s%s", cmd, device.tx_terminator)
	// Get response
	var input strings.Builder
	// Loop until we find the terminator
	for !strings.HasSuffix(input.String(), device.rx_terminator) {
		var c byte
		c, err = device.reader.r.ReadByte()
		if err != nil {
			break
		}
		input.WriteString(string(c))
	}
	return input.String(), err
}

// This should be called in a routine, it will run forever
func (device *Device) RXLogsForever(f func(byte)) {

	for {
		var c, err = device.reader.r.ReadByte()
		if err == nil {
			f(c)
		}
	}
}

// This gets the supported commands from the device via the "help" command
// It expects this format:
// "`help`                 Get help/usage for commands\n"
// "`comm_test`            Request communications test\n"
// TODO: update this to support the defaults file terminators
func (device *Device) LoadCmds() error {
	device.Commands = make([]Cmd, 0)
	if _, err := device.Write("help"); err != nil {
		return err
	}
	scanner := bufio.NewScanner(device.p)
	for scanner.Scan() {
		s := scanner.Text()
		if strings.HasPrefix(s, "`") {
			seperatedCmdString := strings.Split(s, "`")
			if len(seperatedCmdString) != 3 {
				return fmt.Errorf("invalid cmd syntax: %s", s)
			}
			cmd := Cmd{
				CmdText:     seperatedCmdString[1],
				Description: strings.TrimSpace(seperatedCmdString[2]),
			}
			device.Commands = append(device.Commands, cmd)
		}
	}
	return scanner.Err()
}
