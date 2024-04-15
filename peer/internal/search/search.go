package search

import (
	"fmt"
	"io"
	"os"
)

// FileInfo represents information about a file.
type FileInfo struct {
	Name string
	Hash string
	Size int64
	Cost float64
}

func SearchForFile(dirname string, filename string) (bool, error) {
	dir, err := os.Open(dirname)
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return false, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return false, err
	}

	fileFound := false
	for _, file := range files {
		if file.Name() == filename {
			fmt.Println("File found:", file.Name())
			fileFound = true
			break
		}
	}

	if !fileFound {
		fmt.Println("File not found")
	}
	return fileFound, nil
}

func GetAllFiles(dirname string) ([]FileInfo, error) {

	// Create an array to store file objects
	var fileArr []FileInfo
	dir, err := os.Open(dirname)
	if err != nil {
		fmt.Println("Error opening directory:", err)
		return nil, err
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil, err
	}

	for _, file := range files {
		fileInfoObject := FileInfo{
			Name: file.Name(),
			Size: file.Size(),
			Hash: "hash",
			Cost: 10,
		}
		fileArr = append(fileArr, fileInfoObject)
	}

	return fileArr, nil

}

func OutputFileContents(dirname string, filename string) {
	path := dirname + "/" + filename
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, 1000)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		if n > 0 {
			fmt.Println(string(buf[:n]))
		}
	}
}
