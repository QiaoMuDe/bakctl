# Package pool

```go
import "gitee.com/MM-Q/go-kit/pool"
```

Package pool 提供高性能对象池管理功能，通过复用对象优化内存使用。

该包实现了基于 sync.Pool 的多种对象池，用于减少频繁的内存分配和回收。 通过复用对象，可以显著提升应用程序的性能，特别是在高并发场景下。

**主要功能：**
- 字节切片对象池管理
- 动态大小对象获取
- 自动内存回收控制
- 防止内存泄漏的大小限制
- 支持多种对象类型的池化

**性能优化：**
- 使用 sync.Pool 减少 GC 压力
- 支持不同大小的对象需求
- 自动限制大对象回收
- 预热机制提升冷启动性能

**使用示例：**

```go
// 获取字节缓冲区
buf := pool.GetByte(64 * 1024)

// 使用缓冲区进行文件操作
_, err := io.CopyBuffer(dst, src, buf)

// 归还缓冲区到对象池
pool.PutByte(buf)
```

Package pool 提供随机数生成器对象池功能，通过对象池优化随机数生成性能。

随机数生成器对象池专门用于复用math/rand.Rand对象， 避免频繁创建随机数生成器的开销，特别适用于ID生成、测试数据生成等场景。

Package pool 提供Timer对象池功能，通过对象池优化定时器使用。

Timer对象池专门用于复用time.Timer对象，避免频繁创建和销毁定时器的开销。
Timer的创建成本相对较高，特别是在高并发场景下，复用可以显著提升性能。

## Constants

```go
const (
	Byte = 1 << (10 * iota) // 1 字节
	KB                      // 千字节 (1024 B)
	MB                      // 兆字节 (1024 KB)
	GB                      // 吉字节 (1024 MB)
	TB                      // 太字节 (1024 GB)
)
```

字节单位定义

## Functions

### func CalculateBufferSize

```go
func CalculateBufferSize(fileSize int64) int
```

CalculateBufferSize 根据文件大小动态计算最佳缓冲区大小。 采用分层策略，平衡内存使用和I/O性能。

**参数:**
- `fileSize`: 文件大小（字节）

**返回:**
- `int`: 计算出的最佳缓冲区大小（字节）

**缓冲区分配策略:**
- ≤ 0 或 ≤ 4KB: 使用 1KB 缓冲区，确保最小缓冲区大小
- 4KB - 32KB: 使用 8KB 缓冲区
- 32KB - 128KB: 使用 32KB 缓冲区
- 128KB - 512KB: 使用 64KB 缓冲区
- 512KB - 1MB: 使用 128KB 缓冲区
- 1MB - 4MB: 使用 256KB 缓冲区
- 4MB - 16MB: 使用 512KB 缓冲区
- 16MB - 64MB: 使用 1MB 缓冲区
- > 64MB: 使用 2MB 缓冲区

**设计原则:**
- 极小文件: 最小化内存占用
- 小文件: 适度缓冲，节省内存
- 大文件: 增大缓冲区，提升I/O吞吐量
- 超大文件: 限制最大缓冲区，避免过度内存消耗

### func DrainBuffer

```go
func DrainBuffer()
```

DrainBuffer 清空默认缓冲区池

### func DrainByte

```go
func DrainByte()
```

DrainByte 清空默认字节池

### func DrainString

```go
func DrainString()
```

DrainString 清空默认字符串池

### func GetBuffer

```go
func GetBuffer(capacity int) *bytes.Buffer
```

GetBuffer 从默认缓冲区池获取字节缓冲区

**参数:**
- `capacity`: 缓冲区初始容量大小

**返回值:**
- `*bytes.Buffer`: 容量至少为capacity的字节缓冲区

### func GetBufferMaxSize

```go
func GetBufferMaxSize() int
```

GetBufferMaxSize 获取默认缓冲区池的当前最大回收大小

**返回值:**
- `int`: 当前最大回收大小

### func GetByte

```go
func GetByte(size int) []byte
```

GetByte 从默认字节池获取指定大小的缓冲区

**参数:**
- `size`: 缓冲区容量大小

**返回值:**
- `[]byte`: 长度为size，容量至少为size的缓冲区

