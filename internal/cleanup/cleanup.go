package cleanup

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// BackupFileInfo 备份文件信息
type BackupFileInfo struct {
	FilePath    string    // 文件完整路径
	FileName    string    // 文件名
	TaskName    string    // 任务名称
	Timestamp   int64     // 时间戳
	CreatedTime time.Time // 创建时间
	FileSize    int64     // 文件大小
}

// CleanupResult 清理结果
type CleanupResult struct {
	TotalFiles    int      // 总文件数
	DeletedFiles  int      // 删除的文件数
	RetainedFiles int      // 保留的文件数
	DeletedPaths  []string // 删除的文件路径列表
	ErrorFiles    []string // 删除失败的文件路径列表
	TotalSize     int64    // 删除的总文件大小
}

// CleanupBackupFiles 清理历史备份文件
//
// 参数:
//   - storageDir: 备份存储目录
//   - taskName: 任务名称
//   - retainCount: 保留备份数量 (0表示不限制数量)
//   - retainDays: 保留天数 (0表示不限制天数)
//   - backupFileExt: 备份文件扩展名 (如 ".zip")
//
// 返回值:
//   - CleanupResult: 清理结果统计
//   - error: 清理过程中的错误
func CleanupBackupFiles(storageDir, taskName string, retainCount, retainDays int, backupFileExt string) (CleanupResult, error) {
	result := CleanupResult{
		DeletedPaths: make([]string, 0),
		ErrorFiles:   make([]string, 0),
	}

	// 如果两个保留策略都为0，则不进行清理
	if retainCount <= 0 && retainDays <= 0 {
		return result, nil
	}

	// 1. 收集备份文件信息
	backupFiles, err := collectBackupFiles(storageDir, taskName, backupFileExt)
	if err != nil {
		return result, fmt.Errorf("收集备份文件失败: %w", err)
	}

	result.TotalFiles = len(backupFiles)

	// 如果没有备份文件，直接返回
	if len(backupFiles) == 0 {
		return result, nil
	}

	// 2. 按时间戳降序排序（最新的在前面）
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].Timestamp > backupFiles[j].Timestamp
	})

	// 3. 确定需要删除的文件
	filesToDelete := determineFilesToDelete(backupFiles, retainCount, retainDays)

	// 4. 执行删除操作
	for _, fileInfo := range filesToDelete {
		// 删除文件
		if err := os.Remove(fileInfo.FilePath); err != nil {
			result.ErrorFiles = append(result.ErrorFiles, fileInfo.FilePath)
		} else { // 删除成功
			result.DeletedFiles++
			result.DeletedPaths = append(result.DeletedPaths, fileInfo.FilePath)
			result.TotalSize += fileInfo.FileSize
		}
	}

	// 5. 统计保留的文件数
	result.RetainedFiles = result.TotalFiles - result.DeletedFiles

	return result, nil
}

// collectBackupFiles 收集指定目录下的备份文件信息
//
// 参数:
//   - storageDir: 备份存储目录
//   - taskName: 任务名称
//   - backupFileExt: 备份文件扩展名
//
// 返回值:
//   - []BackupFileInfo: 备份文件信息列表
//   - error: 收集过程中的错误
func collectBackupFiles(storageDir, taskName, backupFileExt string) ([]BackupFileInfo, error) {
	var backupFiles []BackupFileInfo

	// 检查目录是否存在
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		return backupFiles, nil // 目录不存在，返回空列表
	}

	// 构建文件名匹配模式: {taskName}_{YYYYMMDD_HHMMSS}.zip
	// 使用正则表达式匹配: taskName_时间字符串.zip
	pattern := fmt.Sprintf(`^%s_(\d{8}_\d{6})%s$`, regexp.QuoteMeta(taskName), regexp.QuoteMeta(backupFileExt))
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("编译正则表达式失败: %w", err)
	}

	// 遍历目录中的文件
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	// 如果目录为空，直接返回
	if len(entries) == 0 {
		return backupFiles, nil
	}

	// 遍历目录中的文件
	for _, entry := range entries {
		// 跳过目录
		if entry.IsDir() {
			continue
		}

		// 获取文件名
		fileName := entry.Name()

		// 检查文件名是否匹配备份文件格式
		matches := regex.FindStringSubmatch(fileName)
		if len(matches) != 2 {
			continue // 不匹配，跳过
		}

		// 解析时间字符串 (格式: YYYYMMDD_HHMMSS)
		timeStr := matches[1]
		createdTime, err := time.Parse("20060102_150405", timeStr)
		if err != nil {
			continue // 时间字符串解析失败，跳过
		}

		// 获取文件完整路径
		filePath := filepath.Join(storageDir, fileName)

		// 获取文件信息
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			continue // 获取文件信息失败，跳过
		}

		// 创建备份文件信息
		backupFileInfo := BackupFileInfo{
			FilePath:    filePath,           // 文件路径
			FileName:    fileName,           // 文件名
			TaskName:    taskName,           // 任务名称
			Timestamp:   createdTime.Unix(), // 转换为Unix时间戳用于兼容
			CreatedTime: createdTime,        // 创建时间
			FileSize:    fileInfo.Size(),    // 文件大小
		}

		backupFiles = append(backupFiles, backupFileInfo)
	}

	return backupFiles, nil
}

