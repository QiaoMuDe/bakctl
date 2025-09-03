// Package cleanup 提供备份任务适配器实现。
//
// 该文件实现了适配器模式，用于解决包之间的循环依赖问题。
// 通过定义BackupTask接口和BackupTaskAdapter结构体，
// 使得cleanup包可以独立于具体的备份任务实现，
// 同时为其他包提供了统一的任务信息访问接口。
//
// 主要组件：
//   - BackupTask: 备份任务接口，定义了获取任务信息的方法
//   - BackupTaskAdapter: 适配器实现，将具体任务结构体适配到接口
//   - NewBackupTaskAdapter: 工厂函数，创建适配器实例
package cleanup

// BackupTaskAdapter 备份任务适配器
// 用于将具体的备份任务结构体适配到清理算法的接口
type BackupTaskAdapter struct {
	ID          int64  // 任务ID
	Name        string // 任务名称
	StorageDir  string // 存储目录
	RetainCount int    // 保留数量
	RetainDays  int    // 保留天数
}

// GetID 获取任务ID
func (a *BackupTaskAdapter) GetID() int64 {
	return a.ID
}

// GetName 获取任务名称
func (a *BackupTaskAdapter) GetName() string {
	return a.Name
}

// GetStorageDir 获取存储目录
func (a *BackupTaskAdapter) GetStorageDir() string {
	return a.StorageDir
}

// GetRetainCount 获取保留数量
func (a *BackupTaskAdapter) GetRetainCount() int {
	return a.RetainCount
}

// GetRetainDays 获取保留天数
func (a *BackupTaskAdapter) GetRetainDays() int {
	return a.RetainDays
}

// NewBackupTaskAdapter 创建备份任务适配器
//
// 参数:
//   - id: 任务ID
//   - name: 任务名称
//   - storageDir: 存储目录
//   - retainCount: 保留数量
//   - retainDays: 保留天数
//
// 返回值:
//   - BackupTask: 备份任务接口实现
func NewBackupTaskAdapter(id int64, name, storageDir string, retainCount, retainDays int) BackupTask {
	return &BackupTaskAdapter{
		ID:          id,
		Name:        name,
		StorageDir:  storageDir,
		RetainCount: retainCount,
		RetainDays:  retainDays,
	}
}
