// Package message 实现 SIP 消息的构造、解析和序列化。
//
// SIP 消息由三部分组成：
//  1. 起始行 (Start-Line)：请求行 or 状态行
//  2. 消息头 (Headers)
//  3. 消息体 (Body)，可为空
//
// 参考规范：RFC 3261
package message

import (
	"fmt"
	"strconv"
	"strings"
)

// ---- 常用头域名称 (RFC 3261 §20) ----

const (
	HeaderVia         = "Via"
	HeaderFrom        = "From"
	HeaderTo          = "To"
	HeaderCallID      = "Call-ID"
	HeaderCSeq        = "CSeq"
	HeaderContact     = "Contact"
	HeaderContentType = "Content-Type"
	HeaderContentLen  = "Content-Length"
	HeaderMaxForwards = "Max-Forwards"
	HeaderUserAgent   = "User-Agent"
	HeaderExpires     = "Expires"
	HeaderAllow       = "Allow"
	HeaderAccept      = "Accept"
	HeaderWWWAuth     = "WWW-Authenticate"
	HeaderAuthorize   = "Authorization"
)

// shortForms 将紧凑头域名映射到完整名称（RFC 3261 §20）
var shortForms = map[string]string{
	"v": HeaderVia,
	"f": HeaderFrom,
	"t": HeaderTo,
	"i": HeaderCallID,
	"m": HeaderContact,
	"c": HeaderContentType,
	"l": HeaderContentLen,
}

// Header 表示单个 SIP 头域，允许同名多值（如多个 Via）。
type Header struct {
	Name   string
	Values []string
}

func (h Header) String() string {
	return fmt.Sprintf("%s: %s", h.Name, strings.Join(h.Values, ", "))
}

// Headers 是有序头域集合，保留插入顺序（SIP 对部分头域顺序有要求）。
type Headers struct {
	list  []*Header       // 保留顺序
	index map[string]int  // 名称 -> list 下标（忽略大小写后的规范名）
}

func NewHeaders() *Headers {
	return &Headers{index: make(map[string]int)}
}

// normalizeName 将紧凑形式转换为规范名，并统一大小写。
func normalizeName(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	if full, ok := shortForms[lower]; ok {
		return full
	}
	// 首字母大写规范化（如 "call-id" -> "Call-Id"，保持 RFC 风格）
	parts := strings.Split(lower, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "-")
}

// Add 追加一个头域值（同名时允许多值，符合 Via 等头域的规范）。
func (h *Headers) Add(name, value string) {
	canon := normalizeName(name)
	if idx, ok := h.index[canon]; ok {
		h.list[idx].Values = append(h.list[idx].Values, value)
	} else {
		h.index[canon] = len(h.list)
		h.list = append(h.list, &Header{Name: canon, Values: []string{value}})
	}
}

// Set 覆盖头域值（只保留最后设置的值）。
func (h *Headers) Set(name, value string) {
	canon := normalizeName(name)
	if idx, ok := h.index[canon]; ok {
		h.list[idx].Values = []string{value}
	} else {
		h.index[canon] = len(h.list)
		h.list = append(h.list, &Header{Name: canon, Values: []string{value}})
	}
}

// Get 返回头域的第一个值；未找到返回 ""。
func (h *Headers) Get(name string) string {
	canon := normalizeName(name)
	if idx, ok := h.index[canon]; ok && len(h.list[idx].Values) > 0 {
		return h.list[idx].Values[0]
	}
	return ""
}

// GetAll 返回头域的所有值。
func (h *Headers) GetAll(name string) []string {
	canon := normalizeName(name)
	if idx, ok := h.index[canon]; ok {
		return h.list[idx].Values
	}
	return nil
}

// Exists 判断头域是否存在。
func (h *Headers) Exists(name string) bool {
	_, ok := h.index[normalizeName(name)]
	return ok
}

// List 按插入顺序返回所有头域（只读）。
func (h *Headers) List() []*Header {
	return h.list
}

// ---- Via 头域解析 ----
// Via: SIP/2.0/UDP pc33.atlanta.com;branch=z9hG4bK776asdhds
// 说明：
//   transport: UDP / TCP / TLS
//   sentBy:    host[:port]
//   branch:    事务标识（必须以 z9hG4bK 开头，RFC 3261 magic cookie）
//   rport:     NAT 穿透使用（RFC 3581）

