package cli

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	orcaBlockchain "orca-peer/internal/blockchain"
	orcaClient "orca-peer/internal/client"
	"orca-peer/internal/fileshare"
	orcaHash "orca-peer/internal/hash"
	"orca-peer/internal/server"
	orcaServer "orca-peer/internal/server"

	orcaStatus "orca-peer/internal/status"
	orcaStore "orca-peer/internal/store"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	ip     string
	Client *orcaClient.Client
)

func StartCLI(bootstrapAddress *string, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey, orcaNetAPIProc *exec.Cmd, startAPIRoutes func(*map[string]fileshare.FileInfo)) {
	fmt.Println("Loading...")
	rpcPort := getPort("Market RPC Server")
	dhtPort := getPort("Market DHT Host")
	httpPort := getPort("HTTP Server")
	passKey := getPassKey()
	for httpPort == "" || rpcPort == "" || dhtPort == "" || passKey == "" {
		fmt.Println("All three ports must be given, please try again.")
		rpcPort = getPort("Market RPC Server")
		dhtPort = getPort("Market DHT Host")
		httpPort = getPort("HTTP Server")
		passKey = getPassKey()
	}
	serverReady := make(chan bool)
	confirming := false
	confirmation := ""
	locationJsonString := orcaStatus.GetLocationData()
	var locationJson map[string]interface{}
	err := json.Unmarshal([]byte(locationJsonString), &locationJson)
	if err != nil {
		fmt.Println("Unable to establish user IP, please try again")
		return
	}
	ip = locationJson["ip"].(string)
	Client = orcaClient.NewClient("files/names/")
	Client.PrivateKey = privKey
	Client.PublicKey = pubKey
	go orcaServer.StartServer(httpPort, dhtPort, rpcPort, serverReady, &confirming, &confirmation, privKey, passKey, Client, startAPIRoutes)
	<-serverReady
	orcaBlockchain.InitBlockchainStats(pubKey)
	fmt.Println("Welcome to Orcanet!")
	fmt.Println("Dive In and Explore! Type 'help' for available commands.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading from stdin:", err)
			continue
		}

		text = strings.TrimSpace(text)
		parts := strings.Fields(text)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		if confirming {
			switch command {
			case "yes":
				confirmation = "yes"
			default:
				confirmation = "no"
			}
			continue
		}

		switch command {
		case "get":
			if len(args) == 1 {
				holders, err := server.SetupCheckHolders(args[0])
				if err != nil {
					fmt.Printf("Error finding holders for file: %x", err)
					continue
				}
				var bestHolder *fileshare.User
				bestHolder = nil
				for _, holder := range holders.Holders {
					if bestHolder == nil {
						bestHolder = holder
					} else if holder.GetPrice() < bestHolder.GetPrice() {
						bestHolder = holder
					}
				}
				if bestHolder == nil {
					fmt.Println("Unable to find holder for this hash.")
					continue
				}
				fmt.Printf("%s - %d OrcaCoin\n", bestHolder.GetIp(), bestHolder.GetPrice())
				pubKeyInterface, err := x509.ParsePKIXPublicKey(bestHolder.Id)
				if err != nil {
					log.Fatal("failed to parse DER encoded public key: ", err)
				}
				rsaPubKey, ok := pubKeyInterface.(*rsa.PublicKey)
				if !ok {
					log.Fatal("not an RSA public key")
				}
				key := orcaServer.ConvertKeyToString(rsaPubKey.N, rsaPubKey.E)
				err = Client.GetFileOnce(bestHolder.GetIp(), bestHolder.GetPort(), args[0], key, fmt.Sprintf("%d", bestHolder.GetPrice()), passKey, "")
				if err != nil {
					fmt.Printf("Error getting file %s", err)
				}
			} else {
				fmt.Println("Usage: get [fileHash]")
			}
		case "store":
			if len(args) == 2 {
				fileName := args[0]
				filePath := "./files/" + fileName
				if _, err := os.Stat(filePath); err == nil {

				} else if os.IsNotExist(err) {
					fmt.Println("file does not exist inside files folder")
					continue
				} else {
					fmt.Println("error checking file's existence, please try again")
					continue
				}
				costPerMB, err := strconv.ParseInt(args[1], 10, 64)
				if err != nil {
					fmt.Println("Error parsing in cost per MB: must be a int64", err)
					continue
				}
				port, err := strconv.ParseInt(httpPort, 10, 64)
				if err != nil {
					fmt.Println("Error parsing in port: must be a integer.", err)
					continue
				}
				err = server.SetupRegisterFile(filePath, fileName, costPerMB, ip, int32(port))
				if err != nil {
					fmt.Printf("Unable to register file on DHT: %s", err)
				} else {
					fmt.Println("Sucessfully registered file on DHT.")
				}
			} else {
				fmt.Println("Usage: store [fileName] [amount]")
			}
		case "import":
			if len(args) == 1 {
				err := Client.ImportFile(args[0])
				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Usage: import [filepath]")
			}
		case "location":
			fmt.Println(orcaStatus.GetLocationData())
		case "network":
			fmt.Println("Testing Network Speeds...")
			networkData := orcaStatus.GetNetworkInfo()
			if networkData.Success {
				fmt.Printf("Latency: %fms, Download: %fMbps, Upload: %fMbps\n", networkData.LatencyMs, networkData.DownloadSpeedMbps, networkData.UploadSpeedMbps)
			} else {
				fmt.Println("Unable to test network speeds. Please try again")
			}

		case "list":
			files := orcaStore.GetAllLocalFiles()
			fmt.Print("Files found: \n")
			for _, file := range files {
				fmt.Println(file.Name)
			}
		case "hash":
			if len(args) == 1 {
				orcaHash.HashFile(args[0])
			} else {
				fmt.Println("Usage: hash [fileName]")
			}
		case "send":
			if len(args) == 3 {
				cost, err := strconv.ParseFloat(args[0], 64)
				if err != nil {
					fmt.Println("Error parsing amount to send")
					continue
				}
				orcaClient.SendTransaction(cost, args[1], args[2], pubKey, privKey)
			} else {
				fmt.Println("Usage: send [amount] [ip] [port]")
			}
		case "exit":
			fmt.Println("Exiting...")

			for _, fileInfo := range orcaStore.GetAllLocalFiles() {
				filePath := "./files/stored/" + fileInfo.Name
				err := os.Remove(filePath)
				if err != nil {
					fmt.Printf("Error cleaning up stored files: %s\n", err)
				}
			}

			err = orcaNetAPIProc.Process.Signal(os.Interrupt)
			if err != nil {
				fmt.Printf("Error killing OrcaNet: %s\n", err)
			}

			return
		case "getdir":
			if len(args) == 3 {
				port, err := strconv.ParseInt(args[1], 10, 32)
				if err != nil {
					fmt.Printf("Invalid port: %s\n", err)
					fmt.Println()
					continue
				}

				go Client.GetDirectory(args[0], int32(port), args[2])
			} else {
				fmt.Println("Usage: getdir [ip] [port] [path]")
			}
		case "storedir":
			if len(args) == 3 {
				go Client.StoreDirectory(args[0], args[1], args[2])
			} else {
				fmt.Println("Usage: storedir [ip] [port] [path]")
			}
		case "help":
			fmt.Println("COMMANDS:")
			fmt.Println(" get [fileHash]                 Request a file from DHT")
			fmt.Println(" store [fileName] [amount]      Store a file on DHT")
			fmt.Println(" getdir [ip] [port] [path]      Request a directory")
			fmt.Println(" storedir [ip] [port] [path]    Request storage of a directory")
			fmt.Println(" import [filepath]              Import a file")
			fmt.Println(" send [amount] [ip]             Send an amount of money to network")
			fmt.Println(" hash [fileName]                Get the hash of a file")
			fmt.Println(" list                           List all files you are storing")
			fmt.Println(" location                       Print your location")
			fmt.Println(" network                        Test speed of network")
			fmt.Println(" exit                           Exit the program")
		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
		}
	}
}

