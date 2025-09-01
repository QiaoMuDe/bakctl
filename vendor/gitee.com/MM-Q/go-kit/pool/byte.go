// Package pool 提供高性能对象池管理功能，通过复用对象优化内存使用。
//
// 该包实现了基于 sync.Pool 的多种对象池，用于减少频繁的内存分配和回收。
// 通过复用对象，可以显著提升应用程序的性能，特别是在高并发场景下。
//
// 主要功能：
//   - 字节切片对象池管理
//   - 动态大小对象获取
//   - 自动内存回收控制
//   - 防止内存泄漏的大小限制
//   - 支持多种对象类型的池化
//
// 性能优化：
//   - 使用 sync.Pool 减少 GC 压力
//   - 支持不同大小的对象需求
//   - 自动限制大对象回收
//   - 预热机制提升冷启动性能
//
// 使用示例：
//
//	// 获取字节缓冲区
//	buf := pool.GetByte(64 * 1024)
//
//	// 使用缓冲区进行文件操作
//	_, err := io.CopyBuffer(dst, src, buf)
//
//	// 归还缓冲区到对象池
//	pool.PutByte(buf)
package pool

import "sync"

// 全局默认对象池实例
var defaultPool = NewBytePool(32*1024, 1024*1024)

// GetByte 从默认字节池获取指定大小的缓冲区
//
// 参数:
//   - size: 缓冲区容量大小
//
// 返回值:
//   - []byte: 长度为size，容量至少为size的缓冲区
func GetByte(size int) []byte {
	return defaultPool.Get(size)
}

// PutByte 将缓冲区归还到默认字节池
//
// 参数:
//   - buffer: 要归还的缓冲区
//
// 说明:
//   - 该函数将缓冲区归还到对象池，以便后续复用。
//   - 只有容量不超过1MB的缓冲区才会被归还，以避免对象池占用过多内存。
func PutByte(buffer []byte) {
	defaultPool.Put(buffer)
}

// GetEmptyByte 从默认字节池获取空缓冲区
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回值:
//   - []byte: 长度为0但容量至少为minCap的缓冲区切片
func GetEmptyByte(minCap int) []byte {
	return defaultPool.GetEmpty(minCap)
}

// SetByteMaxSize 动态调整默认字节池的最大回收大小
//
// 参数:
//   - maxSize: 新的最大回收大小
func SetByteMaxSize(maxSize int) {
	defaultPool.SetMaxSize(maxSize)
}

// GetByteMaxSize 获取默认字节池的当前最大回收大小
//
// 返回值:
//   - int: 当前最大回收大小
func GetByteMaxSize() int {
	return defaultPool.GetMaxSize()
}

// WarmByte 预热默认字节池
//
// 参数:
//   - count: 预分配的缓冲区数量
//   - size: 每个缓冲区的大小
func WarmByte(count int, size int) {
	defaultPool.Warm(count, size)
}

// DrainByte 清空默认字节池
func DrainByte() {
	defaultPool.Drain()
}

// WithByte 使用字节切片执行函数，自动管理获取和归还
//
// 参数:
//   - size: 字节切片初始大小
//   - fn: 使用字节切片的函数
//
// 返回值:
//   - []byte: 函数执行后字节切片的数据副本
//
// 使用示例:
//
//	data := pool.WithByte(1024, func(buf []byte) {
//	    copy(buf, []byte("Hello World"))
//	    // 可以直接操作buf进行读写
//	})
func WithByte(size int, fn func([]byte)) []byte {
	return defaultPool.WithByte(size, fn)
}

// WithEmptyByte 使用空字节切片执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字节切片初始容量
//   - fn: 使用字节切片的函数，通过append等操作构建数据
//
// 返回值:
//   - []byte: 函数执行后字节切片的数据副本
//
// 使用示例:
//
//	data := pool.WithEmptyByte(1024, func(buf []byte) []byte {
//	    buf = append(buf, []byte("Hello")...)
//	    buf = append(buf, ' ')
//	    buf = append(buf, []byte("World")...)
//	    return buf
//	})
func WithEmptyByte(capacity int, fn func([]byte) []byte) []byte {
	return defaultPool.WithEmptyByte(capacity, fn)
}

// BytePool 字节切片对象池，支持自定义配置
type BytePool struct {
	pool        sync.Pool // 缓冲区对象池
	maxSize     int       // 最大回收缓冲区大小
	defaultSize int       // 默认缓冲区大小
}

// NewBytePool 创建新的字节切片对象池
//
// 参数:
//   - defaultSize: 默认缓冲区大小
//   - maxSize: 最大回收缓冲区大小，超过此大小的缓冲区不会被回收
//
// 返回值:
//   - *BytePool: 字节切片对象池实例
func NewBytePool(defaultSize, maxSize int) *BytePool {
	if defaultSize <= 0 {
		defaultSize = 32 * 1024 // 默认32KB
	}
	if maxSize <= 0 {
		maxSize = 1024 * 1024 // 默认1MB
	}

	return &BytePool{
		maxSize:     maxSize,
		defaultSize: defaultSize,
		pool: sync.Pool{
			New: func() any {
				buf := make([]byte, 0, defaultSize) // 长度0，容量defaultSize
				return &buf                         // 返回指针避免装箱
			},
		},
	}
}

