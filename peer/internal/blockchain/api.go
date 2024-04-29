package blockchain

import (
	"fmt"
	"os"
	"os/exec"
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

var filepath string

func StartBitcoinNode() {
	filepath, err := getConfFilePath()
	if err != nil {
		fmt.Println("Error when computing conf path", err)
		return
	}
	filepath = filepath + ""
	c, b := exec.Command("../coin/btcd", "--freshnet"), new(strings.Builder)
	c.Stdout = b
	c.Run()
	print(b.String())
	out, err := exec.Command("../coin/btcd", "--freshnet").Output()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		fmt.Println(line)
	}
	// args := []string{getBalanceCommand}
	// apiCall(args)
}

func ApiCall(confPath string, args []string) {
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
	cmd := exec.Command("../coin/btcctl", append([]string{"--configfile=" + confPath, command}, args...)...)
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
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Construct the file path with the username and $GOPATH
	filePath := fmt.Sprintf("%s/../coin/cmd/btcctl/sample-btcctl.conf", currentDir)

	return filePath, nil
}
