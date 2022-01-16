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

// Invoked by the CLI device RX byte routine
func LogToUI(b byte) {
	s := []byte{b}
	logsBox.Write(s)
	app.ForceDraw()
}

// Builds the Command Box (where you type in commands) UI element
func uiBuildCmdBox(cliDevice device.Device) {
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
				var response, err = cliDevice.TXcmdRXresponse(cmd)
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
			for _, v := range cliDevice.Commands {
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
}

// Builds the Command Hisory List UI element
func uiBuildCmdHistory(mainGrid *tview.Grid) {
	// This is the main "log" of all commands and responses
	cmdHistoryBox = tview.NewTextView()
	cmdHistoryBox.SetTitle(" CLI History ").SetBorder(true)
	historyAndCmdInputGrid := tview.NewGrid().SetRows(0, 3).SetColumns(0)
	historyAndCmdInputGrid.
		AddItem(cmdHistoryBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(cmdBox, 1, 0, 1, 1, 0, 0, false)
	mainGrid.AddItem(historyAndCmdInputGrid, 0, 1, 1, 1, 0, 0, false)
}

func uiBuildCmdListBox(cliDevice device.Device) {
	// This is the left most column
	cmdListBox = tview.NewList().
		SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
			// Send the command
			logToHistory(fmt.Sprintf(">> %s", main_text))
			var response, err = cliDevice.TXcmdRXresponse(main_text)
			if err != nil {
				logToHistory(fmt.Sprintf("error reading from device %s\n", err))
			} else {
				logToHistory(fmt.Sprint(response))
			}
		})
	cmdListBox.SetTitle(" CLI Commands ").SetBorder(true)
}

// Builds the Test Automation File list UI element
func uiBuildTestAutomationFileListBox(cliDevice device.Device, automator test_automation.TestFileAutomator, filesPresent bool) {
	testFilesListBox = tview.NewList().
		SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
			// It would be good if we had already processed and had an object to read from
			// Then we know how many tests we have
			if filesPresent {
				var test_err = runAutomatedTest(automator, index, cliDevice)
				if test_err != nil {
					logToHistory(fmt.Sprintf("%v", test_err))
				}
			}

		})
	testFilesListBox.SetTitle(" Test Files ").
		SetBorder(true)

	var logMessage = "No Test Automation Files detected"
	if filesPresent {
		logMessage = "Test Automation Files Detected!"
		for num, testFile := range automator.TestFiles {
			testFilesListBox.AddItem(testFile.TestFileName, "", getShortcutRuneForCount(num), nil)
		}
	}
	if cliDevice.IsOpen {
		logToHistory(logMessage)
	}
}

// Builds the device logs UI element, this is only built if a log device has been passed in and opened
func uiBuildDeviceLogs() {
	logsBox = tview.NewTextView()
	logsBox.SetTitle(" Logs ").SetBorder(true)
}

// Good example of how rows and columns work:
// https://github.com/rivo/tview/blob/master/demos/grid/main.go
// TODO: If no devices are passed, this should not crash, it should fail gracefully
func CreateTView(cliDevice device.Device, logDevice device.Device) *tview.Application {

	app = tview.NewApplication()
	cmdInputHistory = make([]string, 0)

	// First check for test automation files so its available when creating callback function
	var automator, automator_err = test_automation.NewTestFileAutomator()

	// This is the main UI that holds everything
	mainGrid := tview.NewGrid().
		SetRows(0)

	// Build the UI elements
	uiBuildCmdBox(cliDevice)
	uiBuildCmdHistory(mainGrid)
	uiBuildCmdListBox(cliDevice)
	uiBuildTestAutomationFileListBox(cliDevice, automator, (automator_err == nil))

	// Adjust the UI to account for the Logs UI element
	if logDevice.IsOpen {
		logToHistory("Connected to Logging COM Port")

		// Since there are device logs, place the Test Automation files in a grid with the Command list, left most column
		cmdAndTestFileGrid := tview.NewGrid().SetRows(0, 2).SetColumns(0)
		cmdAndTestFileGrid.
			AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false).
			AddItem(testFilesListBox, 1, 0, 2, 1, 0, 0, false)
		mainGrid.AddItem(cmdAndTestFileGrid, 0, 0, 1, 1, 0, 0, false)

		// Right-most column, logs
		uiBuildDeviceLogs()
		mainGrid.AddItem(logsBox, 0, 2, 1, 1, 0, 0, false)

		mainGrid.SetColumns(40, 80, 0)
	} else {
		mainGrid.AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false)

		// No logs, so the list of Test Automation files is the right most column
		mainGrid.AddItem(testFilesListBox, 0, 2, 1, 1, 0, 0, false)

		mainGrid.SetColumns(40, 0, 40)
	}

	// Load up CLI commands
	loadCLICmds(cliDevice)

	// Now run all the setup, connections, etc.
	// TODO should we move this out of SetFocus??
	mainGrid.SetFocusFunc(func() {
		app.SetFocus(cmdBox)
	})

	app.SetRoot(mainGrid, true).EnableMouse(true)
	return app
}

func loadCLICmds(cliDevice device.Device) {
	// If the CLI device is open, attempt to load commands
	if cliDevice.IsOpen {
		logToHistory("Connected to CLI COM Port")

		logToHistory("Loading commands...")
		// Load the commands from the device
		cliDevice.LoadCmds()
		for num, cmd := range cliDevice.Commands {
			cmdListBox.AddItem(cmd.CmdText, cmd.Description, getShortcutRuneForCount(num), nil)
		}
		if len(cliDevice.Commands) > 0 {
			logToHistory("Commands loaded!")
		} else {
			logToHistory("No commands recevied and parsed when 'help' command sent. \nCheck your response syntax or connection.")
		}

		logToHistory("\n")
	}
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