// determineFilesToDelete 根据保留策略确定需要删除的文件
//
// 参数:
//   - backupFiles: 备份文件信息列表（已按时间戳降序排序）
//   - retainCount: 保留备份数量 (0表示不限制数量)
//   - retainDays: 保留天数 (0表示不限制天数)
//
// 返回值:
//   - []BackupFileInfo: 需要删除的文件信息列表
func determineFilesToDelete(backupFiles []BackupFileInfo, retainCount, retainDays int) []BackupFileInfo {
	var filesToDelete []BackupFileInfo

	// 快速失败：如果两个策略都没有设置，保留所有文件
	if retainCount <= 0 && retainDays <= 0 {
		return filesToDelete // 返回空列表
	}

	// 快速失败：如果没有备份文件，直接返回
	if len(backupFiles) == 0 {
		return filesToDelete
	}

	// 快速失败：如果只有一个文件，不删除任何文件
	if len(backupFiles) <= 1 {
		return filesToDelete
	}

	now := time.Now()

	// 创建保留文件的映射，用于去重
	retainedFiles := make(map[string]bool)

	// 1. 根据保留数量策略确定要保留的文件
	if retainCount > 0 {
		// 保留最新的 retainCount 个文件（但不超过现有文件数量）
		for i := 0; i < retainCount && i < len(backupFiles); i++ {
			retainedFiles[backupFiles[i].FilePath] = true
		}
	}

	// 2. 根据保留天数策略确定要保留的文件
	if retainDays > 0 {
		cutoffTime := now.AddDate(0, 0, -retainDays)
		for _, fileInfo := range backupFiles {
			if fileInfo.CreatedTime.After(cutoffTime) {
				retainedFiles[fileInfo.FilePath] = true
			}
		}
	}

	// 4. 安全检查：如果没有文件被标记为保留，至少保留最新的一个
	if len(retainedFiles) == 0 && len(backupFiles) > 0 {
		retainedFiles[backupFiles[0].FilePath] = true
	}

	// 5. 确定需要删除的文件（不在保留列表中的文件）
	for _, fileInfo := range backupFiles {
		if !retainedFiles[fileInfo.FilePath] {
			filesToDelete = append(filesToDelete, fileInfo)
		}
	}

	// 6. 最终安全检查：确保不会删除所有文件
	if len(filesToDelete) >= len(backupFiles) && len(backupFiles) > 0 {
		// 如果要删除的文件数量等于或超过总文件数，只保留最新的一个
		filesToDelete = backupFiles[1:] // 保留第一个（最新的），删除其余的
	}

	return filesToDelete
}

// FormatCleanupResult 格式化清理结果为可读字符串
//
// 参数:
//   - result: 清理结果
//
// 返回值:
//   - string: 格式化后的结果字符串
func FormatCleanupResult(result CleanupResult) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("清理完成: 总文件 %d 个", result.TotalFiles))

	if result.DeletedFiles > 0 {
		builder.WriteString(fmt.Sprintf(", 删除 %d 个", result.DeletedFiles))
		builder.WriteString(fmt.Sprintf(", 保留 %d 个", result.RetainedFiles))
		builder.WriteString(fmt.Sprintf(", 释放空间 %s", formatBytes(result.TotalSize)))
	} else {
		builder.WriteString(", 无需清理")
	}

	if len(result.ErrorFiles) > 0 {
		builder.WriteString(fmt.Sprintf(", 删除失败 %d 个", len(result.ErrorFiles)))
	}

	return builder.String()
}

// ValidateCleanupParams 验证清理参数的合法性
//
// 参数:
//   - storageDir: 备份存储目录
//   - taskName: 任务名称
//   - retainCount: 保留备份数量
//   - retainDays: 保留天数
//
// 返回值:
//   - error: 参数验证错误，nil表示验证通过
func ValidateCleanupParams(storageDir, taskName string, retainCount, retainDays int) error {
	// 验证存储目录
	if strings.TrimSpace(storageDir) == "" {
		return fmt.Errorf("存储目录不能为空")
	}

	// 验证任务名称
	if strings.TrimSpace(taskName) == "" {
		return fmt.Errorf("任务名称不能为空")
	}

	// 验证保留数量
	if retainCount < 0 {
		return fmt.Errorf("保留数量不能为负数: %d", retainCount)
	}

	// 验证保留天数
	if retainDays < 0 {
		return fmt.Errorf("保留天数不能为负数: %d", retainDays)
	}

	return nil
}

// formatBytes 格式化字节数为可读字符串
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