### func GetByteMaxSize

```go
func GetByteMaxSize() int
```

GetByteMaxSize 获取默认字节池的当前最大回收大小

**返回值:**
- `int`: 当前最大回收大小

### func GetEmptyBuffer

```go
func GetEmptyBuffer(minCap int) *bytes.Buffer
```

GetEmptyBuffer 从默认缓冲区池获取空的字节缓冲区

**参数:**
- `minCap`: 最小容量要求

**返回值:**
- `*bytes.Buffer`: 长度为0但容量至少为minCap的字节缓冲区

### func GetEmptyByte

```go
func GetEmptyByte(minCap int) []byte
```

GetEmptyByte 从默认字节池获取空缓冲区

**参数:**
- `minCap`: 最小容量要求

**返回值:**
- `[]byte`: 长度为0但容量至少为minCap的缓冲区切片

### func GetEmptyString

```go
func GetEmptyString(minCap int) *strings.Builder
```

GetEmptyString 从默认字符串池获取空的字符串构建器

**参数:**
- `minCap`: 最小容量要求

**返回值:**
- `*strings.Builder`: 长度为0但容量至少为minCap的字符串构建器

### func GetRand

```go
func GetRand() *rand.Rand
```

GetRand 从池中获取随机数生成器

**返回值:**
- `*rand.Rand`: 随机数生成器实例

**说明:**
- 返回的生成器已经初始化了随机种子
- 使用完毕后应调用PutRand归还
- 注意：返回的生成器不是线程安全的，不要在多个goroutine间共享

### func GetRandWithSeed

```go
func GetRandWithSeed(seed int64) *rand.Rand
```

GetRandWithSeed 获取指定种子的随机数生成器

**参数:**
- `seed`: 随机数种子

**返回值:**
- `*rand.Rand`: 随机数生成器实例

**说明:**
- 返回的生成器使用指定的种子初始化
- 适用于需要可重现随机序列的场景

### func GetString

```go
func GetString(capacity int) *strings.Builder
```

GetString 从默认字符串池获取字符串构建器

**参数:**
- `capacity`: 字符串构建器初始容量大小

**返回值:**
- `*strings.Builder`: 容量至少为capacity的字符串构建器

### func GetStringMaxSize

```go
func GetStringMaxSize() int
```

GetStringMaxSize 获取默认字符串池的当前最大回收大小

**返回值:**
- `int`: 当前最大回收大小

### func GetTimer

```go
func GetTimer(duration time.Duration) *time.Timer
```

GetTimer 从池中获取定时器并设置超时时间

**参数:**
- `duration`: 定时器超时时间

**返回值:**
- `*time.Timer`: 已设置超时时间的定时器

**说明:**
- 返回的定时器已经启动，会在指定时间后触发
- 适用于超时控制场景，定时器会在指定时间后自动触发
- 使用完毕后应调用PutTimer归还

### func GetTimerEmpty

```go
func GetTimerEmpty() *time.Timer
```

GetTimerEmpty 从池中获取未启动的定时器

**返回值:**
- `*time.Timer`: 未启动的定时器，需要手动调用Reset设置时间

**说明:**
- 适用于需要手动控制定时器启动和停止的场景
- 定时器处于停止状态，不会自动触发
- 获取后需要调用timer.Reset(duration)启动
- 使用完毕后应调用PutTimer归还

### func PutBuffer

```go
func PutBuffer(buffer *bytes.Buffer)
```

PutBuffer 将字节缓冲区归还到默认缓冲区池

**参数:**
- `buffer`: 要归还的字节缓冲区

**说明:**
- 该函数将字节缓冲区归还到对象池，以便后续复用。
- 只有容量不超过64KB的缓冲区才会被归还，以避免对象池占用过多内存。

### func PutByte

```go
func PutByte(buffer []byte)
```

PutByte 将缓冲区归还到默认字节池

**参数:**
- `buffer`: 要归还的缓冲区

**说明:**
- 该函数将缓冲区归还到对象池，以便后续复用。
- 只有容量不超过1MB的缓冲区才会被归还，以避免对象池占用过多内存。

### func PutRand

```go
func PutRand(rng *rand.Rand)
```

