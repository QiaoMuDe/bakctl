package pool

import (
	"bytes"
	"sync"
)

// 全局默认缓冲区池实例
var defaultBufferPool = NewBufferPool(1024, 64*1024)

// GetBuffer 从默认缓冲区池获取字节缓冲区
//
// 参数:
//   - capacity: 缓冲区初始容量大小
//
// 返回值:
//   - *bytes.Buffer: 容量至少为capacity的字节缓冲区
func GetBuffer(capacity int) *bytes.Buffer {
	return defaultBufferPool.Get(capacity)
}

// PutBuffer 将字节缓冲区归还到默认缓冲区池
//
// 参数:
//   - buffer: 要归还的字节缓冲区
//
// 说明:
//   - 该函数将字节缓冲区归还到对象池，以便后续复用。
//   - 只有容量不超过64KB的缓冲区才会被归还，以避免对象池占用过多内存。
func PutBuffer(buffer *bytes.Buffer) {
	defaultBufferPool.Put(buffer)
}

// GetEmptyBuffer 从默认缓冲区池获取空的字节缓冲区
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回值:
//   - *bytes.Buffer: 长度为0但容量至少为minCap的字节缓冲区
func GetEmptyBuffer(minCap int) *bytes.Buffer {
	return defaultBufferPool.GetEmpty(minCap)
}

// SetBufferMaxSize 动态调整默认缓冲区池的最大回收大小
//
// 参数:
//   - maxSize: 新的最大回收大小
func SetBufferMaxSize(maxSize int) {
	defaultBufferPool.SetMaxSize(maxSize)
}

// GetBufferMaxSize 获取默认缓冲区池的当前最大回收大小
//
// 返回值:
//   - int: 当前最大回收大小
func GetBufferMaxSize() int {
	return defaultBufferPool.GetMaxSize()
}

// WarmBuffer 预热默认缓冲区池
//
// 参数:
//   - count: 预分配的字节缓冲区数量
//   - capacity: 每个字节缓冲区的容量
func WarmBuffer(count int, capacity int) {
	defaultBufferPool.Warm(count, capacity)
}

// DrainBuffer 清空默认缓冲区池
func DrainBuffer() {
	defaultBufferPool.Drain()
}

// WithBuffer 使用字节缓冲区执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字节缓冲区初始容量大小
//   - fn: 使用字节缓冲区的函数
//
// 返回值:
//   - []byte: 函数执行后缓冲区的字节数据副本
//
// 使用示例:
//
//	data := pool.WithBuffer(1024, func(buf *bytes.Buffer) {
//	    buf.WriteString("Hello")
//	    buf.WriteByte(' ')
//	    buf.WriteString("World")
//	})
func WithBuffer(capacity int, fn func(*bytes.Buffer)) []byte {
	return defaultBufferPool.WithBuffer(capacity, fn)
}

// BufferPool 字节缓冲区对象池，支持自定义配置
type BufferPool struct {
	pool        sync.Pool // 字节缓冲区对象池
	maxSize     int       // 最大回收缓冲区大小
	defaultSize int       // 默认缓冲区大小
}

// NewBufferPool 创建新的字节缓冲区对象池
//
// 参数:
//   - defaultSize: 默认字节缓冲区容量大小
//   - maxSize: 最大回收缓冲区大小，超过此大小的缓冲区不会被回收
//
// 返回值:
//   - *BufferPool: 字节缓冲区对象池实例
func NewBufferPool(defaultSize, maxSize int) *BufferPool {
	if defaultSize <= 0 {
		defaultSize = 1024 // 默认1KB
	}
	if maxSize <= 0 {
		maxSize = 64 * 1024 // 默认64KB
	}

	return &BufferPool{
		maxSize:     maxSize,
		defaultSize: defaultSize,
		pool: sync.Pool{
			New: func() any {
				buffer := &bytes.Buffer{}
				buffer.Grow(defaultSize) // 预分配容量
				return buffer
			},
		},
	}
}

// Get 获取指定容量的字节缓冲区
//
// 参数:
//   - capacity: 需要的字节缓冲区容量大小
//
// 返回:
//   - *bytes.Buffer: 容量至少为capacity的字节缓冲区
//
// 说明:
//   - 返回的字节缓冲区已经重置为空状态，可以直接使用
//   - 底层容量可能大于capacity，来自对象池的复用缓冲区
func (bp *BufferPool) Get(capacity int) *bytes.Buffer {
	buffer, ok := bp.pool.Get().(*bytes.Buffer)
	if !ok {
		// 类型断言失败，创建新的
		buffer = &bytes.Buffer{}
		buffer.Grow(capacity)
		return buffer
	}

	// 重置缓冲区状态
	buffer.Reset()

	// 如果当前容量不足，扩容到所需大小
	if buffer.Cap() < capacity {
		buffer.Grow(capacity - buffer.Cap())
	}

	return buffer
}

