# Tabouli

## Information  
Tabouli is written in [Go](https://go.dev/). It's a TUI for interacting with firmware that supports a CLI. It also supports Test Automation Files that allow you to put in several commands, and it sends each one, waiting for a response after each.  
![Screenshot](/imgs/tabouli-main.png)

## Installation
We use `asdf` to manage golang installations.  
Here how to install golang using `asdf`, feel free to use whatever method you like.  
Enable golang plugin:  
`asdf plugin-add golang https://github.com/kennyp/asdf-golang.git`  
Install latest:  
`asdf install golang latest`  
Reshim:  
`asdf reshim golang`  
  
## Running From Source  
Install Dependencies:  
`go get`  

Run the TUI from Source:  
`go run main.go tui /dev/tty.usbserial-2111430`

= OR =  

Create a binary:  
`go build -o bin/tabouli main.go`  

Run the binary:  
`tabouli tui /dev/tty.usbserial-2111430`  
If you are running from a binary, make sure that you place your test files and the default file in `/bin` folder along with your binary.

## Example
![Usage](/imgs/tabouli-usage.gif)

## Command History
The TUI supports a command history. Simply hit the up arrow just like a normal terminal.

## Shortcuts
Currently the Commands and Test Files are invokable via auto-assigned shortcut. Simply make sure that the appropriate window has focus for the shortcut to work.  
The shortcut is desigated in yellow next to a command or automated test like: `(1)` 

## Test Automation Files  
The application will look resursively for test files starting with the current executable diretory.   
Test files must start with `test_` and be of file type `.yaml`. For instance: `test_my_quick_test.yaml` will be picked up.  
The contents of the file should look something like this:  

```
commands:
  - help
  - comm_test
  - gpio_set -pin 3 -level 1
  - ble_start_scan -whitelist 0
  - led -which 1 -color red
```  
The software will parse the file and send each line (without the "-", that is for YAML file syntax).  

## Defaults File  
Please update the `defaults.yaml` file to adjust your serial port connection settings:  
- Baud Rate
- Data Bits
- Stop Bits
- Pairity
You may also need to adjust the terminators/delimiters for the end of a command sent to the device, and the end of the response from the device (`\n` or `\r\n` or whatever your's might be). These MUST be in double quotes in the `defaults.yaml` file like: `"\r\n"`.

## Firmware Requirements 
Currently in order to be able to populate the "Commands" window automatically, there is an expected format that the firmware should send back the "help" response, here is how we handle it in our firmware:  
  
    cli_lib.println("`help`                 Get help/usage for commands\n");
    cli_lib.println("`comm_test`            Request communications test\n");
    cli_lib.println("`switch_settings`      Get the user settings\r\n");  // Notice the last command ends in \r\n

So each command has backticks around the command, and the description is outside of the command.  
This is not required, it just fills out the "Commands" column. Typing in commands and Test Automation will still work just fine.

## Future Work   
- Support a "wait" or "sleep" command in the Test Automation file syntax
- Support a "headless mode" for just invoking test automation - not everyone wants a TUI.
