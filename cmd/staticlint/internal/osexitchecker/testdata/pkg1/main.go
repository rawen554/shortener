package main

import "os"

func anotherFunc() {
	os.Exit(1)
}

func main() {
	os.Exit(1) // want "calling os.Exit in main package main func"
}
