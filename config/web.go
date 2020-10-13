package config

import "fmt"

const BASE_URL = ""
const FULL_URL = "http://localhost:8081"+BASE_URL


func UrlFor(url string) string {
	return fmt.Sprintf("%s%s", FULL_URL, url)
}