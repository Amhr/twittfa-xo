package config

import "fmt"

const BASE_URL = ""
const DOMAIN = "localhost:8081"
const FULL_URL = "http://" + DOMAIN + BASE_URL

func UrlFor(url string) string {
	return fmt.Sprintf("%s%s", FULL_URL, url)
}
