package server

import (
	"errors"
	"fmt"
	pb "orca-peer/internal/fileshare"
	"regexp"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
)

type OrcaValidator struct{}

/*
 * Given a list of values from the DHT, select index of the best one. This is determined by
 * checking which value is the longest and is valid according to the OrcaValidator.
 *
 * Parameters:
 *   key: SHA256 Hash String of file being registered
 *   value: A slice of byte slices that represent the values to be compared.
 *
 * Returns:
 *   The index of the best value
 *   An error, if any
 */
func (v OrcaValidator) Select(key string, value [][]byte) (int, error) {
	max := len(value[0])
	maxIndex := 0
	latestTime := ConvertBytesTo64BitInt(value[0][(len(value[0]) - 8):])
	for i := 1; i < len(value); i++ {
		suppliedTime := ConvertBytesTo64BitInt(value[i][(len(value[i]) - 8):])
		fmt.Println(suppliedTime)
		if len(value[i]) >= max {
			if suppliedTime >= latestTime {
				max = len(value[i])
				latestTime = suppliedTime
				maxIndex = i
			}
		}
	}
	return maxIndex, nil
}

/*
 * Validates keys and values that are being put into the OrcaNet market DHT.
 * Keys must conform to a SHA256 hash, Values must conform the specification in /server/README.md
 *
 * Parameters:
 *   key: SHA256 Hash String of file being registered
 *   value: The value to be put into the DHT, must conform to specification in /server/README.md
 *
 * Returns:
 *   An error, if any
 */
func (v OrcaValidator) Validate(key string, value []byte) error {
	// verify key is a sha256 hash
	hexPattern := "^[a-fA-F0-9]{64}$"
	regex := regexp.MustCompile(hexPattern)
	if !regex.MatchString(strings.Replace(key, "orcanet/market/", "", -1)) {
		return errors.New("Provided key is not in the form of a SHA-256 digest!")
	}

	pubKeySet := make(map[string]bool)

	for i := 0; i < len(value)-8; i++ {
		messageLength := uint16(value[i+1])<<8 | uint16(value[i])
		digitalSignatureLength := uint16(value[i+3])<<8 | uint16(value[i+2])
		contentLength := messageLength + digitalSignatureLength
		user := &pb.User{}

		err := proto.Unmarshal(value[i+4:i+4+int(messageLength)], user)
		if err != nil {
			return err
		}

		if pubKeySet[string(user.GetId())] == true {
			return errors.New("Duplicate record for the same public key found!")
		} else {
			pubKeySet[string(user.GetId())] = true
		}

		userMessageBytes := value[i+4 : i+4+int(messageLength)]

		publicKey, err := crypto.UnmarshalRsaPublicKey([]byte(user.GetId()))
		if err != nil {
			return err
		}

		signatureBytes := value[i+4+int(messageLength) : i+4+int(contentLength)]
		valid, err := publicKey.Verify(userMessageBytes, signatureBytes) //this function will automatically compute hash of data to compare signauture

		if err != nil {
			return err
		}

		if !valid {
			return errors.New("Signature invalid!")
		}

		i = i + 4 + int(contentLength) - 1
	}

	currentTime := time.Now().UTC()
	unixTimestamp := currentTime.Unix()
	unixTimestampInt64 := uint64(unixTimestamp)

	suppliedTime := ConvertBytesTo64BitInt(value[len(value)-8:])
	if suppliedTime > unixTimestampInt64 {
		return errors.New("Supplied time cannot be less than current time")
	}
	return nil
}

/*
 * Convert a max 8 byte slice to its 64 bit int value.
 *
 * Parameters:
 *   value: The byte slice to convert to an int
 *
 * Returns:
 *   An unsigned 64 bit int
 */
func ConvertBytesTo64BitInt(value []byte) uint64 {
	suppliedTime := uint64(0)
	shift := 7
	for i := 0; i < len(value); i++ {
		suppliedTime = suppliedTime | (uint64(value[i]) << (shift * 8))
		shift--
	}
	return suppliedTime
}