// Get 获取指定容量的缓冲区
//
// 参数:
//   - size: 需要的缓冲区容量大小
//
// 返回:
//   - []byte: 长度为size，容量至少为size的缓冲区切片
//
// 说明:
//   - 返回的缓冲区长度等于请求的size，可以直接使用
//   - 底层容量可能大于size，来自对象池的复用缓冲区
func (bp *BytePool) Get(size int) []byte {
	bufPtr, ok := bp.pool.Get().(*[]byte)
	if !ok {
		// 类型断言失败，创建新的
		buf := make([]byte, 0, size)
		return buf[:size]
	}

	// 获取缓冲区
	buffer := *bufPtr

	// 缓冲区容量不足，创建新的
	if cap(buffer) < size {
		buf := make([]byte, 0, size)
		return buf[:size]
	}

	return buffer[:size]
}

// Put 归还缓冲区到对象池
//
// 参数:
//   - buffer: 要归还的缓冲区
func (bp *BytePool) Put(buffer []byte) {
	if buffer == nil || cap(buffer) > bp.maxSize {
		return // 不回收nil或超过最大大小的缓冲区
	}

	// 清空缓冲区内容
	buffer = buffer[:0]

	bp.pool.Put(&buffer) // 传入指针避免装箱分配
}

// GetEmpty 获取指定容量的空缓冲区
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回:
//   - []byte: 长度为0但容量至少为minCap的缓冲区切片
//
// 说明:
//   - 适用于需要使用append操作逐步构建数据的场景
//   - 避免频繁的内存重新分配
func (bp *BytePool) GetEmpty(minCap int) []byte {
	bufPtr, ok := bp.pool.Get().(*[]byte)
	if !ok {
		// 类型断言失败，创建新的
		return make([]byte, 0, minCap)
	}

	// 获取缓冲区
	buffer := *bufPtr

	// 缓冲区容量不足，创建新的
	if cap(buffer) < minCap {
		return make([]byte, 0, minCap)
	}

	return buffer[:0] // 返回长度为0但保持容量的切片
}

// SetMaxSize 动态调整最大回收缓冲区大小
//
// 参数:
//   - maxSize: 新的最大回收大小
//
// 说明:
//   - 运行时动态调整配置
//   - 如果新的maxSize小于当前值，建议调用Drain()清空对象池
func (bp *BytePool) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		maxSize = 1024 * 1024 // 默认1MB
	}
	bp.maxSize = maxSize
}

// GetMaxSize 获取当前最大回收缓冲区大小
//
// 返回:
//   - int: 当前最大回收大小
func (bp *BytePool) GetMaxSize() int {
	return bp.maxSize
}

// Warm 预热对象池
//
// 参数:
//   - count: 预分配的缓冲区数量
//   - size: 每个缓冲区的大小
//
// 说明:
//   - 在应用启动时调用，预分配指定数量的缓冲区
//   - 减少冷启动时的内存分配延迟
//   - 提升初期性能表现
func (bp *BytePool) Warm(count int, size int) {
	if count <= 0 || size <= 0 {
		return
	}

	// 预分配指定数量的缓冲区
	buffers := make([][]byte, count)
	for i := 0; i < count; i++ {
		buffers[i] = make([]byte, 0, size) // 长度0，容量size
	}

	// 立即归还到对象池进行预热
	for _, buf := range buffers {
		bp.Put(buf)
	}
}

// Drain 清空对象池中的所有缓冲区
//
// 说明:
//   - 清空当前对象池中的所有缓冲区
//   - 重新创建sync.Pool，释放可能占用的大量内存
//   - 适用于内存紧张或需要重置对象池状态的场景
func (bp *BytePool) Drain() {
	// 创建新的sync.Pool替换旧的
	bp.pool = sync.Pool{
		New: func() any {
			buf := make([]byte, 0, bp.defaultSize) // 长度0，容量defaultSize
			return &buf                            // 返回指针避免装箱
		},
	}
}

// WithByte 使用字节切片执行函数，自动管理获取和归还
//
// 参数:
//   - size: 字节切片初始大小
//   - fn: 使用字节切片的函数
//
// 返回值:
//   - []byte: 函数执行后字节切片的数据副本
//
// 说明:
//   - 自动从对象池获取字节切片
//   - 执行用户提供的函数
//   - 获取字节切片数据的副本
//   - 自动归还字节切片到对象池
//   - 即使函数发生panic也会正确归还资源
func (bp *BytePool) WithByte(size int, fn func([]byte)) []byte {
	buffer := bp.Get(size)
	defer bp.Put(buffer)

	fn(buffer)
	// 返回数据的副本，避免在归还后访问
	result := make([]byte, len(buffer))
	copy(result, buffer)
	return result
}

// WithEmptyByte 使用空字节切片执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字节切片初始容量
//   - fn: 使用字节切片的函数，通过append等操作构建数据
//
// 返回值:
//   - []byte: 函数执行后字节切片的数据副本
//
// 说明:
//   - 自动从对象池获取空字节切片（长度为0）
//   - 执行用户提供的函数，函数需要返回构建后的切片
//   - 获取字节切片数据的副本
//   - 自动归还字节切片到对象池
//   - 即使函数发生panic也会正确归还资源
func (bp *BytePool) WithEmptyByte(capacity int, fn func([]byte) []byte) []byte {
	buffer := bp.GetEmpty(capacity)
	defer bp.Put(buffer)

	result := fn(buffer)
	// 返回数据的副本，避免在归还后访问
	finalResult := make([]byte, len(result))
	copy(finalResult, result)
	return finalResult
}
