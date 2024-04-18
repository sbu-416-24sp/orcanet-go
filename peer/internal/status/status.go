package status

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cbergoon/speedtest-go"
)

type NetworkStatus struct {
	Success           bool
	DownloadSpeedMbps float64
	UploadSpeedMbps   float64
	LatencyMs         float64
}

type PeerNodeFileData struct {
	IsMe          bool    `json:"is_me"`
	Balance       float64 `json:"balance"`
	PublicKey     string  `json:"public_key"`
	PublicKeyPath string  `json:"public_key_path"`
}
type PeerNode struct {
	IsMe      bool
	Balance   float64
	PublicKey string
}
type PeerNodeJSON struct {
	PublicKey string `json:"public_key"`
	Address   string `json:"ip_address"`
	Location  string `json:"location"`
}
type PeerNodes struct {
	Nodes []PeerNodeJSON `json:"peers"`
}

func GetNodeInfo() PeerNode {
	jsonFile, err := os.Open("config/self.json")
	if err != nil {
		fmt.Println("Error on loading config, please try again")
	}
	var meData PeerNodeFileData
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &meData)
	defer jsonFile.Close()
	var me PeerNode
	me.Balance = meData.Balance
	me.IsMe = meData.IsMe
	me.PublicKey = meData.PublicKey
	if me.PublicKey == "" {
		content, error := os.ReadFile(meData.PublicKeyPath)
		if error != nil {
			fmt.Println("Error: unable to read in ")
			return PeerNode{}
		}
		pub_key := string(content)
		me.PublicKey = pub_key
	}
	return me
}

func GetPeerNodeInfo() PeerNodes {
	jsonFile, err := os.Open("config/peers.json")
	if err != nil {
		fmt.Println("Error on loading config, please try again")
	}
	var nodes PeerNodes
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &nodes)
	defer jsonFile.Close()
	return nodes
}

func GetNetworkInfo() NetworkStatus {
	user, _ := speedtest.FetchUserInfo()

	serverList, _ := speedtest.FetchServerList(user)
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest()
		s.DownloadTest()
		s.UploadTest()
		return NetworkStatus{Success: true, LatencyMs: float64(s.Latency), DownloadSpeedMbps: s.DLSpeed, UploadSpeedMbps: s.ULSpeed}
	}
	return NetworkStatus{Success: false}
}
func GetLocationData() string {
	ipapiClient := http.Client{}

	ipv4Req, err := http.NewRequest("GET", "http://httpbin.org/ip", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := ipapiClient.Do(ipv4Req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	ipv4Body := string(body)
	var ipv4JSON map[string]interface{}
	err = json.Unmarshal([]byte(ipv4Body), &ipv4JSON)
	if err != nil {
		log.Fatal("Unable to establish user IP, please try again")
		return ""
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://ipapi.co/%s/json/", ipv4JSON["origin"].(string)), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "ipapi.co/#go-v1.4.01")

	resp, err = ipapiClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}
