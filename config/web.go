package config

import "fmt"

const BASE_URL = ""
const DOMAIN = "localhost:8081"
const IS_HTTPS = ""
const FULL_URL = "http" + IS_HTTPS + "://" + DOMAIN + BASE_URL

func UrlFor(url string) string {
	return fmt.Sprintf("%s%s", FULL_URL, url)
}