// Ask user to enter a port and returns it
func getPort(useCase string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter a port number to start listening to requests for %s: ", useCase)
	for {
		port, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			os.Exit(1)
		}
		port = strings.TrimSpace(port)

		// Validate port
		listener, err := net.Listen("tcp", ":"+port)
		if err == nil {
			defer listener.Close()
			return port
		}

		fmt.Print("Invalid port. Please enter a different port: ")
	}
}

func getPassKey() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter your blockchain wallet passkey: ")
	passKey, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	passKey = strings.TrimSpace(passKey)
	return passKey
}

// detectNAT simulates the process of detecting whether the node is behind NAT.
func detectNAT() bool {
	ipapiClient := http.Client{}

	ipv4Req, err := http.NewRequest("GET", "http://httpbin.org/ip", nil)
	if err != nil {
		fmt.Println("Error creating IPv4 request:", err)
		os.Exit(1)
	}
	resp, err := ipapiClient.Do(ipv4Req)
	if err != nil {
		fmt.Println("Error retrieving IPv4:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading IPv4 response body:", err)
		os.Exit(1)
	}

	var ipv4JSON struct {
		Origin string `json:"origin"`
	}
	err = json.Unmarshal(body, &ipv4JSON)
	if err != nil {
		fmt.Println("Error unmarshalling IPv4 response body:", err)
		os.Exit(1)
	}

	publicIP := net.ParseIP(ipv4JSON.Origin)

	// Define private IP address ranges.
	privateRanges := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
	}

	// Check if the public IP address is within any of the private IP address ranges.
	for _, pr := range privateRanges {
		if pr.Contains(publicIP) {
			return false
		}
	}
	return true
}
