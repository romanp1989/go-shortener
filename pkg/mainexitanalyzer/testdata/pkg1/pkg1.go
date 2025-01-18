package main

import "os"

func main() {
	correctFunc() // correct
	os.Exit(0)    // want "allow not using call os.Exit in main function"
}

func correctFunc() {
	os.Exit(0) // correct
}
