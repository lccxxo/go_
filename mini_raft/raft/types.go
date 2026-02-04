package raft

type NodeState int

const (
	Follower  NodeState = iota // 追随者 被动接收leader发送的日志
	Candidate                  // 候选人 竞选leader
	Leader                     // 领导者 处理客户端请求 并处理内部日志
)

// LogEntry 日志条目
// 记录了每个节点每个任期内都发了什么命令 做了什么 主要目的是为了保持一致性
type LogEntry struct {
	Term    int // 任期号
	Command any // 操作命令
}
