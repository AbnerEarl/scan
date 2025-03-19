package katana

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/AbnerEarl/goutils/cmdc"
	"github.com/AbnerEarl/goutils/uuid"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

//go:embed katana trufflehog.yaml
var assets embed.FS

type ResponseData struct {
	Timestamp string   `json:"timestamp"`
	Request   request  `json:"request"`
	Response  response `json:"response"`
}
type request struct {
	Method   string `json:"method"`
	Endpoint string `json:"endpoint"`
	Raw      string `json:"raw"`
}

type response struct {
	StatusCode    int               `json:"status_code"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	ContentLength int               `json:"content_length"`
	Raw           string            `json:"raw"`
}

type SensitiveResult struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	URL   string `json:"url"`
}

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

func ScanUrl(url string) (string, error) {
	return ScanUrls([]string{url}, 3, true)
}

func ScanUrls(urlList []string, maxDepth int, healthCheck bool) (string, error) {

	uid, _ := uuid.NewV1Str()
	dir := fmt.Sprintf("/tmp/%s", uid)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}
	content, err := assets.ReadFile("katana")
	os.WriteFile(fmt.Sprintf("%s/katana", dir), content, os.ModePerm)

	if healthCheck {
		var test strings.Builder
		var chromePath string
		if chromePath, err = exec.LookPath("chrome"); err == nil {
			test.WriteString(fmt.Sprintf("Potential chrome binary path (linux/osx) => %s\n", chromePath))
		}
		if err != nil {
			if chromePath, err = exec.LookPath("chrome.exe"); err == nil {
				test.WriteString(fmt.Sprintf("Potential chrome.exe binary path (windows) => %s\n", chromePath))
			}
		}
		if err != nil {
			return "", err
		}
	}
	fileName := fmt.Sprintf("%s/%s.json", dir, uid)
	cmd := fmt.Sprintf("cd %s && ./katana -u %s -sf key,fqdn,qurl  -silent -jsonl -o %s -c 10 -d %d", dir, strings.Join(urlList, ","), fileName, maxDepth)
	err = cmdc.Shell(cmd)

	return fileName, err
}

func ScanSensitive(filePath string, yamlFilePath string) ([]SensitiveResult, error) {

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

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)

	// 读取 JSON 数组的开始字符 '['
	//_, err = decoder.Token()
	//if err != nil {
	//	return nil, err
	//}

	var result []SensitiveResult
	// 循环读取每个 JSON 对象
	for {
		var rd ResponseData
		// 逐个解码
		if err = decoder.Decode(&rd); err != nil {
			// 检查是否已到达文件末尾
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		// 处理每个解码的数据
		result = append(result, regexGrep(rd.Response.Body, rd.Request.Endpoint, patterns)...)
	}

	// 读取 JSON 数组的结束字符 ']'
	//_, err = decoder.Token()
	//if err != nil {
	//	return nil, err
	//}

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
