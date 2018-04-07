package main

import "os"

// testing the build serve
func main() {
	a := App{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
		os.Getenv("APP_DB_ADDR"),
		os.Getenv("APP_DB_PORT"))
	a.Run()
}
