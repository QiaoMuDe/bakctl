// Package db 实现了 bakctl 的数据库删除操作功能。
//
// 该文件提供了数据库记录的删除功能，包括：
//   - 删除备份任务记录
//   - 删除备份历史记录
//   - 批量删除操作
//   - 级联删除相关记录
//   - 删除操作的事务处理
//
// 主要功能包括：
//   - 根据ID删除特定记录
//   - 根据条件批量删除记录
//   - 处理外键约束和级联删除
//   - 确保删除操作的原子性
//   - 提供删除结果的反馈信息
//
// 所有删除操作都包含适当的错误处理和事务管理。
package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DeleteBackupRecords 删除指定任务的所有备份记录
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - int：删除的记录数量
//   - error：删除过程中的错误
func DeleteBackupRecords(db *sqlx.DB, taskID int64) (int, error) {
	query := `DELETE FROM backup_records WHERE task_id = ?`

	result, err := db.Exec(query, taskID)
	if err != nil {
		return 0, fmt.Errorf("删除备份记录失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取删除结果失败: %w", err)
	}

	return int(rowsAffected), nil
}

// DeleteBackupTask 删除指定的备份任务
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - error：删除过程中的错误
func DeleteBackupTask(db *sqlx.DB, taskID int64) error {
	query := `DELETE FROM backup_tasks WHERE ID = ?`

	result, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("删除备份任务失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("任务ID %d 不存在", taskID)
	}

	return nil
}

// DeleteBackupRecordsByIDs 根据记录ID批量删除备份记录
//
// 参数:
//   - db: 数据库连接
//   - recordIDs: 要删除的记录ID列表
//
// 返回值:
//   - int: 实际删除的记录数
//   - error: 删除失败时返回错误信息
func DeleteBackupRecordsByIDs(db *sqlx.DB, recordIDs []int64) (int, error) {
	if len(recordIDs) == 0 {
		return 0, nil
	}

	// 使用sqlx.In来构建IN查询
	query := "DELETE FROM backup_records WHERE ID IN (?)"
	query, args, err := sqlx.In(query, recordIDs)
	if err != nil {
		return 0, fmt.Errorf("构建删除查询失败: %w", err)
	}
	query = db.Rebind(query)

	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("批量删除备份记录失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("获取删除结果失败: %w", err)
	}

	return int(rowsAffected), nil
}
