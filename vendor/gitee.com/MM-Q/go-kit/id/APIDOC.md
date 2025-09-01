# Package id

```go
import "gitee.com/MM-Q/go-kit/id"
```

## Functions

### func GenID

```go
func GenID(n int) string
```

GenID 生成ID 用于生成带有时间戳和随机数的ID，格式为：时间戳(16位) + 随机数(n位) 默认使用16位时间戳，提供更好的唯一性保证

**参数:**
- `n`: 随机部分长度

**返回:**
- 生成的ID

### func GenIDWithLen

```go
func GenIDWithLen(tsLen, randLen int) string
```

GenIDWithLen 生成指定长度的ID 用于生成带有自定义时间戳和随机数长度的ID

**参数:**
- `tsLen`: 时间戳长度，-1表示使用完整时间戳，正数时自动限制在16位以内
- `randLen`: 随机部分长度

**返回:**
- 生成的ID

### func GenIDs

```go
func GenIDs(count, n int) []string
```

GenIDs 批量生成ID 用于批量生成多个唯一ID，使用16位时间戳提供更好的唯一性

**参数:**
- `count`: 要生成的ID数量
- `n`: 每个ID随机部分的长度

**返回:**
- ID切片，参数无效时返回nil

### func GenMaskedID

```go
func GenMaskedID() string
```

GenMaskedID 生成带隐藏时间戳的ID 格式: 6位随机字符串 + 微秒时间戳后8位 + 6位随机字符串 总长度: 20位, 时间戳被随机字符包围, 提供更好的隐蔽性

**返回:**
- 20位长度的带隐藏时间戳的ID

### func GenWithPrefix

```go
func GenWithPrefix(prefix string, n int) string
```

GenWithPrefix 生成带前缀的ID 用于生成带有自定义前缀的ID，格式为：prefix_ID

**参数:**
- `prefix`: ID前缀字符串
- `n`: 随机部分长度

**返回:**
- 带前缀的ID字符串

### func MicroTime

```go
func MicroTime() string
```

MicroTime 用于生成基于当前微秒时间戳的ID

**返回:**
- 微秒时间戳字符串

### func NanoTime

```go
func NanoTime() string
```

NanoTime 用于生成基于当前纳秒时间戳的ID

**返回:**
- 纳秒时间戳字符串

### func RandomString

```go
func RandomString(length int) string
```

RandomString 生成指定长度的随机字符串 用于生成仅包含随机字符的字符串，不包含时间戳等其他信息

**参数:**
- `length`: 随机字符串长度

**返回:**
- 生成的随机字符串, 当长度小于0时返回空字符串

### func UUID

```go
func UUID() string
```

UUID 生成类UUID格式 用于生成类似UUID的字符串，格式为：8-4-4-4-12 使用crypto/rand提供强随机性，确保并发安全和高唯一性

**返回:**
- 36位长度的UUID格式字符串

