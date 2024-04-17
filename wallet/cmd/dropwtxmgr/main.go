// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/jessevdk/go-flags"
)

const defaultNet = "mainnet"

var (
	datadir = btcutil.AppDataDir("btcwallet", false)
)

// Flags.
var opts = struct {
	Force      bool          `short:"f" description:"Force removal without prompt"`
	DbPath     string        `long:"db" description:"Path to wallet database"`
	DropLabels bool          `long:"droplabels" description:"Drop transaction labels"`
	Timeout    time.Duration `long:"timeout" description:"Timeout value when opening the wallet database"`
}{
	Force:   false,
	DbPath:  filepath.Join(datadir, defaultNet, wallet.WalletDBName),
	Timeout: wallet.DefaultDBTimeout,
}

func init() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
}

func yes(s string) bool {
	switch s {
	case "y", "Y", "yes", "Yes":
		return true
	default:
		return false
	}
}

func no(s string) bool {
	switch s {
	case "n", "N", "no", "No":
		return true
	default:
		return false
	}
}

func main() {
	os.Exit(mainInt())
}

func mainInt() int {
	fmt.Println("Database path:", opts.DbPath)
	_, err := os.Stat(opts.DbPath)
	if os.IsNotExist(err) {
		fmt.Println("Database file does not exist")
		return 1
	}

	for !opts.Force {
		fmt.Print("Drop all btcwallet transaction history? [y/N] ")

		scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
		if !scanner.Scan() {
			// Exit on EOF.
			return 0
		}
		err := scanner.Err()
		if err != nil {
			fmt.Println()
			fmt.Println(err)
			return 1
		}
		resp := scanner.Text()
		if yes(resp) {
			break
		}
		if no(resp) || resp == "" {
			return 0
		}

		fmt.Println("Enter yes or no.")
	}

	db, err := walletdb.Open("bdb", opts.DbPath, true, opts.Timeout)
	if err != nil {
		fmt.Println("Failed to open database:", err)
		return 1
	}
	defer db.Close()

	fmt.Println("Dropping btcwallet transaction history")

	err = wallet.DropTransactionHistory(db, !opts.DropLabels)
	if err != nil {
		fmt.Println("Failed to drop and re-create namespace:", err)
		return 1
	}

	return 0
}
