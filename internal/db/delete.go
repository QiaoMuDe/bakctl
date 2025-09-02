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
