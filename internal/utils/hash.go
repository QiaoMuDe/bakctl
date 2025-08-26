package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

// 字节单位定义
const (
	Byte = 1 << (10 * iota) // 1 字节
	KB                      // 千字节 (1024 B)
	MB                      // 兆字节 (1024 KB)
	GB                      // 吉字节 (1024 MB)
	TB                      // 太字节 (1024 GB)
)

// 支持的哈希算法列表
var supportedAlgorithms = map[string]func() hash.Hash{
	"md5":    md5.New,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
}

// Checksum 计算文件哈希值
//
// 参数:
//   - filePath: 文件路径
//   - hashFunc: 哈希函数构造器
//
// 返回:
//   - string: 文件的十六进制哈希值
//   - error: 错误信息，如果计算失败
//
// 注意:
//   - 根据文件大小动态分配缓冲区以提高性能
//   - 支持任何实现hash.Hash接口的哈希算法
//   - 使用io.CopyBuffer进行高效的文件读取和哈希计算
func Checksum(filePath string, hashFunc func() hash.Hash) (string, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("文件不存在或无法访问: %v", err)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file failed: %v\n", err)
		}
	}()

	// 创建哈希对象
	hash := hashFunc()

	// 根据文件大小动态分配缓冲区
	fileSize := fileInfo.Size()
	bufferSize := calculateBufferSize(fileSize)
	buffer := make([]byte, bufferSize)

	// 使用 io.CopyBuffer 进行高效复制并计算哈希
	if _, err := io.CopyBuffer(hash, file, buffer); err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 返回哈希值的十六进制表示
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// ChecksumProgress 计算文件哈希值(带进度条)
//
// 参数:
//   - filePath: 文件路径
//   - hashFunc: 哈希函数构造器
//
// 返回:
//   - string: 文件的十六进制哈希值
//   - error: 错误信息，如果计算失败
//
// 注意:
//   - 根据文件大小动态分配缓冲区以提高性能
//   - 支持任何实现hash.Hash接口的哈希算法
//   - 使用io.CopyBuffer进行高效的文件读取和哈希计算
func ChecksumProgress(filePath string, hashFunc func() hash.Hash) (string, error) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("文件不存在或无法访问: %v", err)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("无法打开文件: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("close file failed: %v\n", err)
		}
	}()

	// 创建哈希对象
	hash := hashFunc()

	// 根据文件大小动态分配缓冲区
	fileSize := fileInfo.Size()
	bufferSize := calculateBufferSize(fileSize)
	buffer := make([]byte, bufferSize)

	// 创建进度条
	bar := progressbar.NewOptions64(
		fileSize,                          // 总进度
		progressbar.OptionClearOnFinish(), // 完成后清除进度条
		progressbar.OptionSetDescription(file.Name()+" 计算中"), // 设置进度条描述
	)
	defer func() {
		// 完成进度条
		if err := bar.Finish(); err != nil {
			fmt.Printf("finish progress bar failed: %v\n", err)
		}

		// 关闭进度条
		if err := bar.Close(); err != nil {
			fmt.Printf("close progress bar failed: %v\n", err)
		}
	}()

	// 创建多路写入器
	multiWriter := io.MultiWriter(hash, bar)

	// 使用 io.CopyBuffer 进行高效复制并计算哈希
	if _, err := io.CopyBuffer(multiWriter, file, buffer); err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 获取哈希值的十六进制表示
	hashStr := hex.EncodeToString(hash.Sum(nil))

	// 返回哈希值的十六进制表示
	return hashStr, nil
}

// calculateBufferSize 根据文件大小计算最佳缓冲区大小
//
// 参数:
//   - fileSize: 文件大小（字节）
//
// 返回:
//   - int: 计算出的最佳缓冲区大小（字节）
//
// 注意:
//   - 小文件使用较小的缓冲区以节省内存
//   - 大文件使用较大的缓冲区以提高I/O效率
//   - 缓冲区大小范围：32KB - 4MB
func calculateBufferSize(fileSize int64) int {
	switch {
	case fileSize < 32*KB: // 小于 32KB 的文件使用 32KB 缓冲区
		return int(32 * KB)
	case fileSize < 128*KB: // 32KB-128KB 使用 64KB 缓冲区
		return int(64 * KB)
	case fileSize < 512*KB: // 128KB-512KB 使用 128KB 缓冲区
		return int(128 * KB)
	case fileSize < 1*MB: // 512KB-1MB 使用 256KB 缓冲区
		return int(256 * KB)
	case fileSize < 4*MB: // 1MB-4MB 使用 512KB 缓冲区
		return int(512 * KB)
	case fileSize < 16*MB: // 4MB-16MB 使用 1MB 缓冲区
		return int(1 * MB)
	case fileSize < 64*MB: // 16MB-64MB 使用 2MB 缓冲区
		return int(2 * MB)
	default: // 大于 64MB 的文件使用 4MB 缓冲区
		return int(4 * MB)
	}
}
