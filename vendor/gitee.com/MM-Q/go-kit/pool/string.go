package pool

import (
	"strings"
	"sync"
)

// 全局默认字符串构建器池实例
var defaultStringPool = NewStringPool(1024, 64*1024)

// GetString 从默认字符串池获取字符串构建器
//
// 参数:
//   - capacity: 字符串构建器初始容量大小
//
// 返回值:
//   - *strings.Builder: 容量至少为capacity的字符串构建器
func GetString(capacity int) *strings.Builder {
	return defaultStringPool.Get(capacity)
}

// PutString 将字符串构建器归还到默认字符串池
//
// 参数:
//   - builder: 要归还的字符串构建器
//
// 说明:
//   - 该函数将字符串构建器归还到对象池，以便后续复用。
//   - 只有容量不超过64KB的构建器才会被归还，以避免对象池占用过多内存。
func PutString(builder *strings.Builder) {
	defaultStringPool.Put(builder)
}

// GetEmptyString 从默认字符串池获取空的字符串构建器
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回值:
//   - *strings.Builder: 长度为0但容量至少为minCap的字符串构建器
func GetEmptyString(minCap int) *strings.Builder {
	return defaultStringPool.GetEmpty(minCap)
}

// SetStringMaxSize 动态调整默认字符串池的最大回收大小
//
// 参数:
//   - maxSize: 新的最大回收大小
func SetStringMaxSize(maxSize int) {
	defaultStringPool.SetMaxSize(maxSize)
}

// GetStringMaxSize 获取默认字符串池的当前最大回收大小
//
// 返回值:
//   - int: 当前最大回收大小
func GetStringMaxSize() int {
	return defaultStringPool.GetMaxSize()
}

// WarmString 预热默认字符串池
//
// 参数:
//   - count: 预分配的字符串构建器数量
//   - capacity: 每个字符串构建器的容量
func WarmString(count int, capacity int) {
	defaultStringPool.Warm(count, capacity)
}

// DrainString 清空默认字符串池
func DrainString() {
	defaultStringPool.Drain()
}

// WithString 使用字符串构建器执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字符串构建器初始容量大小
//   - fn: 使用字符串构建器的函数
//
// 返回值:
//   - string: 函数执行后构建的字符串结果
//
// 使用示例:
//
//	result := pool.WithString(64, func(buf *strings.Builder) {
//	    buf.WriteString("Hello")
//	    buf.WriteByte(' ')
//	    buf.WriteString("World")
//	})
func WithString(capacity int, fn func(*strings.Builder)) string {
	return defaultStringPool.WithString(capacity, fn)
}

// StringPool 字符串构建器对象池，支持自定义配置
type StringPool struct {
	pool        sync.Pool // 字符串构建器对象池
	maxSize     int       // 最大回收构建器大小
	defaultSize int       // 默认构建器大小
}

// NewStringPool 创建新的字符串构建器对象池
//
// 参数:
//   - defaultSize: 默认字符串构建器容量大小
//   - maxSize: 最大回收构建器大小，超过此大小的构建器不会被回收
//
// 返回值:
//   - *StringPool: 字符串构建器对象池实例
func NewStringPool(defaultSize, maxSize int) *StringPool {
	if defaultSize <= 0 {
		defaultSize = 1024 // 默认1KB
	}
	if maxSize <= 0 {
		maxSize = 64 * 1024 // 默认64KB
	}

	return &StringPool{
		maxSize:     maxSize,
		defaultSize: defaultSize,
		pool: sync.Pool{
			New: func() any {
				builder := &strings.Builder{}
				builder.Grow(defaultSize) // 预分配容量
				return builder
			},
		},
	}
}

// Get 获取指定容量的字符串构建器
//
// 参数:
//   - capacity: 需要的字符串构建器容量大小
//
// 返回:
//   - *strings.Builder: 容量至少为capacity的字符串构建器
//
// 说明:
//   - 返回的字符串构建器已经重置为空状态，可以直接使用
//   - 底层容量可能大于capacity，来自对象池的复用构建器
func (sp *StringPool) Get(capacity int) *strings.Builder {
	builder, ok := sp.pool.Get().(*strings.Builder)
	if !ok {
		// 类型断言失败，创建新的
		builder = &strings.Builder{}
		builder.Grow(capacity)
		return builder
	}

	// 重置构建器状态
	builder.Reset()

	// 如果当前容量不足，扩容到所需大小
	if builder.Cap() < capacity {
		builder.Grow(capacity - builder.Cap())
	}

	return builder
}

