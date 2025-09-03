package db

import (
	"database/sql"
	"fmt"

	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
)

// TaskExists 检查指定ID的任务是否存在
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - bool：任务是否存在
func TaskExists(db *sqlx.DB, taskID int64) bool {
	var count int
	query := `SELECT COUNT(*) FROM backup_tasks WHERE ID = ?`
	err := db.Get(&count, query, taskID)
	return err == nil && count > 0
}

// GetBackupRecordByTaskAndVersion 根据任务ID和版本ID获取备份记录
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//   - versionID：版本ID
//
// 返回值：
//   - *types.BackupRecord：备份记录，如果未找到则返回nil
//   - error：查询过程中的错误
func GetBackupRecordByTaskAndVersion(db *sqlx.DB, taskID int64, versionID string) (*types.BackupRecord, error) {
	query := `
		SELECT ID, task_id, task_name, version_id, backup_filename, backup_size, 
		       storage_path, status, failure_message, checksum, created_at
		FROM backup_records 
		WHERE task_id = ? AND version_id = ? AND status = 1
	`

	var record types.BackupRecord
	err := db.Get(&record, query, taskID, versionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到指定的备份记录 (任务ID: %d, 版本ID: %s)", taskID, versionID)
		}
		return nil, fmt.Errorf("查询备份记录失败: %w", err)
	}

	return &record, nil
}

// queryGetAllBackupRecords SQL SELECT 语句，用于查询所有备份记录
const queryGetAllBackupRecords = `
	SELECT
		ID,
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

// GetAllBackupRecordsWithLimit 从数据库中获取指定数量的备份记录。
//
// 参数：
//   - db：数据库连接对象
//   - limit：限制返回的记录数量，0表示不限制
//
// 返回值：
//   - []types.BackupRecord：备份记录的切片
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetAllBackupRecordsWithLimit(db *sqlx.DB, limit int) ([]types.BackupRecord, error) {
	query := queryGetAllBackupRecords
	var args []interface{}

	// 如果指定了limit，添加LIMIT子句
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	var records []types.BackupRecord
	err := db.Select(&records, query, args...)
	if err != nil {
		return nil, fmt.Errorf("获取备份记录失败: %w", err)
	}
	return records, nil
}

// GetBatchBackupRecords 批量获取多个任务的备份记录
//
// 参数：
//   - db：数据库连接对象
//   - tasks：任务列表
//
// 返回值：
//   - map[int64][]types.BackupRecord：任务ID为键，备份记录列表为值
//   - error：查询过程中的错误
func GetBatchBackupRecords(db *sqlx.DB, tasks []types.BackupTask) (map[int64][]types.BackupRecord, error) {
	if len(tasks) == 0 {
		return make(map[int64][]types.BackupRecord), nil
	}

	// 提取任务ID
	taskIDs := make([]int64, len(tasks))
	for i, task := range tasks {
		taskIDs[i] = task.ID
	}

	query := `
		SELECT ID, task_id, task_name, version_id, backup_filename, backup_size, 
		       storage_path, status, failure_message, checksum, created_at
		FROM backup_records 
		WHERE task_id IN (?)
		ORDER BY task_id, created_at DESC
	`

	// 使用 sqlx.In 构建 IN 查询
	query, args, err := sqlx.In(query, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("构建批量查询失败: %w", err)
	}
	query = db.Rebind(query)

	var allRecords []types.BackupRecord
	err = db.Select(&allRecords, query, args...)
	if err != nil {
		return nil, fmt.Errorf("批量获取备份记录失败: %w", err)
	}

	// 按任务ID分组
	recordsByTask := make(map[int64][]types.BackupRecord)
	for _, record := range allRecords {
		taskID := record.TaskID
		recordsByTask[taskID] = append(recordsByTask[taskID], record)
	}

	return recordsByTask, nil
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
//   - taskIDs：任务ID整数切片
//
// 返回值：
//   - []types.BackupTask：任务信息列表
//   - error：如果获取过程中发生错误，则返回非 nil 错误信息
func GetTasksByIDs(db *sqlx.DB, taskIDs []int64) ([]types.BackupTask, error) {
	if len(taskIDs) == 0 {
		return []types.BackupTask{}, nil
	}

	// 使用sqlx.In展开参数
	query := `SELECT ID, name, retain_count, retain_days, backup_dir, storage_dir, compress, include_rules, exclude_rules, max_file_size, min_file_size FROM backup_tasks WHERE ID IN (?)`
	query, args, err := sqlx.In(query, taskIDs)
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

// GetBackupRecordsByTaskIDWithLimit 根据任务ID获取指定数量的备份记录
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//   - limit：限制返回的记录数量，0表示不限制
//
// 返回值：
//   - []types.BackupRecord：备份记录列表
//   - error：查询过程中的错误
func GetBackupRecordsByTaskIDWithLimit(db *sqlx.DB, taskID int64, limit int) ([]types.BackupRecord, error) {
	query := `
		SELECT id, task_id, task_name, version_id, backup_filename, backup_size, storage_path, status, failure_message, checksum, created_at
		FROM backup_records 
		WHERE task_id = ?
		ORDER BY created_at DESC
	`
	args := []interface{}{taskID}

	// 如果指定了limit，添加LIMIT子句
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	var records []types.BackupRecord
	err := db.Select(&records, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询备份记录失败: %w", err)
	}

	return records, nil
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

// GetBackupRecordsByTaskID 根据任务ID获取所有备份记录
//
// 参数：
//   - db：数据库连接对象
//   - taskID：任务ID
//
// 返回值：
//   - []types.BackupRecord：备份记录列表
//   - error：查询过程中的错误
func GetBackupRecordsByTaskID(db *sqlx.DB, taskID int64) ([]types.BackupRecord, error) {
	query := `
		SELECT id, task_id, task_name, version_id, backup_filename, backup_size,
		       storage_path, status, failure_message, checksum, created_at
		FROM backup_records 
		WHERE task_id = ?
		ORDER BY created_at DESC
	`

	var records []types.BackupRecord
	err := db.Select(&records, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("查询备份记录失败: %w", err)
	}

	return records, nil
}
