package raft_kv

// StateMachine 状态机接口
type StateMachine interface {
	Apply(cmd []byte) ([]byte, error)
}

type Server struct {
	// 服务唯一性id
	id uint64
	// 用于rpc调用的地址
	address string
	// 状态机
	statemachine StateMachine
}
