package id

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"gitee.com/MM-Q/go-kit/pool"
)

// 随机字符集(固定62位请勿修改)
const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// generateRandomString 生成指定长度的随机字符串
// 使用提供的随机数生成器生成随机字符串
//
// 参数:
//   - r: 随机数生成器
//   - length: 字符串长度
//   - buf: 字符串构建器
func generateRandomString(r interface{ Intn(int) int }, length int, buf *strings.Builder) {
	for i := 0; i < length; i++ {
		buf.WriteByte(chars[r.Intn(62)])
	}
}

// generateTruncatedTimestamp 生成指定长度的截断时间戳
//
// 参数:
//   - tsLen: 时间戳长度，必须大于0，最大16位(微秒时间戳长度)
//
// 返回:
//   - 截断后的时间戳字符串
func generateTruncatedTimestamp(tsLen int) string {
	// 快速失败：参数验证
	if tsLen <= 0 {
		return ""
	}

	// 限制最大长度为16位(微秒时间戳的实际长度)
	if tsLen > 16 {
		tsLen = 16
	}

	// 计算模数：10^tsLen
	tsMod := int64(1)
	for i := 0; i < tsLen; i++ {
		tsMod *= 10
	}

	// 生成格式化字符串并应用
	tsFormat := fmt.Sprintf("%%0%dd", tsLen)
	return fmt.Sprintf(tsFormat, time.Now().UnixMicro()%tsMod)
}

// genIDInternal 内部ID生成方法
// 用于生成带有时间戳和随机数的ID，支持可配置的时间戳长度
//
// 参数:
//   - tsLen: 时间戳长度，-1表示使用完整时间戳，正数时最大限制为16位
//   - randLen: 随机部分长度
//
// 返回:
//   - 生成的ID
func genIDInternal(tsLen, randLen int) string {
	// 快速失败：参数验证
	if randLen < 0 {
		return ""
	}

	// 生成时间戳部分
	var ts string
	switch {
	case tsLen == -1:
		// 使用完整微秒时间戳
		ts = fmt.Sprintf("%d", time.Now().UnixMicro())
	case tsLen <= 0:
		// 时间戳长度无效，不生成时间戳部分
		ts = ""
	default:
		// 使用指定长度的时间戳
		ts = generateTruncatedTimestamp(tsLen)
	}

	// 快速返回：如果只需要时间戳
	if randLen == 0 {
		return ts
	}

	// 计算总长度并生成随机部分
	totalLen := len(ts) + randLen
	r := pool.GetRand()
	defer pool.PutRand(r)

	return pool.WithStrCap(totalLen, func(buf *strings.Builder) {
		buf.WriteString(ts)
		// 生成随机数部分
		generateRandomString(r, randLen, buf)
	})

}

// GenID 生成ID
// 用于生成带有时间戳和随机数的ID，格式为：时间戳(16位) + 随机数(n位)
// 默认使用16位时间戳，提供更好的唯一性保证
//
// 参数:
//   - n: 随机部分长度
//
// 返回:
//   - 生成的ID
func GenID(n int) string {
	return genIDInternal(16, n)
}

// GenIDWithLen 生成指定长度的ID
// 用于生成带有自定义时间戳和随机数长度的ID
//
// 参数:
//   - tsLen: 时间戳长度，-1表示使用完整时间戳，正数时自动限制在16位以内
//   - randLen: 随机部分长度
//
// 返回:
//   - 生成的ID
func GenIDWithLen(tsLen, randLen int) string {
	return genIDInternal(tsLen, randLen)
}

// GenIDs 批量生成ID
// 用于批量生成多个唯一ID，使用16位时间戳提供更好的唯一性
//
// 参数:
//   - count: 要生成的ID数量
//   - n: 每个ID随机部分的长度
//
// 返回:
//   - ID切片，参数无效时返回nil
func GenIDs(count, n int) []string {
	if count <= 0 || n < 0 {
		return nil
	}

	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = genIDInternal(16, n)
	}

	return ids
}

// GenWithPrefix 生成带前缀的ID
// 用于生成带有自定义前缀的ID，格式为：prefix_ID
//
// 参数:
//   - prefix: ID前缀字符串
//   - n: 随机部分长度
//
// 返回:
//   - 带前缀的ID字符串
func GenWithPrefix(prefix string, n int) string {
	if n < 0 {
		return prefix
	}

	id := GenID(n)
	if prefix == "" {
		return id
	}

	return pool.WithStrCap(len(prefix)+len(id)+1, func(buf *strings.Builder) {
		buf.WriteString(prefix)
		buf.WriteByte('_')
		buf.WriteString(id)
	})
}

