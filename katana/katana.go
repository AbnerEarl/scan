package katana

import (
	"embed"
	"fmt"
	"github.com/AbnerEarl/goutils/cmdc"
	"github.com/AbnerEarl/goutils/uuid"
	"os"
	"os/exec"
	"strings"
)

//go:embed katana
var assets embed.FS

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
