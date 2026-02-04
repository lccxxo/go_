package raft

import "sync"

// raft节点结构
type RaftNode struct {
	// 持久化
	id          int        // 节点的唯一标识
	currentTerm int        // 当前任期号
	votedFor    *int       // 投票给了谁
	log         []LogEntry // 日志条目数组

	// 易损性
	commitIndex int       // 已知的已经提交的最高日志索引
	lastApplied int       // 已应用到状态机的最高日志索引
	state       NodeState // 当前节点状态

	// 领导者状态
	nextIndex  map[int]int // key: 节点ID value: 下次发送的日志索引
	matchIndex map[int]int // key: 节点ID value: 已知已复制的最高索引 重新选举使用

	// 网络
	peers []int // 集群网络中其他节点的ID

	mu sync.RWMutex
}

// 创建新的raft节点
func NewRaftNode(id int, peers []int) *RaftNode {
	return &RaftNode{
		id:          id,
		currentTerm: 0,
		votedFor:    nil,
		log:         []LogEntry{},
		commitIndex: 0,
		lastApplied: 0,
		state:       Follower,
		nextIndex:   make(map[int]int),
		matchIndex:  make(map[int]int),
		peers:       peers,
	}
}
