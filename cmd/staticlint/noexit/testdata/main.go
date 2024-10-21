package main

import "os"

func main() {
	os.Exit(1) // want "ос.Exit should not be used in the main function"
}
