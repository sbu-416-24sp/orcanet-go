package cli

import (
	"bufio"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net"
	orcaClient "orca-peer/internal/client"
	"orca-peer/internal/fileshare"
	orcaHash "orca-peer/internal/hash"
	"orca-peer/internal/server"
	orcaServer "orca-peer/internal/server"
	orcaStatus "orca-peer/internal/status"
	orcaStore "orca-peer/internal/store"
	"os"
	"strconv"
	"strings"
)

var (
	ip string
)

func StartCLI(bootstrapAddress *string, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) {
	fmt.Println("Loading...")
	rpcPort := getPort("Market RPC Server")
	dhtPort := getPort("Market DHT Host")
	httpPort := getPort("HTTP Server")
	for httpPort == "" || rpcPort == "" || dhtPort == "" {
		fmt.Println("All three ports must be given, please try again.")
		rpcPort = getPort("Market RPC Server")
		dhtPort = getPort("Market DHT Host")
		httpPort = getPort("HTTP Server")
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
	go orcaServer.StartServer(httpPort, dhtPort, rpcPort, serverReady, &confirming, &confirmation, privKey)
	<-serverReady
	fmt.Println("Welcome to Orcanet!")
	fmt.Println("Dive In and Explore! Type 'help' for available commands.")

	reader := bufio.NewReader(os.Stdin)
	client := orcaClient.NewClient("files/names/")

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
					} else if bestHolder.Price < holder.Price {
						bestHolder = holder
					}
				}
				if bestHolder == nil {
					fmt.Println("Unable to find holder for this hash.")
					continue
				}
				fmt.Printf("%s - %d OrcaCoin\n", bestHolder.GetIp(), bestHolder.GetPrice())
				client.GetFileOnce(string(bestHolder.Ip), string(bestHolder.Port), args[0])
			} else {
				fmt.Println("Usage: get [fileHash]")
				fmt.Println()
			}
		case "store":
			if len(args) == 2 {
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
				err = server.SetupRegisterFile(args[0], costPerMB, ip, int32(port))
				if err != nil {
					fmt.Printf("Unable to register file on DHT: %x", err)
				} else {
					fmt.Println("Sucessfully registered file on DHT.")
				}
			} else {
				fmt.Println("Usage: store [fileHash] [amount]")
				fmt.Println()
			}
		case "import":
			if len(args) == 1 {
				go client.ImportFile(args[0])
			} else {
				fmt.Println("Usage: import [filepath]")
				fmt.Println()
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
			fmt.Print("Files found:")
			for _, file := range files {
				fmt.Println(file.Name)
			}
		case "hash":
			if len(args) == 1 {
				orcaHash.HashFile(args[0])
			} else {
				fmt.Println("Usage: hash [fileName]")
				fmt.Println()
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
				fmt.Println()
			}

		case "exit":
			fmt.Println("Exiting...")
			return
		case "getdir":
			if len(args) == 3 {
				go client.GetDirectory(args[0], args[1], args[2])
			} else {
				fmt.Println("Usage: getdir [ip] [port] [path]")
				fmt.Println()
			}
		case "storedir":
			if len(args) == 3 {
				go client.StoreDirectory(args[0], args[1], args[2])
			} else {
				fmt.Println("Usage: storedir [ip] [port] [path]")
				fmt.Println()
			}
		case "help":
			fmt.Println("COMMANDS:")
			fmt.Println(" get [fileHash]                 Request a file from DHT")
			fmt.Println(" store [fileHash] [amount]      Store a file on DHT")
			fmt.Println(" getdir [ip] [port] [path]      Request a directory")
			fmt.Println(" storedir [ip] [port] [path]    Request storage of a directory")
			fmt.Println(" import [filepath]              Import a file")
			fmt.Println(" send [amount] [ip]             Send an amount of money to network")
			fmt.Println(" hash [fileName]                Get the hash of a file")
			fmt.Println(" list                           List all files you are storing")
			fmt.Println(" location                       Print your location")
			fmt.Println(" network                        Test speed of network")
			fmt.Println(" exit                           Exit the program")
			fmt.Print()
		default:
			fmt.Println("Unknown command. Type 'help' for available commands.")
			fmt.Println()
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
