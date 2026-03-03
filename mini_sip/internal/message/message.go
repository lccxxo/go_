package message

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Method 枚举 SIP 请求方法（RFC 3261 §27.4）
type Method string

const (
	MethodINVITE   Method = "INVITE"
	MethodACK      Method = "ACK"
	MethodBYE      Method = "BYE"
	MethodCANCEL   Method = "CANCEL"
	MethodREGISTER Method = "REGISTER"
	MethodOPTIONS  Method = "OPTIONS"
	MethodSUBSCRIBE Method = "SUBSCRIBE"
	MethodNOTIFY   Method = "NOTIFY"
	MethodREFER    Method = "REFER"
	MethodMESSAGE  Method = "MESSAGE"
	MethodINFO     Method = "INFO"
	MethodUPDATE   Method = "UPDATE"
	MethodPRACK    Method = "PRACK"
)

// SIP 版本常量
const SIPVersion = "SIP/2.0"

// ---- Request ----

// Request 表示一条 SIP 请求消息。
//
// 请求行格式：Method Request-URI SIP/2.0
//
// 常见流程：
//   - REGISTER：UA 向注册服务器注册自己的联系地址
//   - INVITE / ACK / BYE：建立和终止会话（三次握手）
//   - CANCEL：取消尚未完成的 INVITE
//   - OPTIONS：查询对端能力
type Request struct {
	Method     Method
	RequestURI *URI
	Headers    *Headers
	Body       []byte
}

// NewRequest 创建一条带基础头域的请求。
func NewRequest(method Method, requestURI *URI) *Request {
	r := &Request{
		Method:     method,
		RequestURI: requestURI,
		Headers:    NewHeaders(),
	}
	return r
}

// String 序列化为 SIP 文本格式。
func (r *Request) String() string {
	var sb strings.Builder
	// 请求行
	sb.WriteString(string(r.Method))
	sb.WriteString(" ")
	sb.WriteString(r.RequestURI.String())
	sb.WriteString(" ")
	sb.WriteString(SIPVersion)
	sb.WriteString("\r\n")
	// 确保 Content-Length 正确
	if len(r.Body) > 0 {
		r.Headers.Set(HeaderContentLen, strconv.Itoa(len(r.Body)))
	} else {
		r.Headers.Set(HeaderContentLen, "0")
	}
	// 头域
	for _, h := range r.Headers.List() {
		for _, v := range h.Values {
			sb.WriteString(h.Name)
			sb.WriteString(": ")
			sb.WriteString(v)
			sb.WriteString("\r\n")
		}
	}
	sb.WriteString("\r\n")
	if len(r.Body) > 0 {
		sb.Write(r.Body)
	}
	return sb.String()
}

// ---- Response ----

// Response 表示一条 SIP 响应消息。
//
// 状态码分类（RFC 3261 §21）：
//   - 1xx：临时响应（Trying, Ringing）
//   - 2xx：成功（OK）
//   - 3xx：重定向
//   - 4xx：客户端错误（Bad Request, Unauthorized, Not Found）
//   - 5xx：服务器错误
//   - 6xx：全局错误
type Response struct {
	StatusCode int
	Reason     string
	Headers    *Headers
	Body       []byte
}

// 常见状态码
const (
	StatusTrying           = 100
	StatusRinging          = 180
	StatusSessionProgress  = 183
	StatusOK               = 200
	StatusAccepted         = 202
	StatusMultipleChoices  = 300
	StatusBadRequest       = 400
	StatusUnauthorized     = 401
	StatusForbidden        = 403
	StatusNotFound         = 404
	StatusMethodNotAllowed = 405
	StatusRequestTimeout   = 408
	StatusBusyHere         = 486
	StatusServerError      = 500
	StatusDecline          = 603
)

var statusReasons = map[int]string{
	100: "Trying",
	180: "Ringing",
	183: "Session Progress",
	200: "OK",
	202: "Accepted",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Moved Temporarily",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	408: "Request Timeout",
	481: "Call/Transaction Does Not Exist",
	486: "Busy Here",
	487: "Request Terminated",
	488: "Not Acceptable Here",
	500: "Server Internal Error",
	503: "Service Unavailable",
	603: "Decline",
}

// ReasonPhrase 返回状态码对应的标准原因短语。
func ReasonPhrase(code int) string {
	if r, ok := statusReasons[code]; ok {
		return r
	}
	return "Unknown"
}

