package tui

import (
	"errors"
	"fmt"
	"strings"
	"tabouli/internal/device"
	"tabouli/internal/test_automation"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app                  *tview.Application
	cmdBox               *tview.InputField
	cmdHistoryBox        *tview.TextView
	logsBox              *tview.TextView
	cmdListBox           *tview.List
	testFilesListBox     *tview.List
	cmdInputHistory      []string
	cmdInputHistoryIndex int
)

func LogToUI(b byte) {
	s := []byte{b}
	logsBox.Write(s)
	app.ForceDraw()
}

func CreateTView(device device.Device, deviceLogs device.Device) *tview.Application {
	app = tview.NewApplication()
	cmdInputHistory = make([]string, 0)

	cmdBox = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetPlaceholderStyle(tcell.StyleDefault.Blink(true)).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				cmd := cmdBox.GetText()
				logToHistory(fmt.Sprintf(">> %s", cmd))
				cmdBox.SetText("")
				cmdInputHistory = append(cmdInputHistory, cmd)
				cmdInputHistoryIndex = len(cmdInputHistory)
				var response, err = device.TXcmdRXresponse(cmd)
				if err != nil {
					logToHistory(fmt.Sprintf("error reading from device %s\n", err))
				} else {
					logToHistory(fmt.Sprint(response))
				}
			}
		}).
		SetChangedFunc(func(text string) {
			if len(text) == 0 {
				cmdBox.SetFieldBackgroundColor(tcell.ColorBlack)
				return
			}
			for _, v := range device.Commands {
				// Get first portion of what the user has typed in
				// We do this in case a command has parameters
				typedCmd := strings.Split(text, " ")
				supportedCmds := strings.Split(v.CmdText, " ")
				if strings.HasPrefix(typedCmd[0], supportedCmds[0]) {
					cmdBox.SetFieldBackgroundColor(tcell.ColorGreen)
					return
				}
			}
			cmdBox.SetFieldBackgroundColor(tcell.ColorRed)
		})
	cmdBox.SetBorder(true).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if len(cmdInputHistory) == 0 {
			return event
		}
		switch event.Key() {
		case tcell.KeyUp:
			if cmdInputHistoryIndex == 0 {
				cmdBox.SetText(cmdInputHistory[cmdInputHistoryIndex])
				break
			}
			cmdInputHistoryIndex--
			cmdBox.SetText(cmdInputHistory[cmdInputHistoryIndex])

		case tcell.KeyDown:
			if cmdInputHistoryIndex == len(cmdInputHistory)-1 {
				cmdBox.SetText("")
				break
			}
			cmdInputHistoryIndex++
			cmdBox.SetText(cmdInputHistory[cmdInputHistoryIndex])
		default:
			return event
		}
		return nil
	})

	// First check for test automation files so its available when creating callback function
	var automator, automator_err = test_automation.NewTestFileAutomator()

	// Good example of how rows and columns work:
	// https://github.com/rivo/tview/blob/master/demos/grid/main.go

	// This is the main UI that holds everything
	mainGrid := tview.NewGrid().
		SetRows(0).
		SetColumns(40, 0, 40) // This defines 3 columns where the left most and right most are 40 cells

	// This is the main "log" of all commands and responses
	cmdHistoryBox = tview.NewTextView()
	cmdHistoryBox.SetTitle(" CLI History ").SetBorder(true)
	historyAndCmdInputGrid := tview.NewGrid().SetRows(0, 3).SetColumns(0)
	historyAndCmdInputGrid.
		AddItem(cmdHistoryBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(cmdBox, 1, 0, 1, 1, 0, 0, false)
	mainGrid.AddItem(historyAndCmdInputGrid, 0, 1, 1, 1, 0, 0, false)

	// Build UI differently if there is a second serial connection for logs
	noLogs := false
	if noLogs {
		// This is the left most column
		cmdListBox = tview.NewList().
			SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
				// Send the command
				logToHistory(fmt.Sprintf(">> %s", main_text))
				var response, err = device.TXcmdRXresponse(main_text)
				if err != nil {
					logToHistory(fmt.Sprintf("error reading from device %s\n", err))
				} else {
					logToHistory(fmt.Sprint(response))
				}
			})
		cmdListBox.SetTitle(" CLI Commands ").SetBorder(true)
		mainGrid.AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false)

		// This is the list of test files
		testFilesListBox = tview.NewList().
			SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
				// It would be good if we had already processed and had an object to read from
				// Then we know how many tests we have
				if automator_err == nil {
					var test_err = runAutomatedTest(automator, index, device)
					if test_err != nil {
						logToHistory(fmt.Sprintf("%v", test_err))
					}
				}

			})
		testFilesListBox.SetTitle(" Test Files ").
			SetBorder(true)
		mainGrid.AddItem(testFilesListBox, 0, 2, 1, 1, 0, 0, false)

		// Load up the Test File Automation
		logToHistory("Checking for Test Automation files...")
		if automator_err != nil {
			logToHistory(fmt.Sprintf("%v", automator_err))
		} else {
			logToHistory("Test Automation Files Detected!")
			for num, testFile := range automator.TestFiles {
				testFilesListBox.AddItem(testFile.TestFileName, "", getShortcutRuneForCount(num), nil)
			}
		}
	} else {

		// This is the left most column
		// CLI Command List
		cmdListBox = tview.NewList().
			SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
				// Send the command
				logToHistory(fmt.Sprintf(">> %s", main_text))
				var response, err = device.TXcmdRXresponse(main_text)
				if err != nil {
					logToHistory(fmt.Sprintf("error reading from device %s\n", err))
				} else {
					logToHistory(fmt.Sprint(response))
				}
			})
		cmdListBox.SetTitle(" CLI Commands ").SetBorder(true)

		// This is the list of test files
		testFilesListBox = tview.NewList().
			SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
				// It would be good if we had already processed and had an object to read from
				// Then we know how many tests we have
				if automator_err == nil {
					var test_err = runAutomatedTest(automator, index, device)
					if test_err != nil {
						logToHistory(fmt.Sprintf("%v", test_err))
					}
				}

			})
		testFilesListBox.SetTitle(" Test Files ").
			SetBorder(true)
		// Load up the Test File Automation
		logToHistory("Checking for Test Automation files...")
		if automator_err != nil {
			logToHistory(fmt.Sprintf("%v", automator_err))
		} else {
			logToHistory("Test Automation Files Detected!")
			for num, testFile := range automator.TestFiles {
				testFilesListBox.AddItem(testFile.TestFileName, "", getShortcutRuneForCount(num), nil)
			}
		}

		cmdAndTestFileGrid := tview.NewGrid().SetRows(0, 2).SetColumns(0)
		cmdAndTestFileGrid.
			AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false).
			AddItem(testFilesListBox, 1, 0, 2, 1, 0, 0, false)

		mainGrid.AddItem(cmdAndTestFileGrid, 0, 0, 1, 1, 0, 0, false)

		// Right-most column, logs
		logsBox = tview.NewTextView()
		logsBox.SetTitle(" Logs ").SetBorder(true)
		mainGrid.AddItem(logsBox, 0, 2, 1, 1, 0, 0, false)

		mainGrid.SetColumns(40, 80, 0)
	}

	// Attempt to connect to com port
	logToHistory("Connecting to COM Port...")
	if err := device.Open(); err != nil {
		fmt.Print(err)
		app.Stop()
	}
	logToHistory("Connected!")
	logToHistory("Loading commands...")

	// Attempt to connect to com port
	// logToDeviceLogs("Connecting to COM Port...")
	// if err := deviceLogs.Open(); err != nil {
	// 	fmt.Print(err)
	// 	app.Stop()
	// }
	// logToDeviceLogs("Connected!")

	// Load the commands from the device
	// TODO: check for errors
	device.LoadCmds()
	for num, cmd := range device.Commands {
		cmdListBox.AddItem(cmd.CmdText, cmd.Description, getShortcutRuneForCount(num), nil)
	}
	if len(device.Commands) > 0 {
		logToHistory("Commands loaded!")
	} else {
		logToHistory("No commands recevied and parsed when 'help' command sent. \nCheck your response syntax or connection.")
	}

	logToHistory("\n")

	// Now run all the setup, connections, etc.
	// TODO should we move this out of SetFocus??
	mainGrid.SetFocusFunc(func() {
		app.SetFocus(cmdBox)
	})

	app.SetRoot(mainGrid, true).EnableMouse(true)
	return app
}

