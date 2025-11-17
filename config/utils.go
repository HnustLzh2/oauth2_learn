package config

import "strings"

func GetCfg() *App {
	return &cfg
}

// GetOAuth2Client 通过clientID获取客户端
func GetOAuth2Client(clientID string) *OAuth2Client {
	for _, client := range cfg.OAuth2.Client {
		if client.ID == clientID {
			return &client
		}
	}
	return nil
}

// JoinScope 把一组scope拼接成一个字符串
func JoinScope(scope []Scope) string {
	var s []string
	for _, sc := range scope {
		s = append(s, sc.ID)
	}
	return strings.Join(s, ",")
}

// ScopeFilter 使用一个scope字符串过滤一个client的权限范围
func ScopeFilter(clientID string, scope string) []Scope {
	result := make([]Scope, 0)
	cli := GetOAuth2Client(clientID)
	if cli == nil {
		return nil
	}
	splitScope := strings.Split(scope, ",")
	for _, str := range splitScope {
		for _, s := range cli.Scope {
			if s.ID == str {
				result = append(result, s)
			}
		}
	}
	return result
}
