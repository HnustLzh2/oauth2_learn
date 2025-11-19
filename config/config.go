package config

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var cfg App

func YamlSetup() {
	path := flag.String("config", "../../config.yaml", "path to config file")
	flag.Parse()

	if path == nil {
		panic("config file path is nil")
	}
	cfgBytes, err := os.ReadFile(*path)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		panic(err)
	}
}

func JsonSetup() {
	path := flag.String("config", "../../config.json", "path to config file")
	flag.Parse()

	if path == nil {
		panic("config file path is nil")
	}
	cfgBytes, err := os.ReadFile(*path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		panic(err)
	}
}

func ConfigTypeConv(dst string, src string) error {
	srcExt := strings.ToLower(filepath.Ext(src))
	dstExt := strings.ToLower(filepath.Ext(dst))

	var tmp App
	cfgBytes, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	switch srcExt {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(cfgBytes, &tmp)
	case ".json":
		err = json.Unmarshal(cfgBytes, &tmp)
	default:
		return errors.New("config file type not supported")
	}
	if err != nil {
		return err
	}

	var out []byte
	switch dstExt {
	case ".yaml", ".yml":
		out, err = yaml.Marshal(&tmp)
	case ".json":
		out, err = json.MarshalIndent(&tmp, "", "  ")
	default:
		return errors.New("config file type not supported")
	}
	if err != nil {
		return err
	}
	// 	0o644 是以八进制表示的文件权限（file mode）。在 Unix/Linux 中：
	// 三组权限：owner | group | others
	// 每组用 3 bit 表示 r(4) w(2) x(1)
	// 0o644 拆开：
	// 6 → 4+2 → 拥有者可读写
	// 4 → 仅可读（群组）
	// 4 → 仅可读（其他用户）
	// 常见模式及意义：
	// 0o644：普通文本/配置文件默认权限（owner 可读写，其他只读）。
	// 0o600：敏感配置/密钥，仅拥有者读写。
	// 0o755：可执行文件/脚本，所有人可读执行，拥有者还能写。
	// 0o700：仅拥有者读写执行，典型私有脚本。
	// 0o664：群组成员可读写，适合团队共享文件。
	// 选择模式时看需求：是否需要执行、是否允许他人写、是否存放敏感信息等。
	if err := os.WriteFile(dst, out, 0o644); err != nil {
		return err
	}

	// 更新全局配置，便于后续直接使用
	cfg = tmp
	return nil
}
