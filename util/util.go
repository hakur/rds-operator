package util

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Version struct {
	// Major 主版本，通常包含不可逆的变更
	Major int
	// Minor 次要版本，包含小范围功能改动
	Minor int
	// Bugfix 次要版本的bug修复版本
	Bugfix string
}

// ParseVersion 解析版本号，支持 v0.0.0 和 0.0.0 这种方式
func ParseVersion(versionStr string) (version *Version, err error) {
	version = new(Version)
	versionArr := strings.Split(strings.Trim(versionStr, "v"), ".")

	if len(versionArr) < 3 {
		return nil, fmt.Errorf("版本号解析长度小于 3")
	}

	if major, err := strconv.Atoi(versionArr[0]); err != nil {
		return nil, err
	} else {
		version.Major = major
	}

	if minor, err := strconv.Atoi(versionArr[1]); err != nil {
		return nil, err
	} else {
		version.Minor = minor
	}

	version.Bugfix = versionArr[2]

	return
}

func InArray(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func DelArryElement(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func EnvOrDefault(envName, defaultValue string) string {
	val := os.Getenv(envName)
	if val == "" {
		val = defaultValue
	}
	return val
}
