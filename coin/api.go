package main

import (
	"fmt"
	"os/exec"
)

// btcctlConfPath is the path to the btcctl configuration file
const btcctlConfPath = "C:/Users/ANJALI JADEJA/$GOPATH/orcacointest/cmd/btcctl/sample-btcctl.conf"

// Commands
const (
	getBalanceCommand       = "getbalance"
	sendToAddressCommand    = "sendtoaddress"
	generateCommand         = "generate"
	walletPassphraseCommand = "walletpassphrase"
)

func main() {
	args := []string{getBalanceCommand}
	apiCall(args)
}

func apiCall(args []string) {
	argsLen := len(args)
	if argsLen == 0 {
		fmt.Println("No commands were provided")
		return
	}

	for argIdx := 0; argIdx < argsLen; argIdx++ {
		arg := args[argIdx]
		switch arg {
		case getBalanceCommand:
			err := executeCommand(getBalanceCommand)
			if err != nil {
				fmt.Println("Error when running 'getbalance':", err)
				return
			}
		case sendToAddressCommand, generateCommand, walletPassphraseCommand:
			if len(args[argIdx:]) < 2 {
				fmt.Printf("Not enough arguments for the command '%s'\n", arg)
				return
			}

			err := executeCommand(arg, args[argIdx+1:]...)
			if err != nil {
				fmt.Printf("Error running '%s': %v\n", arg, err)
				return
			}
			argIdx++
		}
	}
}

// executeCommand executes the given command with the provided arguments
func executeCommand(command string, args ...string) error {
	cmd := exec.Command("btcctl", append([]string{"--configfile=" + btcctlConfPath, command}, args...)...)
	cmd.Args = append(cmd.Args, "--notls")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running '%s': %w", command, err)
	}
	fmt.Println(string(output))
	return nil
}
