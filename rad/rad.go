package rad

import (
	"embed"
	"fmt"
	"github.com/AbnerEarl/goutils/cmdc"
	"github.com/AbnerEarl/goutils/uuid"
	"os"
)

//go:embed rad_ca.cert rad_ca.key rad_config.yml rad_linux_amd64
var assets embed.FS

func ScanUrl(url string) (string, error) {
	uid, _ := uuid.NewV1Str()
	dir := fmt.Sprintf("/tmp/%s", uid)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	fileName := fmt.Sprintf("result-%s.json", uid)
	content, err := assets.ReadFile("rad_linux_amd64")
	os.WriteFile(fmt.Sprintf("%s/rad_linux_amd64", dir), content, os.ModePerm)
	content, err = assets.ReadFile("rad_config.yml")
	os.WriteFile(fmt.Sprintf("%s/rad_config.yml", dir), content, os.ModePerm)
	content, err = assets.ReadFile("rad_ca.cert")
	os.WriteFile(fmt.Sprintf("%s/rad_ca.cert", dir), content, os.ModePerm)
	content, err = assets.ReadFile("rad_ca.key")
	os.WriteFile(fmt.Sprintf("%s/rad_ca.key", dir), content, os.ModePerm)

	cmd := fmt.Sprintf("cd %s && ./rad_linux_amd64  -t '%s' -json %s", dir, url, fileName)
	fmt.Println(cmd)
	err = cmdc.Shell(cmd)
	//os.RemoveAll(dir)
	return dir + "/" + fileName, err
}
