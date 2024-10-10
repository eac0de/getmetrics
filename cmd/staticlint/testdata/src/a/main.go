package main

import "os"

func main() {
	os.Exit(1) // want "ос.Exit не должен использоваться в функции main"
}