PutRand 将随机数生成器归还到池中

**参数:**
- `rng`: 要归还的随机数生成器

### func PutString

```go
func PutString(builder *strings.Builder)
```

PutString 将字符串构建器归还到默认字符串池

**参数:**
- `builder`: 要归还的字符串构建器

**说明:**
- 该函数将字符串构建器归还到对象池，以便后续复用。
- 只有容量不超过64KB的构建器才会被归还，以避免对象池占用过多内存。

### func PutTimer

```go
func PutTimer(timer *time.Timer)
```

PutTimer 将定时器归还到池中

**参数:**
- `timer`: 要归还的定时器

**说明:**
- 该函数会自动停止定时器并清理状态
- 归还后的定时器会被重置，可以安全复用

### func SetBufferMaxSize

```go
func SetBufferMaxSize(maxSize int)
```

SetBufferMaxSize 动态调整默认缓冲区池的最大回收大小

**参数:**
- `maxSize`: 新的最大回收大小

### func SetByteMaxSize

```go
func SetByteMaxSize(maxSize int)
```

SetByteMaxSize 动态调整默认字节池的最大回收大小

**参数:**
- `maxSize`: 新的最大回收大小

### func SetStringMaxSize

```go
func SetStringMaxSize(maxSize int)
```

SetStringMaxSize 动态调整默认字符串池的最大回收大小

**参数:**
- `maxSize`: 新的最大回收大小

### func WarmBuffer

```go
func WarmBuffer(count int, capacity int)
```

WarmBuffer 预热默认缓冲区池

**参数:**
- `count`: 预分配的字节缓冲区数量
- `capacity`: 每个字节缓冲区的容量

### func WarmByte

```go
func WarmByte(count int, size int)
```

WarmByte 预热默认字节池

**参数:**
- `count`: 预分配的缓冲区数量
- `size`: 每个缓冲区的大小

### func WarmString

```go
func WarmString(count int, capacity int)
```

WarmString 预热默认字符串池

**参数:**
- `count`: 预分配的字符串构建器数量
- `capacity`: 每个字符串构建器的容量

### func WithBuffer

```go
func WithBuffer(capacity int, fn func(*bytes.Buffer)) []byte
```

WithBuffer 使用字节缓冲区执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字节缓冲区初始容量大小
- `fn`: 使用字节缓冲区的函数

**返回值:**
- `[]byte`: 函数执行后缓冲区的字节数据副本

**使用示例:**

```go
data := pool.WithBuffer(1024, func(buf *bytes.Buffer) {
    buf.WriteString("Hello")
    buf.WriteByte(' ')
    buf.WriteString("World")
})
```

### func WithByte

```go
func WithByte(size int, fn func([]byte)) []byte
```

WithByte 使用字节切片执行函数，自动管理获取和归还

**参数:**
- `size`: 字节切片初始大小
- `fn`: 使用字节切片的函数

**返回值:**
- `[]byte`: 函数执行后字节切片的数据副本

**使用示例:**

```go
data := pool.WithByte(1024, func(buf []byte) {
    copy(buf, []byte("Hello World"))
    // 可以直接操作buf进行读写
})
```

### func WithEmptyByte

```go
func WithEmptyByte(capacity int, fn func([]byte) []byte) []byte
```

WithEmptyByte 使用空字节切片执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字节切片初始容量
- `fn`: 使用字节切片的函数，通过append等操作构建数据

**返回值:**
- `[]byte`: 函数执行后字节切片的数据副本

**使用示例:**

```go
data := pool.WithEmptyByte(1024, func(buf []byte) []byte {
    buf = append(buf, []byte("Hello")...)
    buf = append(buf, ' ')
    buf = append(buf, []byte("World")...)
    return buf
})
```

### func WithRand

```go
func WithRand[T any](fn func(*rand.Rand) T) T
```

WithRand 使用随机数生成器执行函数，自动管理获取和归还

**参数:**
- `fn`: 使用随机数生成器的函数

**返回值:**
- `T`: 函数返回的结果

**使用示例:**

```go
// 生成随机整数
num := pool.WithRand(func(rng *rand.Rand) int {
    return rng.Intn(100)
})

// 生成随机字符串
str := pool.WithRand(func(rng *rand.Rand) string {
    return fmt.Sprintf("id_%d", rng.Int63())
})
```