// UUID 生成类UUID格式
// 用于生成类似UUID的字符串，格式为：8-4-4-4-12
// 使用crypto/rand提供强随机性，确保并发安全和高唯一性
//
// 返回:
//   - 36位长度的UUID格式字符串
func UUID() string {
	// 从字节池获取32字节的缓冲区用于加密安全随机数据
	randomBytes := pool.GetByteCap(32)
	defer pool.PutByte(randomBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		// 极少情况下crypto/rand失败时的fallback
		r := pool.GetRand()
		defer pool.PutRand(r)

		return pool.WithStrCap(36, func(buf *strings.Builder) {
			// 8位
			generateRandomString(r, 8, buf)
			buf.WriteByte('-')

			// 4位
			generateRandomString(r, 4, buf)
			buf.WriteByte('-')

			// 4位
			generateRandomString(r, 4, buf)
			buf.WriteByte('-')

			// 4位
			generateRandomString(r, 4, buf)
			buf.WriteByte('-')

			// 12位
			generateRandomString(r, 12, buf)
		})
	}

	// 使用crypto/rand生成的随机字节映射到字符集
	return pool.WithStrCap(36, func(buf *strings.Builder) {
		byteIndex := 0

		// 8位
		for i := 0; i < 8; i++ {
			buf.WriteByte(chars[randomBytes[byteIndex]%62])
			byteIndex++
		}
		buf.WriteByte('-')

		// 4位
		for i := 0; i < 4; i++ {
			buf.WriteByte(chars[randomBytes[byteIndex]%62])
			byteIndex++
		}
		buf.WriteByte('-')

		// 4位
		for i := 0; i < 4; i++ {
			buf.WriteByte(chars[randomBytes[byteIndex]%62])
			byteIndex++
		}
		buf.WriteByte('-')

		// 4位
		for i := 0; i < 4; i++ {
			buf.WriteByte(chars[randomBytes[byteIndex]%62])
			byteIndex++
		}
		buf.WriteByte('-')

		// 12位
		for i := 0; i < 12; i++ {
			buf.WriteByte(chars[randomBytes[byteIndex]%62])
			byteIndex++
		}
	})
}

// GenMaskedID 生成带隐藏时间戳的ID
// 格式: 6位随机字符串 + 微秒时间戳后8位 + 6位随机字符串
// 总长度: 20位, 时间戳被随机字符包围, 提供更好的隐蔽性
//
// 返回:
//   - 20位长度的带隐藏时间戳的ID
func GenMaskedID() string {
	// 获取微秒时间戳的后8位
	microTimestamp := time.Now().UnixMicro()
	last8Digits := microTimestamp % 100000000 // 取后8位
	timestampPart := fmt.Sprintf("%08d", last8Digits)

	// 从随机数池获取随机数生成器
	r := pool.GetRand()
	defer pool.PutRand(r)

	return pool.WithStrCap(20, func(buf *strings.Builder) {
		// 前6位随机字符
		generateRandomString(r, 6, buf)

		// 中间8位时间戳
		buf.WriteString(timestampPart)

		// 后6位随机字符
		generateRandomString(r, 6, buf)
	})
}

// RandomString 生成指定长度的随机字符串
// 用于生成仅包含随机字符的字符串，不包含时间戳等其他信息
//
// 参数:
//   - length: 随机字符串长度
//
// 返回:
//   - 生成的随机字符串, 当长度小于0时返回空字符串
func RandomString(length int) string {
	// 快速失败：参数验证
	if length <= 0 {
		return ""
	}

	// 获取随机数生成器并生成随机字符串
	r := pool.GetRand()
	defer pool.PutRand(r)

	return pool.WithStrCap(length, func(buf *strings.Builder) {
		generateRandomString(r, length, buf)
	})
}

// MicroTime 用于生成基于当前微秒时间戳的ID
//
// 返回:
//   - 微秒时间戳字符串
func MicroTime() string {
	return fmt.Sprintf("%d", time.Now().UnixMicro())
}

// NanoTime 用于生成基于当前纳秒时间戳的ID
//
// 返回:
//   - 纳秒时间戳字符串
func NanoTime() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
