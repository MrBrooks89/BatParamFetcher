package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/gocolly/colly"
)

var cache sync.Map // map[domain]urls
var outputFlag *string

func LogInfo(domain string, cacheURLsFound, waybackURLsFound, cleanedURLs int, cleanedUrls []string) {
	blue := color.New(color.FgBlue).SprintFunc()
	fmt.Printf("[%s] Cleaning URLs for %s\n", blue("INFO"), domain)
	fmt.Printf("[%s] Found %d URLs\n", blue("INFO"), cacheURLsFound)
	//fmt.Printf("[%sINFO%s] Found %d URLs from Wayback Machine\n", waybackURLsFound)
	fmt.Printf("[%s] Found %d URLs after cleaning\n", blue("INFO"), cleanedURLs)
	fmt.Printf("[%s] Extracting URLs with parameters\n", blue("INFO"))

	if cleanedURLs > 0 {
		fmt.Printf("[%s] Saved cleaned URLs to %s\\%s.txt\n", blue("INFO"), *outputFlag, domain)

		// Save cleaned URLs to file
		filename := fmt.Sprintf("%s/%s.txt", *outputFlag, domain)
		if err := SaveToFile(filename, cleanedUrls); err != nil {
			fmt.Printf("Error saving cleaned URLs to file: %s\n", err)
		}
	} else {
		fmt.Printf("[%s] No URLs found after cleaning, skipping file creation\n", blue("INFO"))
	}
}

func main() {
	listFlag := flag.String("l", "", "Path to a file containing a list of domains")
	helpFlag := flag.Bool("h", false, "Show usage information")
	outputFlag = flag.String("o", "results", "Path to the output directory")

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return
	}

	if err := os.MkdirAll(*outputFlag, os.ModePerm); err != nil {
		fmt.Printf("Error creating output directory: %s\n", err)
		return
	}

	var domains []string
	if *listFlag != "" {
		file, err := os.Open(*listFlag)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			domains = append(domains, strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	} else {
		fmt.Println("Please provide a file containing a list of domains using the -l flag.")
		return
	}

	for _, domain := range domains {
		cacheURLsFound, waybackURLsFound := 0, 0 // Reset counts for each domain
		urls, err := FetchUrlsFromWaybackMachine(domain)
		if err != nil {
			fmt.Printf("Error fetching URLs for %s: %s\n", domain, err)
			continue
		}

		cleanedUrls := CleanUrls(urls, "FUZZ")

		cacheURLsFound += len(urls)
		waybackURLsFound += len(cleanedUrls)

		LogInfo(domain, cacheURLsFound, waybackURLsFound, len(cleanedUrls), cleanedUrls)
	}
}

func FetchUrlsFromWaybackMachine(domain string) ([]string, error) {
	if urls, ok := cache.Load(domain); ok {
		return urls.([]string), nil
	}

	waybackURL := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=%s/*&output=txt&collapse=urlkey&fl=original&page=/", domain)

	c := colly.NewCollector()

	var urls []string
	c.OnResponse(func(r *colly.Response) {
		urls = strings.Split(string(r.Body), "\n")
		cache.Store(domain, urls)
	})

	c.Visit(waybackURL)
	c.Wait()

	return urls, nil
}

func CleanUrls(urls []string, placeholder string) []string {
	cleanedUrls := make(map[string]struct{})
	for _, u := range urls {
		// Skip URLs with multiple occurrences of http:// or https://
		if regexp.MustCompile(`(?:https?://){2,}`).MatchString(u) {
			continue
		}

		cleanedUrl := CleanUrl(u)
		if strings.Contains(cleanedUrl, "?") {
			parts := strings.Split(cleanedUrl, "?")
			baseURL := parts[0]
			queryParams := strings.Split(parts[1], "&")
			newQueryParams := make([]string, 0)
			for _, param := range queryParams {
				keyVal := strings.Split(param, "=")
				if len(keyVal) > 1 {
					key := keyVal[0]
					val := keyVal[1]
					val = "FUZZ"
					if len(keyVal) > 2 {
						for i := 2; i < len(keyVal); i++ {
							val += "=" + "FUZZ"
						}
					}
					newQueryParams = append(newQueryParams, key+"="+val)
				}
			}
			if len(newQueryParams) > 0 {
				cleanedUrl = baseURL + "?" + strings.Join(newQueryParams, "&")
				cleanedUrls[cleanedUrl] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(cleanedUrls))
	for url := range cleanedUrls {
		result = append(result, url)
	}
	return result
}

func CleanUrl(url string) string {
	return url
}

func SaveToFile(filename string, urls []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, url := range urls {
		_, err := file.WriteString(url + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