### func WithRandSeed

```go
func WithRandSeed[T any](seed int64, fn func(*rand.Rand) T) T
```

WithRandSeed 使用指定种子的随机数生成器执行函数，自动管理获取和归还

**参数:**
- `seed`: 随机数种子
- `fn`: 使用随机数生成器的函数

**返回值:**
- `T`: 函数返回的结果

**使用示例:**

```go
// 生成可重现的随机序列
nums := pool.WithRandSeed(12345, func(rng *rand.Rand) []int {
    result := make([]int, 5)
    for i := range result {
        result[i] = rng.Intn(100)
    }
    return result
})
```

### func WithString

```go
func WithString(capacity int, fn func(*strings.Builder)) string
```

WithString 使用字符串构建器执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字符串构建器初始容量大小
- `fn`: 使用字符串构建器的函数

**返回值:**
- `string`: 函数执行后构建的字符串结果

**使用示例:**

```go
result := pool.WithString(64, func(buf *strings.Builder) {
    buf.WriteString("Hello")
    buf.WriteByte(' ')
    buf.WriteString("World")
})
```

## Types

### type BufferPool

```go
type BufferPool struct {
	// Has unexported fields.
}
```

BufferPool 字节缓冲区对象池，支持自定义配置

#### func NewBufferPool

```go
func NewBufferPool(defaultSize, maxSize int) *BufferPool
```

NewBufferPool 创建新的字节缓冲区对象池

**参数:**
- `defaultSize`: 默认字节缓冲区容量大小
- `maxSize`: 最大回收缓冲区大小，超过此大小的缓冲区不会被回收

**返回值:**
- `*BufferPool`: 字节缓冲区对象池实例

#### func (*BufferPool) Drain

```go
func (bp *BufferPool) Drain()
```

Drain 清空对象池中的所有字节缓冲区

**说明:**
- 清空当前对象池中的所有字节缓冲区
- 重新创建sync.Pool，释放可能占用的大量内存
- 适用于内存紧张或需要重置对象池状态的场景

#### func (*BufferPool) Get

```go
func (bp *BufferPool) Get(capacity int) *bytes.Buffer
```

Get 获取指定容量的字节缓冲区

**参数:**
- `capacity`: 需要的字节缓冲区容量大小

**返回:**
- `*bytes.Buffer`: 容量至少为capacity的字节缓冲区

**说明:**
- 返回的字节缓冲区已经重置为空状态，可以直接使用
- 底层容量可能大于capacity，来自对象池的复用缓冲区

#### func (*BufferPool) GetEmpty

```go
func (bp *BufferPool) GetEmpty(minCap int) *bytes.Buffer
```

GetEmpty 获取指定容量的空字节缓冲区

**参数:**
- `minCap`: 最小容量要求

**返回:**
- `*bytes.Buffer`: 长度为0但容量至少为minCap的字节缓冲区

**说明:**
- 适用于需要逐步写入数据的场景
- 避免频繁的内存重新分配

#### func (*BufferPool) GetMaxSize

```go
func (bp *BufferPool) GetMaxSize() int
```

GetMaxSize 获取当前最大回收缓冲区大小

**返回:**
- `int`: 当前最大回收大小

#### func (*BufferPool) Put

```go
func (bp *BufferPool) Put(buffer *bytes.Buffer)
```

Put 归还字节缓冲区到对象池

**参数:**
- `buffer`: 要归还的字节缓冲区

#### func (*BufferPool) SetMaxSize

```go
func (bp *BufferPool) SetMaxSize(maxSize int)
```

SetMaxSize 动态调整最大回收缓冲区大小

**参数:**
- `maxSize`: 新的最大回收大小

**说明:**
- 运行时动态调整配置
- 如果新的maxSize小于当前值，建议调用Drain()清空对象池

#### func (*BufferPool) Warm

```go
func (bp *BufferPool) Warm(count int, capacity int)
```

Warm 预热对象池

**参数:**
- `count`: 预分配的字节缓冲区数量
- `capacity`: 每个字节缓冲区的容量

