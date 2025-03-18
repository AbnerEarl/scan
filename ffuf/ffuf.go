package ffuf

import (
	"embed"
	"fmt"
	"github.com/AbnerEarl/goutils/cmdc"
	"github.com/AbnerEarl/goutils/uuid"
	"os"
	"strings"
)

//go:embed ffuf onelistforallshort.txt
var assets embed.FS

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
