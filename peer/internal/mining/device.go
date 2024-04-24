package mining

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Device struct {
	DeviceId      string `json:"device_id"`
	DeviceName    string `json:"device_name"`
	HashPower     string `json:"hash_power"`
	Status        string `json:"status"`
	Power         string `json:"power"`
	Profitability string `json:"profitablity"`
}

type DeviceManager struct {
	Devices []Device
	Changed bool
	Mutex   sync.Mutex
}

var Manager DeviceManager

func InitDeviceTracker() {
	Manager = DeviceManager{
		Devices: make([]Device, 0), // Initialize an empty slice of jobs
		Mutex:   sync.Mutex{},      // Initialize a mutex
		Changed: false,
	}
	devs, err := LoadHistory()
	if err != nil {
		fmt.Println("Error loading devices, no devices will be shown.")
		return
	}
	Manager.Devices = devs
	for {
		time.Sleep(10 * time.Second)
		Manager.Mutex.Lock()
		if Manager.Changed {
			SaveHistory(Manager.Devices)
		}
		Manager.Mutex.Unlock()
	}
}

func LoadHistory() ([]Device, error) {
	fileData, err := os.ReadFile("./internal/mining/devices.json")
	if err != nil {
		return nil, err
	}
	var jobs []Device
	err = json.Unmarshal(fileData, &jobs)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
func SaveHistory(devices []Device) error {
	Manager.Changed = false
	jsonData, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	err = os.WriteFile("./internal/mining/devices.json", jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func GetDeviceList() {

}