func logToDeviceLogs(msg string) {
	fmt.Fprintln(logsBox, msg)
}

func logToHistory(msg string) {
	fmt.Fprintln(cmdHistoryBox, msg)
}

func getShortcutRuneForCount(count int) rune {
	// Shortcuts take a rune, I have not found a clean way to convert and handle how we want.

	// ascii 0 - 9
	if count >= 0 && count < 10 {
		return rune(0x30 + count)
	}

	// ascii a - z
	if count >= 10 && count < 37 {
		return rune(0x61 + (count - 10))
	}

	return rune(0x21 + (count - 37))
}

func runAutomatedTest(automator test_automation.TestFileAutomator, testIndex int, device device.Device) error {

	if automator.TestInProgress {
		return errors.New("test in progress, please wait")
	}

	if testIndex >= len(automator.TestFiles) {
		return errors.New("index of test is out of bounds")
	}

	automator.TestInProgress = true

	var testFile = automator.TestFiles[testIndex]
	var testSteps = len(testFile.CommandList)
	logToHistory(fmt.Sprintf("\n=== Starting Automated Test: %s with %d commands ===", testFile.TestFileName, testSteps))
	for _, cmd := range testFile.CommandList {
		logToHistory(fmt.Sprintf(">> %s", cmd))
		var response, err = device.TXcmdRXresponse(cmd)
		if err != nil {
			logToHistory(fmt.Sprintf("error reading from device %s\n", err))
		} else {
			logToHistory(fmt.Sprint(response))
		}
	}
	logToHistory(fmt.Sprintln("\n=== Test Complete ==="))

	automator.TestInProgress = false
	return nil
}
