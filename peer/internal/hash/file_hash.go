package hash

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type MemSize int

const (
	Byte     = 1
	Kilobyte = Byte * 1000
	Megabyte = Kilobyte * 1000
)

type NameMap struct {
	mapping map[string]string
	path    string
}

/*
File data read only
*/

type DataStore struct {
	path       string
	buf        map[string][]byte
	buf_size   int
	buf_cap    int
	drive_size int
	drive_cap  int
}

func NewNameStore(path string) *NameMap {
	Assert(os.MkdirAll(path, 0755) == nil, "Failed to create namestore dir")
	name_map := &NameMap{
		mapping: map[string]string{},
		path:    path,
	}
	name_map.Recover()

	return name_map
}

func NewDataStore(path string) *DataStore {
	Assert(os.MkdirAll(path, 0755) == nil, "Failed to create namestore dir")
	return &DataStore{
		path:       path,
		buf:        map[string][]byte{},
		buf_size:   0,
		buf_cap:    4 * Kilobyte,
		drive_size: 0,
		drive_cap:  100 * Megabyte,
	}
}

func HashFile(address string) (string, error) {
	f, err := os.Open("./files/" + address)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
		return "", err
	}
	str := hex.EncodeToString(h.Sum(nil))
	fmt.Printf("%s", str);
	fmt.Println("");
	return str, nil
}

func (nmp *NameMap) GetFileHash(name string) string {
	return nmp.mapping[name]
}

func (nmp *NameMap) PutFileHash(name string, hash_val string) {
	nmp.mapping[name] = hash_val
	Assert(nmp.SaveMapping() == nil, "Failed to save mapping")
}

func (nmp *NameMap) SaveMapping() error {
	res, err := json.Marshal(nmp.mapping)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(nmp.path, "mapping"), res, 0644)
}

func (npm *NameMap) Recover() {
	res, err := os.ReadFile(filepath.Join(npm.path, "mapping"))
	if err == nil {
		var mapping map[string]string
		Assert(json.Unmarshal(res, &mapping) == nil, "Failed to parse mapping")
		npm.mapping = mapping
	}
}

func (ds *DataStore) GetFile(hash_val string) ([]byte, error) {
	if data, ok := ds.buf[hash_val]; ok {
		return data, nil
	}

	file, err := ds.OpenFile(hash_val)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(bufio.NewReader(file))
	if err != nil && err != io.EOF {
		return []byte{}, err
	}

	ds.BufferPut(hash_val, data)

	return data, nil
}

func (ds *DataStore) PutFile(data []byte) (string, error) {
	checksum := sha256.Sum256(data)
	hash_val := fmt.Sprintf("%x", checksum)
	if err := ds.DrivePut(hash_val, data); err != nil {
		return "", err
	}

	ds.BufferPut(hash_val, data)

	return hash_val, nil
}

func (ds *DataStore) BufferPut(hash_val string, data []byte) {
	if len(data)+ds.buf_size > ds.buf_cap {
		ds.EvictBuffer()
	}
	ds.buf[hash_val] = data
}

func (ds *DataStore) EvictBuffer() {
	largest_file_hash := ""
	largest_file_size := 0
	for hash_val, data := range ds.buf {
		if len(data) > largest_file_size {
			largest_file_hash = hash_val
		}
	}
	if largest_file_hash != "" {
		delete(ds.buf, largest_file_hash)
	}
}

func (ds *DataStore) DrivePut(hash_val string, data []byte) error {
	if len(data)+ds.drive_size > ds.drive_cap {
		fmt.Printf("Drive evict %d %d\n", ds.drive_size, len(data))
		ds.DriveEvict()
	}
	return ds.WriteFile(hash_val, data)
}

func (ds *DataStore) DriveEvict() {
	entries, err := os.ReadDir(ds.path)
	Assert(err == nil, "Todo handle directory read failure during drive eviction")

	largest_file_hash := ""
	largest_file_size := 0
	for _, entry := range entries {
		info, err := entry.Info()
		Assert(err == nil, "Todo handle stat on dir entry fail")
		if info.Size() > int64(largest_file_size) {
			largest_file_hash = info.Name()
		}
	}
	if largest_file_hash != "" {
		Assert(os.Remove(filepath.Join(ds.path, largest_file_hash)) == nil, "Todo remove file failed")
	}
}

func (ds *DataStore) OpenFile(hash_val string) (*os.File, error) {
	return os.Open(filepath.Join(ds.path, hash_val))
}

func (ds *DataStore) WriteFile(hash_val string, data []byte) error {
	return os.WriteFile(filepath.Join(ds.path, hash_val), data, 0444)
}
