package db

import (
	"database/sql"
	"fmt"

	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
)

// queryGetAllBackupTasks SQL SELECT 语句，用于查询所有备份任务
const queryGetAllBackupTasks = `
	SELECT
		ID,
		name,
		retain_count,
		retain_days,
		backup_dir,
		storage_dir,
		compress,
		include_rules,
		exclude_rules,
		max_file_size,
		min_file_size
	FROM backup_tasks
`

// GetAllBackupTasks 从数据库中获取所有备份任务。
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupTask：所有备份任务的切片
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllBackupTasks(db *sqlx.DB) ([]types.BackupTask, error) {
	var tasks []types.BackupTask
	err := db.Select(&tasks, queryGetAllBackupTasks)
	if err != nil {
		return nil, fmt.Errorf("获取所有备份任务失败: %w", err)
	}
	return tasks, nil
}

// queryGetAllBackupRecords SQL SELECT 语句，用于查询所有备份记录
const queryGetAllBackupRecords = `
	SELECT
		task_id,
		task_name,
		version_id,
		backup_filename,
		backup_size,
		status,
		failure_message,
		checksum,
		storage_path,
		created_at
	FROM backup_records
	ORDER BY created_at DESC
`

// GetAllBackupRecords 从数据库中获取所有备份记录。
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupRecord：所有备份记录的切片
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllBackupRecords(db *sqlx.DB) ([]types.BackupRecord, error) {
	var records []types.BackupRecord
	err := db.Select(&records, queryGetAllBackupRecords)
	if err != nil {
		return nil, fmt.Errorf("获取所有备份记录失败: %w", err)
	}
	return records, nil
}

// GetTaskIDByName 根据任务名称从数据库中获取任务ID。
// 如果找到任务，返回任务ID和nil错误；如果未找到，返回0和sql.ErrNoRows错误。
//
// 参数：
//   - db：数据库连接对象
//   - taskName：任务名称
//
// 返回值：
//   - int64：任务ID
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskIDByName(db *sqlx.DB, taskName string) (int64, error) {
	var taskID int64
	query := `SELECT ID FROM backup_tasks WHERE name = ?`
	err := db.Get(&taskID, query, taskName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("未找到名称为 '%s' 的任务", taskName)
		}
		return 0, fmt.Errorf("根据任务名称 '%s' 获取ID失败: %w", taskName, err)
	}
	return taskID, nil
}

// GetTaskNameByID 根据任务ID从数据库中获取任务名称。
// 如果找到任务，返回任务名称和nil错误；如果未找到，返回空字符串和更具体的错误信息。
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - string：任务名称
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskNameByID(db *sqlx.DB, taskID int64) (string, error) {
	var taskName string
	query := `SELECT name FROM backup_tasks WHERE ID = ?`
	err := db.Get(&taskName, query, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("未找到ID为 %d 的任务", taskID)
		}
		return "", fmt.Errorf("根据任务ID %d 获取名称失败: %w", taskID, err)
	}
	return taskName, nil
}

// GetTaskByID 根据任务ID从数据库中获取完整的任务信息。
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - *types.BackupTask：任务信息
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTaskByID(db *sqlx.DB, taskID int64) (*types.BackupTask, error) {
	var task types.BackupTask
	query := `SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, compress, include_rules, exclude_rules, max_file_size, min_file_size FROM backup_tasks WHERE ID = ?`
	err := db.Get(&task, query, taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到ID为 %d 的任务", taskID)
		}
		return nil, fmt.Errorf("根据任务ID %d 获取任务信息失败: %w", taskID, err)
	}
	return &task, nil
}

// GetTasksByIDs 根据任务ID列表从数据库中批量获取任务信息。
//
// 参数：
//   - db：数据库连接对象
//   - taskIDs：任务ID字符串切片
//
// 返回值：
//   - []types.BackupTask：任务信息列表
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTasksByIDs(db *sqlx.DB, taskIDs []string) ([]types.BackupTask, error) {
	if len(taskIDs) == 0 {
		return []types.BackupTask{}, nil
	}

	// 将字符串ID转换为interface{}切片，供sqlx.In使用
	args := make([]interface{}, len(taskIDs))
	for i, id := range taskIDs {
		args[i] = id
	}

	// 使用sqlx.In展开参数
	query := `SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, compress, include_rules, exclude_rules, max_file_size, min_file_size FROM backup_tasks WHERE ID IN (?)`
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("构建批量查询SQL失败: %w", err)
	}

	// 重新绑定查询参数以适配数据库驱动
	query = db.Rebind(query)

	// 执行查询
	var tasks []types.BackupTask
	err = db.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("批量获取任务信息失败: %w", err)
	}

	return tasks, nil
}

// GetAllTasks 从数据库中获取所有任务信息。
//
// 参数：
//   - db：数据库连接对象
//
// 返回值：
//   - []types.BackupTask：所有任务信息列表
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllTasks(db *sqlx.DB) ([]types.BackupTask, error) {
	var tasks []types.BackupTask
	query := `SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, compress, include_rules, exclude_rules, max_file_size, min_file_size FROM backup_tasks ORDER BY ID`

	err := db.Select(&tasks, query)
	if err != nil {
		return nil, fmt.Errorf("获取所有任务信息失败: %w", err)
	}

	return tasks, nil
}