**说明:**
- 在应用启动时调用，预分配指定数量的字节缓冲区
- 减少冷启动时的内存分配延迟
- 提升初期性能表现

#### func (*BufferPool) WithBuffer

```go
func (bp *BufferPool) WithBuffer(capacity int, fn func(*bytes.Buffer)) []byte
```

WithBuffer 使用字节缓冲区执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字节缓冲区初始容量大小
- `fn`: 使用字节缓冲区的函数

**返回值:**
- `[]byte`: 函数执行后缓冲区的字节数据副本

**说明:**
- 自动从对象池获取字节缓冲区
- 执行用户提供的函数
- 获取缓冲区字节数据的副本
- 自动归还字节缓冲区到对象池
- 即使函数发生panic也会正确归还资源

### type BytePool

```go
type BytePool struct {
	// Has unexported fields.
}
```

BytePool 字节切片对象池，支持自定义配置

#### func NewBytePool

```go
func NewBytePool(defaultSize, maxSize int) *BytePool
```

NewBytePool 创建新的字节切片对象池

**参数:**
- `defaultSize`: 默认缓冲区大小
- `maxSize`: 最大回收缓冲区大小，超过此大小的缓冲区不会被回收

**返回值:**
- `*BytePool`: 字节切片对象池实例

#### func (*BytePool) Drain

```go
func (bp *BytePool) Drain()
```

Drain 清空对象池中的所有缓冲区

**说明:**
- 清空当前对象池中的所有缓冲区
- 重新创建sync.Pool，释放可能占用的大量内存
- 适用于内存紧张或需要重置对象池状态的场景

#### func (*BytePool) Get

```go
func (bp *BytePool) Get(size int) []byte
```

Get 获取指定容量的缓冲区

**参数:**
- `size`: 需要的缓冲区容量大小

**返回:**
- `[]byte`: 长度为size，容量至少为size的缓冲区切片

**说明:**
- 返回的缓冲区长度等于请求的size，可以直接使用
- 底层容量可能大于size，来自对象池的复用缓冲区

#### func (*BytePool) GetEmpty

```go
func (bp *BytePool) GetEmpty(minCap int) []byte
```

GetEmpty 获取指定容量的空缓冲区

**参数:**
- `minCap`: 最小容量要求

**返回:**
- `[]byte`: 长度为0但容量至少为minCap的缓冲区切片

**说明:**
- 适用于需要使用append操作逐步构建数据的场景
- 避免频繁的内存重新分配

#### func (*BytePool) GetMaxSize

```go
func (bp *BytePool) GetMaxSize() int
```

GetMaxSize 获取当前最大回收缓冲区大小

**返回:**
- `int`: 当前最大回收大小

#### func (*BytePool) Put

```go
func (bp *BytePool) Put(buffer []byte)
```

Put 归还缓冲区到对象池

**参数:**
- `buffer`: 要归还的缓冲区

#### func (*BytePool) SetMaxSize

```go
func (bp *BytePool) SetMaxSize(maxSize int)
```

SetMaxSize 动态调整最大回收缓冲区大小

**参数:**
- `maxSize`: 新的最大回收大小

**说明:**
- 运行时动态调整配置
- 如果新的maxSize小于当前值，建议调用Drain()清空对象池

#### func (*BytePool) Warm

```go
func (bp *BytePool) Warm(count int, size int)
```

Warm 预热对象池

**参数:**
- `count`: 预分配的缓冲区数量
- `size`: 每个缓冲区的大小

**说明:**
- 在应用启动时调用，预分配指定数量的缓冲区
- 减少冷启动时的内存分配延迟
- 提升初期性能表现

#### func (*BytePool) WithByte

```go
func (bp *BytePool) WithByte(size int, fn func([]byte)) []byte
```

WithByte 使用字节切片执行函数，自动管理获取和归还

**参数:**
- `size`: 字节切片初始大小
- `fn`: 使用字节切片的函数

**返回值:**
- `[]byte`: 函数执行后字节切片的数据副本

**说明:**
- 自动从对象池获取字节切片
- 执行用户提供的函数
- 获取字节切片数据的副本
- 自动归还字节切片到对象池
- 即使函数发生panic也会正确归还资源

