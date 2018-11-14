package main

import "os"
import "fmt"

func IsFileExist(filename string) bool {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func Welcome() {
	fmt.Printf("\n====================================================================\n")
	fmt.Printf("                           欢迎来到区块链系统!\n")
	fmt.Printf("====================================================================\n")
}
