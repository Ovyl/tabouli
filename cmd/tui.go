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

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Runs the tui app to help with the aid of creating the test files",
	Example: `tabouli tui /dev/tty.usbserial-1446120:115200
	`,
	Args: argsValidation,
	Run: func(cmd *cobra.Command, args []string) {
		//address, baudrate, _ := getPathAndBaudrate(args[0])
		config := serial.Config{
			Address:  args[0],
			BaudRate: 9600,
			DataBits: 8,
			StopBits: 1,
			Parity:   "N",
			Timeout:  1 * time.Second,
		}
		terminators := device.Terminators{}
		terminators.RX = "\r\n"
		terminators.TX = "\n"
		defaults, err := defaults.NewDefaultFileHanlder()
		if err == nil {
			config.BaudRate = defaults.SerialBaudRate
			config.DataBits = defaults.SerialDataBits
			config.StopBits = defaults.SerialStopBits
			config.Parity = defaults.SerialParity
			terminators.TX = defaults.TXTerminator
			terminators.RX = defaults.RXTerminator
		} else {
			log.Fatal(err)
		}

		cliDevice := device.NewDevice(config, terminators)

		logDeviceconfig := serial.Config{
			Address:  args[1],
			BaudRate: 115200,
			DataBits: 8,
			StopBits: 1,
			Parity:   "N",
			Timeout:  1 * time.Second,
		}
		logTerminators := device.Terminators{}
		logTerminators.RX = "\n"
		logTerminators.TX = "\n"
		logDevice := device.NewDevice(logDeviceconfig, logTerminators)

		if err := logDevice.Open(); err != nil {
			log.Fatal(err)
		}

		ui := tui.CreateTView(cliDevice, logDevice)

		go logDevice.RXLogsForever(tui.LogToUI)

		if err := ui.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tuiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tuiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
