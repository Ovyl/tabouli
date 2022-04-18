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
	footer               *tview.Flex
	footerOptions        *tview.TextView
	cmdInputHistory      []string
	cmdInputHistoryIndex int
	cliDevice            device.Device
	logsDevice           device.Device
)

// Invoked by the CLI device RX byte routine
func LogToUI(b byte) {
	s := []byte{b}
	logsBox.Write(s)
	app.ForceDraw()
}

// Builds the Command Box (where you type in commands) UI element
func uiBuildCmdBox() {
	cmdBox = tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetPlaceholderStyle(tcell.StyleDefault.Blink(true))
	cmdBox.SetBorder(true)
	cmdBox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
	cmdBox.SetDoneFunc(func(key tcell.Key) {
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
	}).SetChangedFunc(func(text string) {
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
}

// Builds the Command Hisory List UI element
func uiBuildCmdHistory() {
	// This is the main "log" of all commands and responses
	cmdHistoryBox = tview.NewTextView()
	cmdHistoryBox.SetTitle(" CLI History ").SetBorder(true)
}

func uiBuildCmdListBox() {
	// This is the left most column
	cmdListBox = tview.NewList()
	cmdListBox.SetTitle(" CLI Commands ").SetBorder(true)
	cmdListBox.SetSelectedFunc(func(index int, main_text string, sec_text string, r rune) {
		// Send the command
		logToHistory(fmt.Sprintf(">> %s", main_text))
		var response, err = cliDevice.TXcmdRXresponse(main_text)
		if err != nil {
			logToHistory(fmt.Sprintf("error reading from device %s\n", err))
		} else {
			logToHistory(fmt.Sprint(response))
		}
	})
}

func uiBuildHistoryAndCmdBoxGrid(mainGrid *tview.Grid) {
	historyAndCmdInputGrid := tview.NewGrid().SetRows(0, 3).SetColumns(0)
	historyAndCmdInputGrid.
		AddItem(cmdHistoryBox, 0, 0, 1, 1, 0, 0, false).
		AddItem(cmdBox, 1, 0, 1, 1, 0, 0, false)
	mainGrid.AddItem(historyAndCmdInputGrid, 0, 1, 1, 1, 0, 0, false)
}

// Builds the Test Automation File list UI element
func uiBuildTestAutomationFileListBox(automator test_automation.TestFileAutomator, filesPresent bool) {
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
func uiBuildDeviceLogsWithCLI() {
	logsBox = tview.NewTextView()
	logsBox.SetTitle(" Device Logs ").SetBorder(true)
	logToLogs("Connected to Log COM Port.\n\n")
}

func uiBuildDeviceLogs(mainGrid *tview.Grid) {
	mainGrid.SetColumns(0)
	mainGrid.SetRows(0, 2)

	logsBox = tview.NewTextView()
	logsBox.SetFocusFunc(func() {
		app.SetFocus(footerOptions)
	})
	logsBox.SetTitle(" Device Logs ").SetBorder(true)
	mainGrid.AddItem(logsBox, 0, 0, 1, 1, 0, 0, false)
	logToLogs("Connected to Log COM Port.\n\n")

	footerOptions = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
	footerOptions.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// This is Shift + C
		if event.Rune() == rune(67) {
			logsBox.Clear()
		}
		return event
	})
	optionText := fmt.Sprintf(`  ["%d"] %s[""] `, 0, "Shift + C: Clear Logs")
	footerOptions.Highlight(optionText)
	fmt.Fprint(footerOptions, optionText)
	footer = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(footerOptions, 0, 1, false)

	mainGrid.AddItem(footer, 1, 0, 1, 1, 0, 0, false)
	app.SetFocus(footer)
}

// Good example of how rows and columns work:
// https://github.com/rivo/tview/blob/master/demos/grid/main.go
// TODO: If no devices are passed, this should not crash, it should fail gracefully
func CreateTView(cli device.Device, logs device.Device) *tview.Application {

	cliDevice = cli
	logsDevice = logs

	app = tview.NewApplication()
	cmdInputHistory = make([]string, 0)

	// First check for test automation files so its available when creating callback function
	var automator, automator_err = test_automation.NewTestFileAutomator()

	// This is the main UI that holds everything
	mainGrid := tview.NewGrid().
		SetRows(0)

	// To DRY up the code, this is common if the CLI is open
	if cliDevice.IsOpen {

		// Build the UI elements for the CLI
		uiBuildCmdBox()
		uiBuildCmdListBox()
		uiBuildCmdHistory()
		uiBuildHistoryAndCmdBoxGrid(mainGrid)
		uiBuildTestAutomationFileListBox(automator, (automator_err == nil))

		// Attemptt to load the commands from the help command response
		loadCLICmds(&cliDevice)
	}

	// There are 3 different UI's: CLI + Logs, CLI + no Logs, no CLI + Logs
	if cliDevice.IsOpen && logsDevice.IsOpen {
		// This will be the UI with columns: | Commands + Test Automation | Command History | Logs |

		// Build the UI elements for the Logs + CLI
		cmdAndTestFileGrid := tview.NewGrid().SetRows(0, 2).SetColumns(0)
		cmdAndTestFileGrid.
			AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false).
			AddItem(testFilesListBox, 1, 0, 2, 1, 0, 0, false)
		mainGrid.AddItem(cmdAndTestFileGrid, 0, 0, 1, 1, 0, 0, false)
		uiBuildDeviceLogsWithCLI()

		mainGrid.AddItem(logsBox, 0, 2, 1, 1, 0, 0, false)
		mainGrid.SetColumns(40, 100, 0)
	} else if cliDevice.IsOpen && !logsDevice.IsOpen {
		// This will be the UI with columns: | Commands | Command History | Test Automation |
		// Set the CLI element locations
		mainGrid.AddItem(cmdListBox, 0, 0, 1, 1, 0, 0, false)
		mainGrid.AddItem(testFilesListBox, 0, 2, 1, 1, 0, 0, false)
		mainGrid.SetColumns(40, 0, 40)
	} else {
		// This will be the UI with columns: | Logs |
		uiBuildDeviceLogs(mainGrid)
	}

	// Now run all the setup, connections, etc.
	mainGrid.SetFocusFunc(func() {
		if cliDevice.IsOpen {
			app.SetFocus(cmdBox)
		}
	})

	app.SetRoot(mainGrid, true).EnableMouse(true)
	return app
}

func loadCLICmds(cliDevice *device.Device) {
	// If the CLI device is open, attempt to load commands
	var count = 0
	if cliDevice.IsOpen {
		logToHistory("Connected to CLI COM Port")

		logToHistory("Loading commands...")
		// Load the commands from the device
		cliDevice.LoadCmds()
		for num, cmd := range cliDevice.Commands {
			cmdListBox.AddItem(cmd.CmdText, cmd.Description, getShortcutRuneForCount(num), nil)
		}
		count = len(cliDevice.Commands)
		if count > 0 {
			logToHistory(fmt.Sprintf("%d Commands loaded!\n", len(cliDevice.Commands)))
		} else {
			logToHistory("No commands recevied and parsed when 'help' command sent. \nCheck your response syntax or connection.\n")
		}
	}
}

func logToLogs(msg string) {
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
