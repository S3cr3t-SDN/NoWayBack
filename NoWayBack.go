package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func main() {

	if len(os.Args) > 1 {

		input := os.Args[1]
		if FileExists(input) {

			// dealing with input as valid file containing hosts list
			hosts := RemoveDup(FiletoArray(input))
			for _, host := range hosts {
				PrintOut(WebArchive(host, 20, 3))
			}

		} else {
			// deal with input as hostname
			PrintOut(WebArchive(input, 20, 3))
		}

	} else {

		fmt.Println("\n[+] Usage: ")
		fmt.Println("  ./" + filepath.Base(os.Args[0]) + " domain.com")
		fmt.Println("   OR   ")
		fmt.Println("  ./" + filepath.Base(os.Args[0]) + " list.txt")
		fmt.Println()

	}

}

func WebArchive(hostname string, timeout, retries int) []string {

	hostname = CleanHost(hostname)
	if retries != 0 && len(hostname) > 3 {

		req, _ := http.NewRequest("GET", "http://web.archive.org/cdx/search/cdx?url="+hostname+"/*&collapse=urlkey&filter=!statuscode:404", nil)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
		req.Header.Add("Connection", "Close")

		client := http.Client{Timeout: time.Duration(timeout) * time.Second}
		resp, err := client.Do(req)

		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "timeout") {
				if retries != 0 {
					retries = retries - 1
					timeout = timeout + 20
					return WebArchive(hostname, timeout, retries)
				} else {
					return []string{"https://" + hostname}
				}
			} else {
				//fmt.Println("WebArchive - client.Do : " + err.Error())
				if retries != 0 {
					retries = retries - 1
					return WebArchive(hostname, timeout, retries)
				} else {
					return []string{"https://" + hostname}
				}
			}
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if retries != 0 {
				retries = retries - 1
				return WebArchive(hostname, timeout, retries)
			} else {
				return []string{"https://" + hostname}
			}
		}
		body_text := string(body)
		// rate limiting check
		if strings.Contains(body_text, "429 Too Many Requests") {
			// sleep for 5 seconds and retry
			time.Sleep(time.Duration(5))
			if retries != 0 {
				timeout = timeout + 10
				return WebArchive(hostname, timeout, retries)
			}
		}

		defer resp.Body.Close()
		return ProcessResponse(hostname, body_text)

	} else {
		return []string{"https://" + hostname}
	}

}

func ProcessResponse(hostname, resp string) []string {

	urls := []string{"https://" + hostname}
	lines := strings.Split(resp, "\n")

	for _, line := range lines {

		line := strings.Split(line, " ")
		for _, part := range line {

			if strings.HasPrefix(part, "http") {

				if !CheckExtension(part) {
					part = strings.Replace(part, ":80", "", -1)
					part = strings.Replace(part, ":443", "", -1)
					urls = append(urls, part)
				}

			}
		}
	}

	return RemoveDup(urls)
}

func CheckExtension(url string) bool {

	exts := []string{".gif", ".jpg", ".ico", ".png", ".jpeg", ".css", ".eot", ".rtf", ".ttf", ".otf", ".svg", ".exe", ".woff", ".woff2", ".swf"}

	url_ := strings.Split(url, "?")
	for _, ext := range exts {
		if strings.HasSuffix(strings.ToLower(url_[0]), ext) {
			return true
		}
	}

	return false
}

func CleanHost(str string) string {
	// can't be better :(, but its stil working lol
	str = strings.Replace(str, "http://", "", -1)
	str = strings.Replace(str, "https://", "", -1)
	str = strings.Replace(str, "*.", "", -1)
	str = strings.Replace(str, ".*", ".com", -1)
	// this to remove anything after hostname ex: google.com[/here]
	re, _ := regexp.Compile(`(?m)\/(.*)`)
	str = re.ReplaceAllString(str, "")
	return str
}

func FiletoArray(FilePath string) []string {

	lines := []string{}

	if FileExists(FilePath) {
		file, err := os.Open(FilePath)
		if err != nil {
			fmt.Println(err.Error())
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err.Error())
		}
	}
	return lines
}

func RemoveDup(subs []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, sub := range subs {
		if _, value := allKeys[sub]; !value {
			allKeys[sub] = true
			list = append(list, sub)
		}
	}
	return list
}

func PrintOut(urls []string) {
	for _, url := range urls {
		fmt.Println(url)
	}
}

func FileExists(path string) bool {
	_, err := os.Open(path)
	return err == nil
}