// Put 归还字符串构建器到对象池
//
// 参数:
//   - builder: 要归还的字符串构建器
func (sp *StringPool) Put(builder *strings.Builder) {
	if builder == nil || builder.Cap() > sp.maxSize {
		return // 不回收nil或超过最大大小的构建器
	}

	// 重置构建器状态，清空内容
	builder.Reset()

	sp.pool.Put(builder)
}

// GetEmpty 获取指定容量的空字符串构建器
//
// 参数:
//   - minCap: 最小容量要求
//
// 返回:
//   - *strings.Builder: 长度为0但容量至少为minCap的字符串构建器
//
// 说明:
//   - 适用于需要逐步构建字符串的场景
//   - 避免频繁的内存重新分配
func (sp *StringPool) GetEmpty(minCap int) *strings.Builder {
	builder, ok := sp.pool.Get().(*strings.Builder)
	if !ok {
		// 类型断言失败，创建新的
		builder = &strings.Builder{}
		builder.Grow(minCap)
		return builder
	}

	// 重置构建器状态
	builder.Reset()

	// 如果当前容量不足，扩容到所需大小
	if builder.Cap() < minCap {
		builder.Grow(minCap - builder.Cap())
	}

	return builder
}

// SetMaxSize 动态调整最大回收构建器大小
//
// 参数:
//   - maxSize: 新的最大回收大小
//
// 说明:
//   - 运行时动态调整配置
//   - 如果新的maxSize小于当前值，建议调用Drain()清空对象池
func (sp *StringPool) SetMaxSize(maxSize int) {
	if maxSize <= 0 {
		maxSize = 64 * 1024 // 默认64KB
	}
	sp.maxSize = maxSize
}

// GetMaxSize 获取当前最大回收构建器大小
//
// 返回:
//   - int: 当前最大回收大小
func (sp *StringPool) GetMaxSize() int {
	return sp.maxSize
}

// Warm 预热对象池
//
// 参数:
//   - count: 预分配的字符串构建器数量
//   - capacity: 每个字符串构建器的容量
//
// 说明:
//   - 在应用启动时调用，预分配指定数量的字符串构建器
//   - 减少冷启动时的内存分配延迟
//   - 提升初期性能表现
func (sp *StringPool) Warm(count int, capacity int) {
	if count <= 0 || capacity <= 0 {
		return
	}

	// 预分配指定数量的字符串构建器
	builders := make([]*strings.Builder, count)
	for i := 0; i < count; i++ {
		builder := &strings.Builder{}
		builder.Grow(capacity)
		builders[i] = builder
	}

	// 立即归还到对象池进行预热
	for _, builder := range builders {
		sp.Put(builder)
	}
}

// Drain 清空对象池中的所有字符串构建器
//
// 说明:
//   - 清空当前对象池中的所有字符串构建器
//   - 重新创建sync.Pool，释放可能占用的大量内存
//   - 适用于内存紧张或需要重置对象池状态的场景
func (sp *StringPool) Drain() {
	// 创建新的sync.Pool替换旧的
	sp.pool = sync.Pool{
		New: func() any {
			builder := &strings.Builder{}
			builder.Grow(sp.defaultSize) // 预分配容量
			return builder
		},
	}
}

// WithString 使用字符串构建器执行函数，自动管理获取和归还
//
// 参数:
//   - capacity: 字符串构建器初始容量大小
//   - fn: 使用字符串构建器的函数
//
// 返回值:
//   - string: 函数执行后构建的字符串结果
//
// 说明:
//   - 自动从对象池获取字符串构建器
//   - 执行用户提供的函数
//   - 获取构建的字符串结果
//   - 自动归还字符串构建器到对象池
//   - 即使函数发生panic也会正确归还资源
func (sp *StringPool) WithString(capacity int, fn func(*strings.Builder)) string {
	builder := sp.Get(capacity)
	defer sp.Put(builder)

	fn(builder)
	return builder.String()
}
