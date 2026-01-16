package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

type Key struct {
	RawKey  []byte // 原始键
	Version uint64 // 版本号
}

// 将Key复合键转为字节切片
func (k *Key) Bytes() []byte {
	// 版本号 + 原始键长度 + 原始键内容
	buf := make([]byte, len(k.RawKey)+8+4)
	binary.BigEndian.PutUint64(buf[0:8], k.Version)
	binary.BigEndian.PutUint32(buf[8:12], uint32(len(k.RawKey)))
	copy(buf[12:], k.RawKey)
	return buf
}

// 将字节切片转为Key复合键
func DecodeKey(data []byte) *Key {
	if len(data) < 12 {
		return nil
	}

	version := binary.BigEndian.Uint64(data[0:8])
	keyLen := binary.BigEndian.Uint32(data[8:12])
	if len(data) < 12+int(keyLen) {
		return nil
	}

	rawKey := make([]byte, keyLen)
	copy(rawKey, data[12:12+keyLen])
	return &Key{RawKey: rawKey, Version: version}
}

type KeyValue struct {
	Key   []byte
	Value []byte
}

type KVEngine interface {
	Get(key []byte) ([]byte, bool)
	Set(key, value []byte)
	Delete(key []byte)
	Iterate(prefix []byte) []KeyValue
}

// 内存中键值对存储实现
type MemoryKV struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMemoryKV() *MemoryKV {
	return &MemoryKV{
		data: make(map[string][]byte),
	}
}

func (kv *MemoryKV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, exists := kv.data[string(key)]
	if !exists {
		return nil, false
	}
	return val, true
}

func (kv *MemoryKV) Set(key, value []byte) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[string(key)] = value
}

func (kv *MemoryKV) Delete(key []byte) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.data, string(key))
}

