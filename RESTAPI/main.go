package main

import "os"

// testing the build server
func main() {
	a := App{}
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))
	a.Run(
		os.Getenv("APP_DB_ADDR"),
		os.Getenv("APP_DB_PORT"))
}
