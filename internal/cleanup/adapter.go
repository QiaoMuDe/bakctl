package cleanup

// BackupTaskAdapter 备份任务适配器
// 用于将具体的备份任务结构体适配到清理算法的接口
type BackupTaskAdapter struct {
	ID          int64
	Name        string
	StorageDir  string
	RetainCount int
	RetainDays  int
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
