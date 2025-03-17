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

func ScanUrl(url, proxy string, depth int) (string, error) {
	uid, _ := uuid.NewV1Str()
	dir := fmt.Sprintf("/tmp/%s", uid)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
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

	httpProxy := ""
	if proxy != "" {
		httpProxy = fmt.Sprintf("-http-proxy %s", proxy)
	}
	if depth > 1 {
		op := fmt.Sprintf("sed -i 's/max_depth: 1/max_depth: %d/g' %s/rad_config.yml", depth, dir)
		err = cmdc.Shell(op)
		if err != nil {
			return "", err
		}
		op = fmt.Sprintf("sed -i 's/max_interactive_depth: 1/max_interactive_depth: %d/g' %s/rad_config.yml", depth, dir)
		err = cmdc.Shell(op)
		if err != nil {
			return "", err
		}
	}

	cmd := fmt.Sprintf("cd %s && ./rad_linux_amd64  -t '%s' %s --config rad_config.yml -json %s", dir, url, httpProxy, fileName)
	err = cmdc.Shell(cmd)
	//os.RemoveAll(dir)
	return dir + "/" + fileName, err
}

func ScanUrlFile(urlFile, proxy string, depth int) (string, error) {
	uid, _ := uuid.NewV1Str()
	dir := fmt.Sprintf("/tmp/%s", uid)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", err
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

	httpProxy := ""
	if proxy != "" {
		httpProxy = fmt.Sprintf("-http-proxy %s", proxy)
	}
	if depth > 1 {
		op := fmt.Sprintf("sed -i 's/max_depth: 1/max_depth: %d/g' %s/rad_config.yml", depth, dir)
		err = cmdc.Shell(op)
		if err != nil {
			return "", err
		}
		op = fmt.Sprintf("sed -i 's/max_interactive_depth: 1/max_interactive_depth: %d/g' %s/rad_config.yml", depth, dir)
		err = cmdc.Shell(op)
		if err != nil {
			return "", err
		}
	}

	cmd := fmt.Sprintf("cd %s && ./rad_linux_amd64  --uf '%s' %s --config rad_config.yml -json %s", dir, urlFile, httpProxy, fileName)
	err = cmdc.Shell(cmd)
	//os.RemoveAll(dir)
	return dir + "/" + fileName, err
}
