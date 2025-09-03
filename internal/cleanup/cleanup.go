// Package cleanup 提供备份文件清理功能。
//
// 该包实现了基于保留策略的备份文件清理算法，支持按数量和时间两种维度进行清理。
// 主要功能包括：
//   - 收集指定目录下的备份文件信息
//   - 根据保留数量和保留天数策略确定需要删除的文件
//   - 执行文件删除操作并提供详细的清理结果统计
//   - 支持文件名格式解析和时间戳提取
//
// 清理策略说明：
//   - retainCount: 保留最新的N个备份文件（按时间戳排序）
//   - retainDays: 保留最近N天内的备份文件
//   - 同时设置时：先按天数过滤，然后每天只保留最新的N个备份文件
//   - 当两个策略都为0时，不执行任何清理操作
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
	CreatedTime time.Time // 创建时间
}

// CleanupResult 清理结果
type CleanupResult struct {
	TotalFiles   int      // 总文件数
	DeletedFiles int      // 删除的文件数
	ErrorFiles   []string // 删除失败的文件路径列表
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
		ErrorFiles: make([]string, 0),
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

	// 2. 按创建时间降序排序（最新的在前面）
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].CreatedTime.After(backupFiles[j].CreatedTime)
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
		}
	}

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

		// 创建备份文件信息
		backupFileInfo := BackupFileInfo{
			FilePath:    filePath,    // 文件路径
			CreatedTime: createdTime, // 创建时间
		}

		backupFiles = append(backupFiles, backupFileInfo)
	}

	return backupFiles, nil
}

// determineFilesToDelete 根据保留策略确定需要删除的文件
//
// 清理策略：
//   - 只设置retainCount：保留最新的N个备份文件
//   - 只设置retainDays：保留最近N天内的所有备份文件
//   - 同时设置：先按天数过滤，然后每天只保留最新的N个备份文件
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

	// 当前时间
	now := time.Now()

	// 情况1：只设置了保留数量，不限制天数
	if retainCount > 0 && retainDays <= 0 {
		// 保留最新的 retainCount 个文件，删除其余的
		if len(backupFiles) > retainCount {
			filesToDelete = backupFiles[retainCount:]
		}
		return filesToDelete
	}

	// 情况2：只设置了保留天数，不限制数量
	if retainDays > 0 && retainCount <= 0 {
		cutoffTime := now.AddDate(0, 0, -retainDays)
		for _, fileInfo := range backupFiles {
			if fileInfo.CreatedTime.Before(cutoffTime) || fileInfo.CreatedTime.Equal(cutoffTime) {
				filesToDelete = append(filesToDelete, fileInfo)
			}
		}
		return filesToDelete
	}

	// 情况3：同时设置了保留数量和保留天数
	if retainCount > 0 && retainDays > 0 {
		return determineFilesToDeleteWithBothPolicies(backupFiles, retainCount, retainDays, now)
	}

	return filesToDelete
}

// determineFilesToDeleteWithBothPolicies 处理同时设置保留数量和保留天数的情况
//
// 逻辑：先按天数过滤，然后每天只保留最新的N个备份文件
//
// 参数:
//   - backupFiles: 备份文件信息列表（已按时间戳降序排序）
//   - retainCount: 保留备份数量
//   - retainDays: 保留天数
//   - now: 当前时间
//
// 返回值:
//   - []BackupFileInfo: 需要删除的文件信息列表
func determineFilesToDeleteWithBothPolicies(backupFiles []BackupFileInfo, retainCount, retainDays int, now time.Time) []BackupFileInfo {
	var filesToDelete []BackupFileInfo

	// 1. 先按天数过滤：超过保留天数的文件直接删除
	cutoffTime := now.AddDate(0, 0, -retainDays)
	var filesWithinDays []BackupFileInfo

	for _, fileInfo := range backupFiles {
		if fileInfo.CreatedTime.Before(cutoffTime) || fileInfo.CreatedTime.Equal(cutoffTime) {
			// 超过保留天数，直接删除
			filesToDelete = append(filesToDelete, fileInfo)
		} else {
			// 在保留天数内，加入候选列表
			filesWithinDays = append(filesWithinDays, fileInfo)
		}
	}

	// 2. 对保留天数内的文件按日期分组
	dailyGroups := groupFilesByDate(filesWithinDays)

	// 3. 每天只保留最新的 retainCount 个文件
	for _, dailyFiles := range dailyGroups {
		// 每天的文件已经按时间降序排序，保留前 retainCount 个
		if len(dailyFiles) > retainCount {
			// 删除超出保留数量的文件
			filesToDelete = append(filesToDelete, dailyFiles[retainCount:]...)
		}
	}

	return filesToDelete
}

// groupFilesByDate 按日期对备份文件进行分组
//
// 参数:
//   - backupFiles: 备份文件信息列表
//
// 返回值:
//   - map[string][]BackupFileInfo: 按日期分组的文件映射，key为日期字符串(YYYY-MM-DD)
func groupFilesByDate(backupFiles []BackupFileInfo) map[string][]BackupFileInfo {
	dailyGroups := make(map[string][]BackupFileInfo)

	for _, fileInfo := range backupFiles {
		// 获取日期字符串 (YYYY-MM-DD)
		dateKey := fileInfo.CreatedTime.Format("2006-01-02")

		// 添加到对应日期的组中
		dailyGroups[dateKey] = append(dailyGroups[dateKey], fileInfo)
	}

	// 对每天的文件按时间降序排序（最新的在前面）
	for dateKey := range dailyGroups {
		sort.Slice(dailyGroups[dateKey], func(i, j int) bool {
			return dailyGroups[dateKey][i].CreatedTime.After(dailyGroups[dateKey][j].CreatedTime)
		})
	}

	return dailyGroups
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
