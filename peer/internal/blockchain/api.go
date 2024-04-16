package blockchain

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// btcctlConfPath is the path to the btcctl configuration file
// const btcctlConfPath = filepath.Join(currentUser.HomeDir, "$GOPATH", "orcacointest", "cmd", "btcctl", "sample-btcctl.conf")

// Commands
const (
	getBalanceCommand       = "getbalance"
	sendToAddressCommand    = "sendtoaddress"
	generateCommand         = "generate"
	walletPassphraseCommand = "walletpassphrase"
)

func StartBitcoinNode() {
	filepath, err := getConfFilePath()
	if err != nil {
		fmt.Println("Error when computing conf path", err)
		return
	}
	fmt.Println("Current path to bitcoin config file" + filepath)
	exec.Command("btcd", "--freshnet")
	// args := []string{getBalanceCommand}
	// apiCall(args)
}

func apiCall(confPath string, args []string) {
	argsLen := len(args)
	if argsLen == 0 {
		fmt.Println("No commands were provided")
		return
	}

	for argIdx := 0; argIdx < argsLen; argIdx++ {
		arg := args[argIdx]
		switch arg {
		case getBalanceCommand:
			err := executeCommand(confPath, getBalanceCommand)
			if err != nil {
				fmt.Println("Error when running 'getbalance':", err)
				return
			}
		case sendToAddressCommand, generateCommand, walletPassphraseCommand:
			if len(args[argIdx:]) < 2 {
				fmt.Printf("Not enough arguments for the command '%s'\n", arg)
				return
			}

			err := executeCommand(confPath, arg, args[argIdx+1:]...)
			if err != nil {
				fmt.Printf("Error running '%s': %v\n", arg, err)
				return
			}
			argIdx++
		}
	}
}

// executeCommand executes the given command with the provided arguments
func executeCommand(confPath string, command string, args ...string) error {
	cmd := exec.Command("btcctl", append([]string{"--configfile=" + confPath, command}, args...)...)
	cmd.Args = append(cmd.Args, "--notls")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running '%s': %w", command, err)
	}
	fmt.Println(string(output))
	return nil
}

// Retrieves the user's filepath for sample-btcctl.conf
func getConfFilePath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Get the value of the $GOPATH environment variable
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", fmt.Errorf("$GOPATH is not set")
	}

	// Replace backslashes with forward slashes (if running on Windows)
	goPath = strings.ReplaceAll(goPath, "\\", "/")

	// Construct the file path with the username and $GOPATH
	filePath := fmt.Sprintf("%s/orcacoin-go/cmd/btcctl/sample-btcctl.conf", currentUser.HomeDir)

	return filePath, nil
}
