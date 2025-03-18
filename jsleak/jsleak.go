package jsleak

import (
	"crypto/tls"
	"embed"
	"fmt"
	"github.com/AbnerEarl/goutils/uuid"
	"gopkg.in/yaml.v2"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

//go:embed trufflehog.yaml
var assets embed.FS

type patternDef struct {
	Name       string `yaml:"name"`
	Regex      string `yaml:"regex"`
	Confidence string `yaml:"confidence"`
}

type patternWrapper struct {
	Pattern patternDef `yaml:"pattern"`
}

type yamlPatterns struct {
	Patterns []patternWrapper `yaml:"patterns"`
}

type SensitiveResult struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	URL   string `json:"url"`
}

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: time.Second,
			DualStack: true,
		}).DialContext,
	},
}

func request(fullurl string, checkStatus bool) string {
	req, err := http.NewRequest("GET", fullurl, nil)
	if err != nil {
		return ""
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.100 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if checkStatus && resp.StatusCode == 404 {
		return ""
	}

	var bodyString string
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ""
		}
		bodyString = string(bodyBytes)
	}
	return bodyString
}

func regexGrep(content string, baseUrl string, patterns []patternDef) []SensitiveResult {
	var result []SensitiveResult
	for _, p := range patterns {
		r := regexp.MustCompile(p.Regex)
		matches := r.FindAllString(content, -1)
		for _, v := range matches {
			result = append(result, SensitiveResult{
				Name:  p.Name,
				Value: v,
				URL:   baseUrl,
			})
		}
	}
	return result
}

func linkFinder(content, baseURL string, completeURL, checkStatus bool) ([]string, error) {
	linkRegex := `(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{3,}(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:[\?|#][^"|']{0,}|)))(?:"|')`
	r := regexp.MustCompile(linkRegex)
	matches := r.FindAllString(content, -1)

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, match := range matches {
		cleanedMatch := strings.Trim(match, `"'`)
		link, err := url.Parse(cleanedMatch)
		if err != nil {
			continue
		}
		if completeURL {
			link = base.ResolveReference(link)
		}
		if checkStatus {
			res := request(link.String(), true)
			if res != "" {
				result = append(result, link.String())
			}
		} else {
			result = append(result, link.String())
		}
	}
	return result, nil
}

func ScanUrl(url string, depth int) ([]SensitiveResult, error) {
	return ScanUrlList([]string{url}, depth, "", true, true, false, false)
}

func ScanUrlList(urlList []string, depth int, yamlFilePath string, enableSecretFinder, enableLinkFinder, checkStatus, completeURL bool) ([]SensitiveResult, error) {
	concurrency := 10
	var patterns []patternDef
	if strings.Trim(yamlFilePath, " ") == "" {
		uid, _ := uuid.NewV1Str()
		dir := fmt.Sprintf("/tmp/%s", uid)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
		content, err := assets.ReadFile("trufflehog.yaml")
		os.WriteFile(fmt.Sprintf("%s/trufflehog.yaml", dir), content, os.ModePerm)
		yamlFilePath = dir + "/trufflehog.yaml"
	}
	if yamlFilePath != "" {
		loadedPatterns, err := loadPatternsFromYAML(yamlFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading YAML patterns: %v\n", err)
			return nil, err
		}
		for _, pw := range loadedPatterns.Patterns {
			patterns = append(patterns, pw.Pattern)
		}
	}

	urls := make(chan string, concurrency)
	go func() {
		for _, u := range urlList {
			urls <- u
		}
		close(urls)
	}()

	var result []SensitiveResult
	wg := sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for vUrl := range urls {
				res := request(vUrl, false)

				if enableSecretFinder && len(patterns) > 0 {
					result = append(result, regexGrep(res, vUrl, patterns)...)
				}

				if enableLinkFinder && depth > 0 {
					linkList, err := linkFinder(res, vUrl, completeURL, checkStatus)
					if err == nil {
						r, e := ScanUrlList(linkList, depth-1, yamlFilePath, enableSecretFinder, enableLinkFinder, checkStatus, completeURL)
						if e == nil {
							result = append(result, r...)
						}
					}
				}

			}
		}()
	}
	wg.Wait()
	return result, nil
}

func loadPatternsFromYAML(filePath string) (*yamlPatterns, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	var yp yamlPatterns
	if err := decoder.Decode(&yp); err != nil {
		return nil, err
	}
	return &yp, nil
}
