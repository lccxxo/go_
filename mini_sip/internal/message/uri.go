package message

import (
	"fmt"
	"strconv"
	"strings"
)

// URI 表示 SIP URI（RFC 3261 §19.1）。
//
// 格式：sip:user[:password]@host[:port][;uri-parameters][?headers]
//
// 示例：
//
//	sip:alice@atlanta.com
//	sip:alice:secretword@atlanta.com;transport=tcp
//	sips:alice@atlanta.com?subject=project%20x&priority=urgent
type URI struct {
	Scheme   string            // "sip" 或 "sips"
	User     string            // 用户名
	Password string            // 密码（一般不用）
	Host     string            // 主机名或 IP
	Port     int               // 端口，0 表示使用协议默认值
	Params   map[string]string // URI 参数（transport, lr, maddr 等）
	Headers  map[string]string // URI 头部（?subject=xxx）
}

// ParseURI 将字符串解析为 URI。
func ParseURI(s string) (*URI, error) {
	s = strings.TrimSpace(s)
	uri := &URI{Params: make(map[string]string), Headers: make(map[string]string)}

	// 提取 Scheme
	schemeEnd := strings.Index(s, ":")
	if schemeEnd < 0 {
		return nil, fmt.Errorf("missing scheme in URI: %q", s)
	}
	uri.Scheme = strings.ToLower(s[:schemeEnd])
	if uri.Scheme != "sip" && uri.Scheme != "sips" {
		return nil, fmt.Errorf("unsupported URI scheme: %q", uri.Scheme)
	}
	rest := s[schemeEnd+1:]

	// 分离 headers（?...）
	if qIdx := strings.Index(rest, "?"); qIdx >= 0 {
		for _, kv := range strings.Split(rest[qIdx+1:], "&") {
			kv = strings.TrimSpace(kv)
			if kv == "" {
				continue
			}
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				uri.Headers[parts[0]] = parts[1]
			}
		}
		rest = rest[:qIdx]
	}

	// 分离 URI params（;...）
	paramParts := strings.Split(rest, ";")
	rest = paramParts[0]
	for _, seg := range paramParts[1:] {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		kv := strings.SplitN(seg, "=", 2)
		key := strings.ToLower(kv[0])
		val := ""
		if len(kv) == 2 {
			val = kv[1]
		}
		uri.Params[key] = val
	}

	// 分离 userinfo(@host)
	atIdx := strings.LastIndex(rest, "@")
	if atIdx >= 0 {
		userInfo := rest[:atIdx]
		rest = rest[atIdx+1:]
		// user[:password]
		if colonIdx := strings.Index(userInfo, ":"); colonIdx >= 0 {
			uri.User = userInfo[:colonIdx]
			uri.Password = userInfo[colonIdx+1:]
		} else {
			uri.User = userInfo
		}
	}

	// host[:port]
	// 处理 IPv6 地址 [::1]:5060
	if strings.HasPrefix(rest, "[") {
		end := strings.Index(rest, "]")
		if end < 0 {
			return nil, fmt.Errorf("invalid IPv6 in URI: %q", rest)
		}
		uri.Host = rest[:end+1]
		rest = rest[end+1:]
		if strings.HasPrefix(rest, ":") {
			port, err := strconv.Atoi(rest[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid port in URI: %w", err)
			}
			uri.Port = port
		}
	} else {
		colonIdx := strings.LastIndex(rest, ":")
		if colonIdx >= 0 {
			port, err := strconv.Atoi(rest[colonIdx+1:])
			if err == nil {
				uri.Port = port
				uri.Host = rest[:colonIdx]
			} else {
				uri.Host = rest
			}
		} else {
			uri.Host = rest
		}
	}

	if uri.Host == "" {
		return nil, fmt.Errorf("missing host in URI: %q", s)
	}
	return uri, nil
}

// String 将 URI 序列化为字符串。
func (u *URI) String() string {
	var sb strings.Builder
	sb.WriteString(u.Scheme)
	sb.WriteString(":")
	if u.User != "" {
		sb.WriteString(u.User)
		if u.Password != "" {
			sb.WriteString(":")
			sb.WriteString(u.Password)
		}
		sb.WriteString("@")
	}
	sb.WriteString(u.Host)
	if u.Port > 0 {
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(u.Port))
	}
	for k, v := range u.Params {
		sb.WriteString(";")
		sb.WriteString(k)
		if v != "" {
			sb.WriteString("=")
			sb.WriteString(v)
		}
	}
	first := true
	for k, v := range u.Headers {
		if first {
			sb.WriteString("?")
			first = false
		} else {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}

// Clone 深拷贝 URI。
func (u *URI) Clone() *URI {
	n := &URI{
		Scheme:   u.Scheme,
		User:     u.User,
		Password: u.Password,
		Host:     u.Host,
		Port:     u.Port,
		Params:   make(map[string]string, len(u.Params)),
		Headers:  make(map[string]string, len(u.Headers)),
	}
	for k, v := range u.Params {
		n.Params[k] = v
	}
	for k, v := range u.Headers {
		n.Headers[k] = v
	}
	return n
}