type Via struct {
	Transport string
	SentBy    string
	Params    map[string]string // branch, rport, received 等
}

func ParseVia(v string) (*Via, error) {
	// SIP/2.0/UDP pc33.atlanta.com;branch=z9hG4bK776asdhds
	parts := strings.SplitN(v, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Via: %q", v)
	}
	proto := strings.Split(parts[0], "/")
	if len(proto) != 3 {
		return nil, fmt.Errorf("invalid Via protocol: %q", parts[0])
	}
	transport := strings.ToUpper(proto[2])

	segments := strings.Split(parts[1], ";")
	sentBy := strings.TrimSpace(segments[0])
	params := make(map[string]string)
	for _, seg := range segments[1:] {
		seg = strings.TrimSpace(seg)
		kv := strings.SplitN(seg, "=", 2)
		if len(kv) == 2 {
			params[strings.ToLower(kv[0])] = kv[1]
		} else {
			params[strings.ToLower(kv[0])] = ""
		}
	}
	return &Via{Transport: transport, SentBy: sentBy, Params: params}, nil
}

func (v *Via) String() string {
	var sb strings.Builder
	sb.WriteString("SIP/2.0/")
	sb.WriteString(v.Transport)
	sb.WriteString(" ")
	sb.WriteString(v.SentBy)
	for k, val := range v.Params {
		sb.WriteString(";")
		sb.WriteString(k)
		if val != "" {
			sb.WriteString("=")
			sb.WriteString(val)
		}
	}
	return sb.String()
}

// ---- Address (From / To / Contact) 解析 ----
// 格式：[display-name] <sip:user@host[:port]>[;tag=xxx]
// 或：sip:user@host（无尖括号形式）

type Address struct {
	DisplayName string
	URI         *URI
	Tag         string   // From/To 的 tag 参数
	Params      map[string]string
}

func ParseAddress(s string) (*Address, error) {
	s = strings.TrimSpace(s)
	addr := &Address{Params: make(map[string]string)}

	var uriStr string
	var paramStr string

	ltIdx := strings.Index(s, "<")
	gtIdx := strings.Index(s, ">")

	if ltIdx >= 0 && gtIdx > ltIdx {
		addr.DisplayName = strings.TrimSpace(s[:ltIdx])
		// 去掉引号
		addr.DisplayName = strings.Trim(addr.DisplayName, "\"")
		uriStr = s[ltIdx+1 : gtIdx]
		paramStr = s[gtIdx+1:]
	} else {
		// 无尖括号：sip:user@host;tag=xxx
		// URI 到第一个 ; 为止
		semiIdx := strings.Index(s, ";")
		if semiIdx >= 0 {
			uriStr = s[:semiIdx]
			paramStr = s[semiIdx:]
		} else {
			uriStr = s
		}
	}

	uri, err := ParseURI(uriStr)
	if err != nil {
		return nil, err
	}
	addr.URI = uri

	// 解析地址级参数（;tag=xxx）
	for _, seg := range strings.Split(strings.TrimPrefix(paramStr, ";"), ";") {
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
		if key == "tag" {
			addr.Tag = val
		} else {
			addr.Params[key] = val
		}
	}
	return addr, nil
}

func (a *Address) String() string {
	var sb strings.Builder
	if a.DisplayName != "" {
		sb.WriteString("\"")
		sb.WriteString(a.DisplayName)
		sb.WriteString("\" ")
	}
	sb.WriteString("<")
	sb.WriteString(a.URI.String())
	sb.WriteString(">")
	if a.Tag != "" {
		sb.WriteString(";tag=")
		sb.WriteString(a.Tag)
	}
	for k, v := range a.Params {
		sb.WriteString(";")
		sb.WriteString(k)
		if v != "" {
			sb.WriteString("=")
			sb.WriteString(v)
		}
	}
	return sb.String()
}

// ---- CSeq 解析 ----
// CSeq: 314159 INVITE

type CSeq struct {
	Seq    uint32
	Method string
}

func ParseCSeq(s string) (*CSeq, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid CSeq: %q", s)
	}
	n, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid CSeq number: %w", err)
	}
	return &CSeq{Seq: uint32(n), Method: strings.ToUpper(parts[1])}, nil
}

func (c *CSeq) String() string {
	return fmt.Sprintf("%d %s", c.Seq, c.Method)
}
