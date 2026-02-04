package raft

// 主要内容是节点操作时的结构体

// 投票请求
type RequestVoteRequest struct {
	Term         int // 候选人任期
	CandidateID  int // 候选人ID
	LastLogIndex int // 候选人最后日志索引
	LastLogTerm  int // 候选人最后日志任期
}

// 投票响应
type RequestVoteResponse struct {
	Term        int  // 当前任期
	VoteGranted bool // 是否投票
}

// 追加日志请求
type AppendEntriesRequest struct {
	Term         int        // Leader的任期
	LeaderID     int        // Leader的ID
	PrevLogIndex int        // 前一个的日志的索引
	PrevLogTerm  int        // 前一个日志的任期
	Entries      []LogEntry // 日志条目
	LeaderCommit int        // Leader提交索引
}

// 追加日志响应
type AppendEntriesResponse struct {
	Term          int  // 当前任期
	Success       bool // 是否成功
	ConflictTerm  int  // 冲突的任期
	ConflictIndex int  // 冲突的第一个索引
}
