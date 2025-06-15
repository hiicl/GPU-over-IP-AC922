package scheduler

import (
    "errors"
    "sync"
    "time"

    "juice-refactor/pkg/util"
)

// scheduler 包提供GPU资源调度功能，管理GPU资源的占用和释放
// 通过互斥锁确保并发安全，并实现超时自动释放机制防止死锁
// 主要功能：
//   - 安全地占用和释放GPU资源
//   - 提供GPU占用状态查询
//   - 超时自动释放机制防止资源死锁

// Scheduler 结构体管理GPU资源调度
// mu: 互斥锁，保护inUse映射的并发访问
// inUse: 记录GPU占用状态的映射表（key: GPU UUID, value: 是否被占用）
// timeout: 资源占用超时时间（超过此时间未释放将自动释放）
type Scheduler struct {
    mu      sync.Mutex     // 互斥锁，保护inUse映射的并发访问
    inUse   map[string]bool // key: GPU UUID，value: 是否被占用
    timeout time.Duration   // 资源占用超时时间（单位：duration）
}

// NewScheduler 创建并初始化一个新的调度器实例
// timeout: 资源锁自动释放的超时时间（防止死锁）
// 返回初始化后的Scheduler指针
func NewScheduler(timeout time.Duration) *Scheduler {
    return &Scheduler{
        inUse:   make(map[string]bool), // 初始化GPU占用状态映射
        timeout: timeout,                // 设置超时时间
    }
}

// Acquire 尝试占用指定的GPU资源
// uuid: 要占用的GPU的唯一标识符
// 返回值：error - 如果GPU已被占用则返回错误，否则返回nil
// 注意：占用成功后会自动启动超时释放协程（防止死锁）
func (s *Scheduler) Acquire(uuid string) error {
    // 加锁确保并发安全
    s.mu.Lock()
    defer s.mu.Unlock() // 确保函数返回时解锁

    // 检查GPU是否已被占用
    if s.inUse[uuid] {
        return errors.New("GPU already in use")
    }

    // 标记GPU为已占用状态
    s.inUse[uuid] = true

    // 启动goroutine在超时后自动释放资源（防止死锁）
    go func() {
        // 等待超时时间
        time.Sleep(s.timeout)
        // 超时后释放资源
        s.Release(uuid)
    }()

    // 记录资源获取日志
    util.Log.Infof("GPU %s acquired", uuid)
    return nil
}

// Release 释放指定的GPU资源
// uuid: 要释放的GPU的唯一标识符
// 注意：如果GPU未被占用，则不执行任何操作
func (s *Scheduler) Release(uuid string) {
    // 加锁确保并发安全
    s.mu.Lock()
    defer s.mu.Unlock() // 确保函数返回时解锁

    // 检查GPU是否处于占用状态
    if s.inUse[uuid] {
        // 从占用映射中删除该GPU（释放资源）
        delete(s.inUse, uuid)
        // 记录资源释放日志
        util.Log.Infof("GPU %s released", uuid)
    }
}

// IsInUse 检查指定GPU是否被占用
// uuid: 要检查的GPU的唯一标识符
// 返回值：bool - true表示GPU已被占用，false表示可用
func (s *Scheduler) IsInUse(uuid string) bool {
    // 加锁确保并发安全
    s.mu.Lock()
    defer s.mu.Unlock() // 确保函数返回时解锁

    // 返回GPU的占用状态
    return s.inUse[uuid]
}
