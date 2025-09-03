// Package db 实现了 bakctl 的数据库清理功能。
//
// 该文件提供了数据库中孤儿记录的清理功能，包括：
//   - 检测和清理孤儿备份记录
//   - 验证备份文件的存在性
//   - 删除无效的数据库记录
//   - 维护数据库的数据一致性
//   - 提供清理操作的统计信息
//
// 主要功能包括：
//   - 扫描数据库中的备份记录
//   - 验证对应的备份文件是否存在
//   - 删除文件已不存在的孤儿记录
//   - 生成清理操作的详细报告
//   - 确保数据库操作的事务性
//
// 孤儿记录清理有助于保持数据库的整洁和数据的一致性。
package db

import (
	"fmt"
	"os"

	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
)

// CleanupOrphanRecords 清理孤儿备份记录（文件不存在但数据库有记录）
//
// 参数:
//   - db: 数据库连接
//   - taskID: 任务ID（0表示清理所有任务的孤儿记录）
//
// 返回值:
//   - int: 清理的记录数
//   - error: 清理失败时返回错误信息
func CleanupOrphanRecords(db *sqlx.DB, taskID int64) (int, error) {
	var records []types.BackupRecord
	var err error

	// 1. 获取备份记录
	if taskID > 0 {
		records, err = GetBackupRecordsByTaskID(db, taskID) // 如果指定任务ID，只获取该任务的记录
	} else {
		records, err = GetAllBackupRecords(db) // 否则获取所有任务的记录
	}

	if err != nil {
		return 0, fmt.Errorf("获取备份记录失败: %w", err)
	}

	if len(records) == 0 {
		return 0, nil // 没有记录需要清理
	}

	// 2. 遍历记录，检查文件是否存在，收集孤儿记录的ID
	var orphanIDs []int64
	for _, record := range records {
		// 跳过失败的备份记录（status为false的记录本身就不应该有对应文件）
		if !record.Status {
			continue
		}

		// 检查StoragePath指向的文件是否存在
		if _, err := os.Stat(record.StoragePath); os.IsNotExist(err) {
			orphanIDs = append(orphanIDs, record.ID)
		}
	}

	// 3. 如果没有孤儿记录，直接返回
	if len(orphanIDs) == 0 {
		return 0, nil
	}

	// 4. 根据ID批量删除孤儿记录（复用现有的函数）
	deletedCount, err := DeleteBackupRecordsByIDs(db, orphanIDs)
	if err != nil {
		return 0, fmt.Errorf("删除孤儿记录失败: %w", err)
	}

	return deletedCount, nil
}
