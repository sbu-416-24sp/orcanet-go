 package manageOrcaNet

 import (
     "fmt"
     "os"
     "os/exec"
     "strings"
     "bufio"
     "runtime"
     "path/filepath"
     "time"
     "io"
 )


// we need to get the execPath so that we can run the other executables regardless what the PWD is
func getExePath() (string, error) {
    exePath, err := os.Executable()
    if err != nil {
        fmt.Println("Could not find OrcaNetAPIServer executable")
        return "", err
    }
    return exePath, nil
}

const (
     orcaNetPath string = "./OrcaNet/OrcaNet"
     btcctlPath string = "./OrcaNet/cmd/btcctl/btcctl"
     orcaWalletPath string = "./OrcaWallet/btcwallet"
 )

var cmdProcess *exec.Cmd

func Start(params ...string) error {
    exePath, err := getExePath();
    if err != nil {
        return err;
    }
    var orcaNetFullPath = filepath.Join(exePath, "..", "..", orcaNetPath)
    _, err = os.Stat(orcaNetFullPath)
    if os.IsNotExist(err) {
        fmt.Println("Cannot find OrcaNet executable")
        return err
    }

    cmdProcess = exec.Command(orcaNetFullPath, params...)

    stdout, err := cmdProcess.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    stderr, err := cmdProcess.StderrPipe()
    if err != nil {
        return fmt.Errorf("failed to create stderr pipe: %w", err)
    }

    fmt.Println("Start OrcaNet with params: ", params)
    if err := cmdProcess.Start();  err != nil {
        fmt.Println("Failed to start OrcaNet:", err)
        return err
    }
    fmt.Println("OrcaNet started successfully")
    go printOutput(stdout)
    go printOutput(stderr)


    return nil
}


func printOutput(r io.Reader) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading stream: %v\n", err)
    }
}


// Stop: ends the running OrcaNet instance if its running
//TO DO: right now it sleeps for 5 seconds instead of just waiting for the interrupt (fix this)
func Stop() error {
    if cmdProcess == nil || cmdProcess.Process == nil {
        fmt.Println("OrcaNet process is not currently running.")
        return fmt.Errorf("OrcaNet process is not running")
    }

    fmt.Println("Stopping OrcaNet...")
    // send interrupt sig
    if err := cmdProcess.Process.Signal(os.Interrupt); err != nil {
        fmt.Println("Failed to send interrupt:", err)
        return err
    }
    time.Sleep(5 * time.Second)
    fmt.Println("OrcaNet stopped successfully.")
    return nil
}

//startOrcaWallet: starts the OrcaWallet
func StartOrcaWallet() error {
    // get path relative to executable so it can be run anywhere and not just from the location of the exe
    exePath, err := getExePath();
    if err != nil {
        return err;
    }
    var walletFullPath = filepath.Join(exePath, "..", "..", orcaWalletPath)
    // check for the existence of the executable 
    _, err = os.Stat(walletFullPath)
    if os.IsNotExist(err) {
        fmt.Println("Cannot find Orcawallet executable")
        return err
    }

    cmd := exec.Command(walletFullPath)
    if err := cmd.Start(); err != nil {
        fmt.Println(err)
        fmt.Println("failed to start wallet executable")
        return nil
    }
    fmt.Println("Wallet started successfully")
    return nil
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
	lines := strings.Split(content, "\n");
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
    // get the bttcl full path
    exePath, err := getExePath()
    if err != nil {
        fmt.Println("error finding APIServer exec")
        return "", err
    }
    
    var btcctlFullPath = filepath.Join(exePath, "..", "..", btcctlPath)
    _, err = os.Stat(btcctlFullPath)
    if os.IsNotExist(err) {
        fmt.Println("Error finding btcctl full path")
        return "", err 
    }
    params :=  strings.Split(cmdStr, " ") 
    params = append(params, "--rpcuser=" + strings.TrimSpace(rpcInfo[0]) + "=", "--rpcpass=" + strings.TrimSpace(rpcInfo[1]) + "=")

    fmt.Println(params)
    cmd := exec.Command(btcctlFullPath, params...)
    // get the stdout of cmd, CAN HANG but shouldn't be a problem in a btcctl command
    stdout, err := cmd.CombinedOutput() 
    if err != nil {
        return "", fmt.Errorf("failed to execute btcctl commands '%s': %s, error: %v", cmdStr, stdout, err)
    }
    fmt.Println(err);
    return string(stdout), nil
}


