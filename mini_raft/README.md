# Mini Raft - Golang Raft 协议实现

这是一个使用 Golang 实现的简化版 Raft 共识协议，包含了 Raft 的核心功能。

## 功能特性

### 核心组件

1. **节点状态管理**
   - Follower（跟随者）
   - Candidate（候选人）
   - Leader（领导者）

2. **领导者选举**

3. **日志复制**

4. **持久化状态**


## 工作流程

### 1. 领导者选举

1. 所有节点初始状态为 Follower
2. 如果 Follower 在选举超时内没有收到心跳，转为 Candidate
3. Candidate 增加任期号，投票给自己，并向其他节点请求投票
4. 获得多数投票的 Candidate 成为 Leader
5. Leader 定期发送心跳维持权威

### 2. 日志复制

1. 客户端向 Leader 提交命令
2. Leader 将命令追加到本地日志
3. Leader 通过 AppendEntries RPC 将日志复制到 Follower
4. 当多数节点复制成功后，Leader 提交日志
5. Leader 通知 Follower 提交日志

### 3. 安全性保证

- **选举安全性**：一个任期内最多一个 Leader
- **日志匹配性**：如果两个日志在相同索引处的条目任期号相同，则它们之前的所有条目都相同
- **Leader 完整性**：如果某个日志条目在某个任期被提交，那么这个条目必然出现在更高任期的 Leader 日志中


## 许可证

MIT License