#### func (*BytePool) WithEmptyByte

```go
func (bp *BytePool) WithEmptyByte(capacity int, fn func([]byte) []byte) []byte
```

WithEmptyByte 使用空字节切片执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字节切片初始容量
- `fn`: 使用字节切片的函数，通过append等操作构建数据

**返回值:**
- `[]byte`: 函数执行后字节切片的数据副本

**说明:**
- 自动从对象池获取空字节切片（长度为0）
- 执行用户提供的函数，函数需要返回构建后的切片
- 获取字节切片数据的副本
- 自动归还字节切片到对象池
- 即使函数发生panic也会正确归还资源

### type StringPool

```go
type StringPool struct {
	// Has unexported fields.
}
```

StringPool 字符串构建器对象池，支持自定义配置

#### func NewStringPool

```go
func NewStringPool(defaultSize, maxSize int) *StringPool
```

NewStringPool 创建新的字符串构建器对象池

**参数:**
- `defaultSize`: 默认字符串构建器容量大小
- `maxSize`: 最大回收构建器大小，超过此大小的构建器不会被回收

**返回值:**
- `*StringPool`: 字符串构建器对象池实例

#### func (*StringPool) Drain

```go
func (sp *StringPool) Drain()
```

Drain 清空对象池中的所有字符串构建器

**说明:**
- 清空当前对象池中的所有字符串构建器
- 重新创建sync.Pool，释放可能占用的大量内存
- 适用于内存紧张或需要重置对象池状态的场景

#### func (*StringPool) Get

```go
func (sp *StringPool) Get(capacity int) *strings.Builder
```

Get 获取指定容量的字符串构建器

**参数:**
- `capacity`: 需要的字符串构建器容量大小

**返回:**
- `*strings.Builder`: 容量至少为capacity的字符串构建器

**说明:**
- 返回的字符串构建器已经重置为空状态，可以直接使用
- 底层容量可能大于capacity，来自对象池的复用构建器

#### func (*StringPool) GetEmpty

```go
func (sp *StringPool) GetEmpty(minCap int) *strings.Builder
```

GetEmpty 获取指定容量的空字符串构建器

**参数:**
- `minCap`: 最小容量要求

**返回:**
- `*strings.Builder`: 长度为0但容量至少为minCap的字符串构建器

**说明:**
- 适用于需要逐步构建字符串的场景
- 避免频繁的内存重新分配

#### func (*StringPool) GetMaxSize

```go
func (sp *StringPool) GetMaxSize() int
```

GetMaxSize 获取当前最大回收构建器大小

**返回:**
- `int`: 当前最大回收大小

#### func (*StringPool) Put

```go
func (sp *StringPool) Put(builder *strings.Builder)
```

Put 归还字符串构建器到对象池

**参数:**
- `builder`: 要归还的字符串构建器

#### func (*StringPool) SetMaxSize

```go
func (sp *StringPool) SetMaxSize(maxSize int)
```

SetMaxSize 动态调整最大回收构建器大小

**参数:**
- `maxSize`: 新的最大回收大小

**说明:**
- 运行时动态调整配置
- 如果新的maxSize小于当前值，建议调用Drain()清空对象池

#### func (*StringPool) Warm

```go
func (sp *StringPool) Warm(count int, capacity int)
```

Warm 预热对象池

**参数:**
- `count`: 预分配的字符串构建器数量
- `capacity`: 每个字符串构建器的容量

**说明:**
- 在应用启动时调用，预分配指定数量的字符串构建器
- 减少冷启动时的内存分配延迟
- 提升初期性能表现

#### func (*StringPool) WithString

```go
func (sp *StringPool) WithString(capacity int, fn func(*strings.Builder)) string
```

WithString 使用字符串构建器执行函数，自动管理获取和归还

**参数:**
- `capacity`: 字符串构建器初始容量大小
- `fn`: 使用字符串构建器的函数

**返回值:**
- `string`: 函数执行后构建的字符串结果

**说明:**
- 自动从对象池获取字符串构建器
- 执行用户提供的函数
- 获取构建的字符串结果
- 自动归还字符串构建器到对象池
- 即使函数发生panic也会正确归还资源