// Put 归还字节缓冲区到对象池
//
// 参数:
//   - buffer: 要归还的字节缓冲区
func (bp *BufferPool) Put(buffer *bytes.Buffer) {
	if buffer == nil || buffer.Cap() > bp.maxSize {
		return // 不回收nil或超过最大大小的缓冲区
	}

	// 重置缓冲区状态，清空内容
	buffer.Reset()

	bp.pool.Put(buffer)
}

// GetEmpty 获取指定容量的空字节缓冲区
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回:
//   - *bytes.Buffer: 长度为0但容量至少为minCap的字节缓冲区
//
// 说明:
//   - 适用于需要逐步写入数据的场景
//   - 避免频繁的内存重新分配
func (bp *BufferPool) GetEmpty(minCap int) *bytes.Buffer {
	buffer, ok := bp.pool.Get().(*bytes.Buffer)
	if !ok {
		// 类型断言失败，创建新的
		buffer = &bytes.Buffer{}
		buffer.Grow(minCap)
		return buffer
	}

	// 重置缓冲区状态
	buffer.Reset()

	// 如果当前容量不足，扩容到所需大小
	if buffer.Cap() < minCap {
		buffer.Grow(minCap - buffer.Cap())
	}

	return buffer
}

// SetMaxSize 动态调整最大回收缓冲区大小
//
// 参数:
//   - maxSize: 新的最大回收大小
//
// 说明:
//   - 运行时动态调整配置
//   - 如果新的maxSize小于当前值，建议调用Drain()清空对象池
func (bp *BufferPool) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		maxSize = 64 * 1024 // 默认64KB
	}
	bp.maxSize = maxSize
}

// GetMaxSize 获取当前最大回收缓冲区大小
//
// 返回:
//   - int: 当前最大回收大小
func (bp *BufferPool) GetMaxSize() int {
	return bp.maxSize
}

// Warm 预热对象池
//
// 参数:
//   - count: 预分配的字节缓冲区数量
//   - capacity: 每个字节缓冲区的容量
//
// 说明:
//   - 在应用启动时调用，预分配指定数量的字节缓冲区
//   - 减少冷启动时的内存分配延迟
//   - 提升初期性能表现
func (bp *BufferPool) Warm(count int, capacity int) {
	if count <= 0 || capacity <= 0 {
		return
	}

	// 预分配指定数量的字节缓冲区
	buffers := make([]*bytes.Buffer, count)
	for i := 0; i < count; i++ {
		buffer := &bytes.Buffer{}
		buffer.Grow(capacity)
		buffers[i] = buffer
	}

	// 立即归还到对象池进行预热
	for _, buffer := range buffers {
		bp.Put(buffer)
	}
}

// Drain 清空对象池中的所有字节缓冲区
//
// 说明:
//   - 清空当前对象池中的所有字节缓冲区
//   - 重新创建sync.Pool，释放可能占用的大量内存
//   - 适用于内存紧张或需要重置对象池状态的场景
func (bp *BufferPool) Drain() {
	// 创建新的sync.Pool替换旧的
	bp.pool = sync.Pool{
		New: func() any {
			buffer := &bytes.Buffer{}
			buffer.Grow(bp.defaultSize) // 预分配容量
			return buffer
		},
	}
}

// WithBuffer 使用字节缓冲区执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字节缓冲区初始容量大小
//   - fn: 使用字节缓冲区的函数
//
// 返回值:
//   - []byte: 函数执行后缓冲区的字节数据副本
//
// 说明:
//   - 自动从对象池获取字节缓冲区
//   - 执行用户提供的函数
//   - 获取缓冲区字节数据的副本
//   - 自动归还字节缓冲区到对象池
//   - 即使函数发生panic也会正确归还资源
func (bp *BufferPool) WithBuffer(capacity int, fn func(*bytes.Buffer)) []byte {
	buffer := bp.Get(capacity)
	defer bp.Put(buffer)

	fn(buffer)
	// 返回字节数据的副本，避免在归还后访问
	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())
	return result
}
