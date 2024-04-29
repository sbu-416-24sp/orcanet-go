package blockchain

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	orcaNetPath    string = "./OrcaNet/OrcaNet"
	btcctlPath     string = "./OrcaNet/cmd/btcctl/btcctl"
	orcaWalletPath string = "./OrcaWallet/btcwallet"
)

var cmdProcess *exec.Cmd

func printOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading stream: %v\n", err)
	}
}

// startOrcaWallet: starts the OrcaWallet
func StartOrcaWallet() (*exec.Cmd, error) {
	// check for the existence of the executable
	_, err := os.Stat(orcaWalletPath)
	if os.IsNotExist(err) {
		fmt.Println("Cannot find Orcawallet executable")
		return nil, err
	}

	cmd := exec.Command(orcaWalletPath)
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		fmt.Println("failed to start wallet executable")
		return nil, err
	}
	fmt.Println("Wallet started successfully")
	return cmd, err
}

// getBtcdConfFilePath returns the file path for btcd.conf based on the user's OS
func getBtcdConfFilePath() string {
	const defaultConfigFilename = "btcd.conf"
	homeDir := getUserHomeDir()
	if homeDir == "" {
		return "" // Return empty string if the home directory can't be determined
	}

	// Determine the application data directory based on the operating system
	var appDataDir string
	switch runtime.GOOS {
	case "windows":
		appDataDir = filepath.Join(homeDir, "AppData", "Roaming", "Btcd")
	case "darwin": // macOS
		appDataDir = filepath.Join(homeDir, "Library", "Application Support", "Btcd")
	case "linux":
		appDataDir = filepath.Join(homeDir, ".btcd")
	default:
		appDataDir = filepath.Join(homeDir, ".btcd") // Default to a Unix-style hidden directory
	}

	return filepath.Join(appDataDir, defaultConfigFilename)
}

// getUserHomeDir returns the home directory of the current user
func getUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return ""
	}
	return homeDir
}

// returns an array [rpcuser, rpcpass]
func readRPCInfo(path string) ([]string, error) {
	body, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error reading the btcd.conf file")
	}

	content := string(body)
	var rpcInfo []string
	// find the line with "rpcuser" and "rpcpass"
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "rpcuser") || strings.HasPrefix(line, "rpcpass") {
			parts := strings.Split(line, "=")
			// if len(parts) == 2 {
			// 	fmt.Println(parts)
			// 	rpcInfo = append(rpcInfo, strings.TrimSpace(parts[1]))
			// }
			rpcInfo = append(rpcInfo, strings.TrimSpace(parts[1]))
		}
	}

	if len(rpcInfo) < 2 {
		return nil, fmt.Errorf("error finding rpcuser and rpc pass")
	}
	return rpcInfo, nil
}

// callBtcctlCmd: calls a Btcctl command Exactly as specified in string param, and returns the stdout of btcctl as a string
// its a singular string, but you can pass as many arguments, we will split the arguments in this fn
func CallBtcctlCmd(cmdStr string) (string, error) {
	// get the rpc values
	rpcInfo, err := readRPCInfo(getBtcdConfFilePath())
	if err != nil {
		return "", fmt.Errorf("failed to get rpc info")
	}
	fmt.Println(rpcInfo)
	params := strings.Split(cmdStr, " ")
	params = append(params, "--rpcuser="+strings.TrimSpace(rpcInfo[0])+"=", "--rpcpass="+strings.TrimSpace(rpcInfo[1])+"=")

	fmt.Println(params)
	cmd := exec.Command(btcctlPath, params...)
	// get the stdout of cmd, CAN HANG but shouldn't be a problem in a btcctl command
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute btcctl commands '%s': %s, error: %v", cmdStr, stdout, err)
	}
	fmt.Println(err)
	return string(stdout), nil
}

func unlockWallet(walletPass string) error {
	command := fmt.Sprintf("--wallet walletpassphrase %s 100", walletPass)
	stdout, err := CallBtcctlCmd(command)
	if err != nil {
		return fmt.Errorf("failed to unlock wallet: %s, error: %v", stdout, err)
	}
	return nil
}

func sendCoins(numCoins string, address string) error {
	command := fmt.Sprintf("--wallet sendtoaddress %s %s", address, numCoins)
	stdout, err := CallBtcctlCmd(command)
	if err != nil {
		return fmt.Errorf("failed to send coins: %s, error: %v", stdout, err)
	}
	return nil
}

// sendToAddress: endpoint to send n coins to an address
// if you want to send coins to a specific wallet, ask the recepient to getNewAddress and pass that address to the query string
// Usage: make a JSON request with 2 fields "coins" and "address"
func SendToAddress(coins string, address string, senderWalletPass string) error {
	if coins == "" || address == "" || senderWalletPass == "" {
		return errors.New("missing parameter")
	}

	if _, err := strconv.ParseFloat(coins, 64); err != nil {
		return errors.New("invalid coin amount")
	}

	if err := unlockWallet(senderWalletPass); err != nil {
		return errors.New("unable to unlock wallet")
	}

	if err := sendCoins(coins, address); err != nil {
		return errors.New("unable to send coins")
	}

	return nil
}
