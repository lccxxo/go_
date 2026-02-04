package raft

// 用于状态转换

// BecomeFollower 转换为 Follower 状态
// term: 新的任期号
// leaderID: 当前 Leader 的 ID
func (n *RaftNode) BecomeFollower(term int, leaderID *int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Follower
	n.currentTerm = term
	// 转换为 Follower 时重置投票状态
	// 只有在新任期时才重置 votedFor
	if term > n.currentTerm {
		n.votedFor = nil
	}
}

// BecomeCandidate 转换为 Candidate 状态并开始选举
func (n *RaftNode) BecomeCandidate() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Candidate
	n.currentTerm++    // 增加任期号
	n.votedFor = &n.id // 投票给自己
}

// BecomeLeader 转换为 Leader 状态
func (n *RaftNode) BecomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.state = Leader

	// 初始化 Leader 状态
	// nextIndex: 初始化为 Leader 最后日志索引 + 1
	lastLogIndex := len(n.log)
	for _, peerID := range n.peers {
		n.nextIndex[peerID] = lastLogIndex
		n.matchIndex[peerID] = 0
	}
}
