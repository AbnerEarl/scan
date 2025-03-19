package ffuf

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/AbnerEarl/goutils/cmdc"
	"github.com/AbnerEarl/goutils/uuid"
	"os"
	"strings"
)

//go:embed ffuf onelistforallshort.txt
var assets embed.FS

type ResponseData struct {
	Input            input             `json:"input"`
	Position         int               `json:"position"`
	Status           int               `json:"status"`
	Length           int               `json:"length"`
	Words            int               `json:"words"`
	Lines            int               `json:"lines"`
	ContentType      string            `json:"content-type"`
	RedirectLocation string            `json:"redirectlocation"`
	Url              string            `json:"url"`
	Duration         int               `json:"duration"`
	Scraper          map[string]string `json:"scraper"`
	ResultFile       string            `json:"resultfile"`
	Host             string            `json:"host"`
}

type input struct {
	FFUFHASH string `json:"FFUFHASH"`
	FUZZ     string `json:"FUZZ"`
}

func ScanUrl(url string) (string, error) {
	return ScanUrls(url, "", 3)
}

func ScanUrls(url, wordPath string, maxDepth int) (string, error) {

	uid, _ := uuid.NewV1Str()
	dir := fmt.Sprintf("/tmp/%s", uid)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
	}
	content, err := assets.ReadFile("ffuf")
	os.WriteFile(fmt.Sprintf("%s/ffuf", dir), content, os.ModePerm)
	if strings.Trim(wordPath, " ") == "" {
		wordPath = fmt.Sprintf("%s/onelistforallshort.txt", dir)
		content, err = assets.ReadFile("onelistforallshort.txt")
		os.WriteFile(wordPath, content, os.ModePerm)
	}

	fileName := fmt.Sprintf("%s/%s.json", dir, uid)
	cmd := fmt.Sprintf("cd %s && ./ffuf  -w %s -u  %s/FUZZ -maxtime-job 60 -recursion -recursion-depth %d -json -o %s", dir, wordPath, url, maxDepth, fileName)
	err = cmdc.Shell(cmd)

	return fileName, err
}

func ReadContent(filePath string) ([]ResponseData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建 JSON 解码器
	decoder := json.NewDecoder(file)

	// 读取 JSON 数组的开始字符 '['
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}

	var result []ResponseData
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
		result = append(result, rd)
	}

	// 读取 JSON 数组的结束字符 ']'
	_, err = decoder.Token()
	if err != nil {
		return nil, err
	}

	return result, nil
}