func (kv *MemoryKV) Iterate(prefix []byte) []KeyValue {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	var result []KeyValue
	prefixStr := string(prefix)

	for k, v := range kv.data {
		if prefix == nil && len(k) >= len(prefixStr) && k[:len(prefixStr)] == prefixStr {
			result = append(result, KeyValue{
				Key:   []byte(k),
				Value: v,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return bytes.Compare(result[i].Key, result[j].Key) < 0
	})

	return result
}

// 活跃事务信息
type ActiveTxn struct {
	ModifiedKeys [][]byte
	mu           sync.RWMutex
}

type MVCC struct {
	kv         KVEngine
	version    uint64
	activeTxns *sync.Map // map[uint64]*ActiveTxn
}

func NewMVCC(kv KVEngine) *MVCC {
	return &MVCC{
		kv:         kv,
		version:    0,
		activeTxns: &sync.Map{},
	}
}

// 开始事务
func (m *MVCC) BeginTransaction() *Transaction {
	version := atomic.AddUint64(&m.version, 1)

	var activeXid []uint64
	m.activeTxns.Range(func(key, value interface{}) bool {
		activeXid = append(activeXid, key.(uint64))
		return true
	})

	// 注册新事务
	txn := &ActiveTxn{
		ModifiedKeys: make([][]byte, 0),
	}
	m.activeTxns.Store(version, txn)

	return &Transaction{
		mvcc:      m,
		version:   version,
		activeXid: activeXid,
		localData: make(map[string][]byte),
	}
}

// Transaction MVCC事务
type Transaction struct {
	mvcc      *MVCC
	version   uint64
	activeXid []uint64
	localData map[string][]byte // 本地修改缓存
}

// 写入数据
func (tx *Transaction) Set(key, value []byte) error {
	// 检查写写冲突
	latestVersion, err := tx.findLatestVersion(key)
	if err != nil {
		return err
	}

	if latestVersion > 0 && !tx.isVisible(latestVersion) {
		return fmt.Errorf("serialization error, try again")
	}

	// 记录修改的键
	if txn, ok := tx.mvcc.activeTxns.Load(latestVersion); ok {
		txn.(*ActiveTxn).mu.Lock()
		txn.(*ActiveTxn).ModifiedKeys = append(txn.(*ActiveTxn).ModifiedKeys, key)
		txn.(*ActiveTxn).mu.Unlock()
	}

	// 写入本地存储
	tx.localData[string(key)] = value

	// 创建复合键并写入存储
	compositeKey := &Key{
		RawKey:  key,
		Version: tx.version,
	}
	tx.mvcc.kv.Set(compositeKey.Bytes(), value)

	return nil
}

// 删除数据
func (tx *Transaction) Delete(key []byte) error {
	return tx.Set(key, nil)
}

// 获取数据
func (tx *Transaction) Get(key []byte) ([]byte, bool) {
	// 先查看是否命中本地缓存
	if val, ok := tx.localData[string(key)]; ok {
		return val, val != nil
	}

	// 在存储中查找所有版本
	allVersions := tx.findAllVersions(key)

	// 找到对当前事务可见的最新版本
	for _, kv := range allVersions {
		keyObj := DecodeKey(kv.Key)
		if keyObj != nil && tx.isVisible(keyObj.Version) {
			return kv.Value, kv.Value != nil
		}
	}

	return nil, false
}

// 从活跃事务表中移除
func (tx *Transaction) Commit() {
	tx.mvcc.activeTxns.Delete(tx.version)
}

// 回滚事务
func (tx *Transaction) Rollback() {
	// 删除本事务写入的所有数据
	if txn, ok := tx.mvcc.activeTxns.Load(tx.version); ok {
		txn.(*ActiveTxn).mu.RLock()
		modifiedKeys := make([][]byte, len(txn.(*ActiveTxn).ModifiedKeys))
		copy(modifiedKeys, txn.(*ActiveTxn).ModifiedKeys)
		txn.(*ActiveTxn).mu.RUnlock()

		for _, key := range modifiedKeys {
			compositeKey := &Key{
				RawKey:  key,
				Version: tx.version,
			}
			tx.mvcc.activeTxns.Delete(compositeKey.Bytes())
		}
	}

	// 从活跃事务表中删除
	tx.mvcc.activeTxns.Delete(tx.version)
}

// 打印所有可见数据
func (tx *Transaction) PrintAll() {
	records := make(map[string][]byte)

	// 获取存储中的所有数据
	allData := tx.mvcc.kv.Iterate(nil)

	// 过滤出对当前事务可见的数据
	for _, kv := range allData {
		keyObj := DecodeKey(kv.Key)
		if keyObj != nil && tx.isVisible(keyObj.Version) {
			// 对于同一个raw_key,只保存最新可见版本
			rawKeyStr := string(keyObj.RawKey)
			if existing, exists := records[rawKeyStr]; !exists || keyObj.Version > tx.getVersionOfRecord(existing) {
				records[rawKeyStr] = kv.Value
			}
		}
	}

	// 排序并打印
	var keys []string
	for k := range records {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	fmt.Print("可见数据: ")
	for _, k := range keys {
		v := records[k]
		if v != nil {
			fmt.Printf("%s=%s ", k, string(v))
		}
	}
	fmt.Println()
}

// 辅助方法

// 获取最后的版本
func (tx *Transaction) findLatestVersion(key []byte) (uint64, error) {
	allVersions := tx.findAllVersions(key)
	if len(allVersions) == 0 {
		return 0, nil
	}

	// 最后一个元素是版本号最大的，因为是按照版本号排序进行返回数据
	lastKey := DecodeKey(allVersions[len(allVersions)-1].Key)
	if lastKey != nil {
		return lastKey.Version, nil
	}

	return 0, nil
}

// 是否可见
func (tx *Transaction) isVisible(version uint64) bool {
	// 如果版本属于其他活跃事务，则不可见
	for _, active := range tx.activeXid {
		if active == version {
			return false
		}
	}

	// 只能看到版本号 <= 当前事务版本号的数据
	return version <= tx.version
}

// 获取所有版本
func (tx *Transaction) findAllVersions(key []byte) []KeyValue {
	// 获取指定原始键的所有版本
	// 由于编码方式是 version + key, 使用前缀匹配去查找
	// 简单遍历所有数据

	allData := tx.mvcc.kv.Iterate(nil)
	var result []KeyValue

	for _, kv := range allData {
		keyObj := DecodeKey(kv.Key)
		if keyObj != nil && bytes.Equal(keyObj.RawKey, key) {
			result = append(result, kv)
		}
	}

	// 按照版本号排序（从小到大）
	sort.Slice(result, func(i, j int) bool {
		keyI := DecodeKey(result[i].Key)
		keyJ := DecodeKey(result[j].Key)
		for keyI != nil && keyJ != nil {
			return keyI.Version < keyJ.Version
		}
		return false
	})

	return result
}

// 从存储中拿到版本
func (tx *Transaction) getVersionOfRecord(keyBytes []byte) uint64 {
	keyObj := DecodeKey(keyBytes)
	if keyObj != nil {
		return keyObj.Version
	}
	return 0
}

func main() {
	fmt.Println("=== MVCC 实现测试 ===")
	// 创建存储引擎和MVCC
	kv := NewMemoryKV()
	mvcc := NewMVCC(kv)

	tx0 := mvcc.BeginTransaction()
	tx0.Set([]byte("a"), []byte("a1"))
	tx0.Set([]byte("b"), []byte("b1"))
	tx0.Set([]byte("c"), []byte("c1"))
	tx0.Set([]byte("d"), []byte("d1"))
	tx0.Set([]byte("e"), []byte("e1"))
	tx0.Commit()
	fmt.Println("T0 提交了初始数据")

	// 开启一个事务
	tx1 := mvcc.BeginTransaction()
	tx1.Set([]byte("a"), []byte("a2"))
	tx1.Set([]byte("e"), []byte("e2"))
	fmt.Println("T1 修改了 a 和 e（未提交）")
	tx1.PrintAll()

	// 开启一个新的事务
	tx2 := mvcc.BeginTransaction()
	tx2.Delete([]byte("b"))
	fmt.Println("T2 删除了 b（未提交）")
	tx2.PrintAll()

	// 提交 T1
	tx1.Commit()
	fmt.Println("T1 提交后：")
	fmt.Print("T2 看到的数据: ")
	tx2.PrintAll()

	// 再开启一个新的事务
	tx3 := mvcc.BeginTransaction()
	fmt.Print("T3 看到的数据: ")
	tx3.PrintAll()

	// T3 写新的数据
	tx3.Set([]byte("f"), []byte("f1"))
	fmt.Println("T3 写入了 f=f1")

	// T2 写同样的数据，会冲突
	err := tx2.Set([]byte("f"), []byte("f2"))
	if err != nil {
		fmt.Printf("T2 写入 f 时发生冲突: %v\n", err)
	}

	// 打印最终状态
	fmt.Println("\n=== 最终状态 ===")
	txFinal := mvcc.BeginTransaction()
	txFinal.PrintAll()

	// 回滚未提交的事务
	tx2.Rollback()
	tx3.Rollback()
}