// NewResponse 创建一条响应（自动填充原因短语）。
func NewResponse(code int) *Response {
	return &Response{
		StatusCode: code,
		Reason:     ReasonPhrase(code),
		Headers:    NewHeaders(),
	}
}

// String 序列化为 SIP 文本格式。
func (r *Response) String() string {
	var sb strings.Builder
	// 状态行
	sb.WriteString(SIPVersion)
	sb.WriteString(" ")
	sb.WriteString(strconv.Itoa(r.StatusCode))
	sb.WriteString(" ")
	sb.WriteString(r.Reason)
	sb.WriteString("\r\n")
	// Content-Length
	if len(r.Body) > 0 {
		r.Headers.Set(HeaderContentLen, strconv.Itoa(len(r.Body)))
	} else {
		r.Headers.Set(HeaderContentLen, "0")
	}
	for _, h := range r.Headers.List() {
		for _, v := range h.Values {
			sb.WriteString(h.Name)
			sb.WriteString(": ")
			sb.WriteString(v)
			sb.WriteString("\r\n")
		}
	}
	sb.WriteString("\r\n")
	if len(r.Body) > 0 {
		sb.Write(r.Body)
	}
	return sb.String()
}

// ---- Parser ----

// Parse 将原始字节解析为 Request 或 Response。
// 返回值类型可断言为 *Request 或 *Response。
func Parse(data []byte) (interface{}, error) {
	// 分离 header 部分和 body 部分（\r\n\r\n 或 \n\n）
	var headerBytes, body []byte
	if idx := bytes.Index(data, []byte("\r\n\r\n")); idx >= 0 {
		headerBytes = data[:idx]
		body = data[idx+4:]
	} else if idx := bytes.Index(data, []byte("\n\n")); idx >= 0 {
		headerBytes = data[:idx]
		body = data[idx+2:]
	} else {
		headerBytes = data
	}

	lines := splitLines(string(headerBytes))
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty SIP message")
	}

	startLine := lines[0]
	headers, err := parseHeaders(lines[1:])
	if err != nil {
		return nil, err
	}

	// 根据 Content-Length 截断 body
	if cl := headers.Get(HeaderContentLen); cl != "" {
		n, err := strconv.Atoi(strings.TrimSpace(cl))
		if err == nil && n >= 0 && n <= len(body) {
			body = body[:n]
		}
	}

	if strings.HasPrefix(startLine, "SIP/") {
		// 响应
		resp, err := parseResponseLine(startLine)
		if err != nil {
			return nil, err
		}
		resp.Headers = headers
		resp.Body = body
		return resp, nil
	}
	// 请求
	req, err := parseRequestLine(startLine)
	if err != nil {
		return nil, err
	}
	req.Headers = headers
	req.Body = body
	return req, nil
}

func splitLines(s string) []string {
	// 统一换行符
	s = strings.ReplaceAll(s, "\r\n", "\n")
	raw := strings.Split(s, "\n")
	var out []string
	for i := 0; i < len(raw); i++ {
		line := raw[i]
		// 处理头域折叠（continuation lines）
		for i+1 < len(raw) && len(raw[i+1]) > 0 && (raw[i+1][0] == ' ' || raw[i+1][0] == '\t') {
			i++
			line += " " + strings.TrimSpace(raw[i])
		}
		out = append(out, line)
	}
	return out
}

func parseHeaders(lines []string) (*Headers, error) {
	h := NewHeaders()
	for _, line := range lines {
		if line == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			return nil, fmt.Errorf("invalid header line: %q", line)
		}
		name := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])
		h.Add(name, value)
	}
	return h, nil
}

func parseRequestLine(line string) (*Request, error) {
	// METHOD Request-URI SIP/2.0
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %q", line)
	}
	if parts[2] != SIPVersion {
		return nil, fmt.Errorf("unsupported SIP version: %q", parts[2])
	}
	uri, err := ParseURI(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid Request-URI: %w", err)
	}
	return &Request{Method: Method(parts[0]), RequestURI: uri}, nil
}

func parseResponseLine(line string) (*Response, error) {
	// SIP/2.0 200 OK
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid response line: %q", line)
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %w", err)
	}
	reason := ""
	if len(parts) == 3 {
		reason = parts[2]
	}
	return &Response{StatusCode: code, Reason: reason}, nil
}
