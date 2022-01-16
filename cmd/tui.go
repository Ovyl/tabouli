/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"tabouli/internal/defaults"
	"tabouli/internal/device"
	"tabouli/internal/tui"
	"time"

	"github.com/goburrow/serial"
	"github.com/spf13/cobra"
)

var CLI_Port_Address string
var Logs_Port_Address string

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Runs the tui app to help with the aid of creating the test files",
	Example: `tabouli tui /dev/tty.usbserial-1446120:115200
	`,
	Args: argsValidation,
	Run: func(cmd *cobra.Command, args []string) {
		var cliPassed = false
		var logsPassed = false
		argCLI, err := cmd.Flags().GetString("cli")
		if err == nil && argCLI != "" {
			cliPassed = true
		}
		argLogs, err := cmd.Flags().GetString("logs")
		if err == nil && argLogs != "" {
			logsPassed = true
		}
		defaults, err := defaults.NewDefaultFileHanlder(cliPassed, logsPassed)
		if err != nil {
			log.Fatal(err)
		}

		var cliDevice device.Device
		var logDevice device.Device

		if cliPassed {
			cliDevice = initCLIDevice(argCLI, defaults)
			// Attempt to open the cli device
			if err := cliDevice.Open(); err != nil {
				log.Fatalf("Failed to open com port for provided CLI device address: %v", err)
			}
		}
		if logsPassed {
			logDevice = initLogDevice(argLogs, defaults)
			// Attempt to open the log device
			if err := logDevice.Open(); err != nil {
				log.Fatalf("Failed to open com port for provided Log device address: %v", err)
			}
		}

		// Build the UI
		ui := tui.CreateTView(cliDevice, logDevice)

		// Setup callback for logs received from device to display in the UI
		if logsPassed {
			go logDevice.RXLogsForever(tui.LogToUI)
		}

		if err := ui.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func initCLIDevice(addr string, defFile defaults.DefaultFile) device.Device {
	config := serial.Config{
		Address:  addr,
		BaudRate: defFile.CLIBaudRate,
		DataBits: defFile.CLIDataBits,
		StopBits: defFile.CLIStopBits,
		Parity:   defFile.CLIParity,
		Timeout:  1 * time.Second,
	}
	terminators := device.Terminators{
		TX: defFile.CLITXTerminator,
		RX: defFile.CLIRXTerminator,
	}

	cliDevice := device.NewDevice(config, terminators)
	return cliDevice
}

func initLogDevice(addr string, defFile defaults.DefaultFile) device.Device {
	config := serial.Config{
		Address:  addr,
		BaudRate: defFile.LogsBaudRate,
		DataBits: defFile.LogsDataBits,
		StopBits: defFile.LogsStopBits,
		Parity:   defFile.LogsParity,
		Timeout:  1 * time.Second,
	}
	terminators := device.Terminators{
		TX: defFile.LogsTXTerminator,
		RX: defFile.LogsRXTerminator,
	}

	logsDevice := device.NewDevice(config, terminators)
	return logsDevice
}

func init() {
	tuiCmd.Flags().StringVarP(&CLI_Port_Address, "cli", "c", "", "-cli /dev/tty.usb123")
	tuiCmd.Flags().StringVarP(&Logs_Port_Address, "logs", "l", "", "-logs /dev/tty.usb123")
	rootCmd.AddCommand(tuiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tuiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tuiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
