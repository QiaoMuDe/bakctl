package cleanup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gitee.com/MM-Q/colorlib"
)

// 演示如何在实际项目中使用清理算法

// MockBackupTask 模拟的备份任务结构体（对应实际项目中的 types.BackupTask）
type MockBackupTask struct {
	ID          int64
	Name        string
	StorageDir  string
	RetainCount int
	RetainDays  int
}

// 实现 BackupTask 接口
func (t *MockBackupTask) GetID() int64          { return t.ID }
func (t *MockBackupTask) GetName() string       { return t.Name }
func (t *MockBackupTask) GetStorageDir() string { return t.StorageDir }
func (t *MockBackupTask) GetRetainCount() int   { return t.RetainCount }
func (t *MockBackupTask) GetRetainDays() int    { return t.RetainDays }

// DemoCleanupAlgorithm 演示清理算法的完整流程
func DemoCleanupAlgorithm() {
	fmt.Println("=== 备份文件清理算法演示 ===\n")

	// 1. 创建测试环境
	testDir := createTestEnvironment()
	defer cleanupTestEnvironment(testDir)

	// 2. 创建模拟任务
	task := &MockBackupTask{
		ID:          1,
		Name:        "demo_task",
		StorageDir:  testDir,
		RetainCount: 3, // 保留最新3个
		RetainDays:  7, // 保留7天内的
	}

	// 3. 创建颜色库
	cl := colorlib.New()
	cl.SetColor(true)

	fmt.Printf("任务信息:\n")
	fmt.Printf("  - 任务ID: %d\n", task.GetID())
	fmt.Printf("  - 任务名称: %s\n", task.GetName())
	fmt.Printf("  - 存储目录: %s\n", task.GetStorageDir())
	fmt.Printf("  - 保留数量: %d\n", task.GetRetainCount())
	fmt.Printf("  - 保留天数: %d\n", task.GetRetainDays())
	fmt.Println()

	// 4. 显示清理前的文件状态
	fmt.Println("清理前的备份文件:")
	showBackupFiles(task, ".zip")
	fmt.Println()

	// 5. 预览清理操作
	fmt.Println("清理预览:")
	previewCleanup(task, ".zip")
	fmt.Println()

	// 6. 执行清理
	fmt.Println("执行清理:")
	if err := CleanupBackupFilesWithLogging(task, ".zip", cl); err != nil {
		log.Printf("清理失败: %v", err)
		return
	}
	fmt.Println()

	// 7. 显示清理后的文件状态
	fmt.Println("清理后的备份文件:")
	showBackupFiles(task, ".zip")
	fmt.Println()

	fmt.Println("演示完成！")
}

// createTestEnvironment 创建测试环境
func createTestEnvironment() string {
	// 创建临时目录
	testDir := filepath.Join(os.TempDir(), "bakctl_cleanup_demo")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("创建测试目录失败: %v", err)
	}

	// 创建测试备份文件
	now := time.Now()
	testFiles := []struct {
		name    string
		age     time.Duration
		content string
	}{
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -15).Unix()) + ".zip", 15 * 24 * time.Hour, "15天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -10).Unix()) + ".zip", 10 * 24 * time.Hour, "10天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -8).Unix()) + ".zip", 8 * 24 * time.Hour, "8天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -5).Unix()) + ".zip", 5 * 24 * time.Hour, "5天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -3).Unix()) + ".zip", 3 * 24 * time.Hour, "3天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.AddDate(0, 0, -1).Unix()) + ".zip", 1 * 24 * time.Hour, "1天前的备份"},
		{"demo_task_" + fmt.Sprintf("%d", now.Unix()) + ".zip", 0, "今天的备份"},
		{"other_task_" + fmt.Sprintf("%d", now.Unix()) + ".zip", 0, "其他任务的备份"},
		{"readme.txt", 0, "说明文件"},
	}

	for _, file := range testFiles {
		filePath := filepath.Join(testDir, file.name)
		if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
			log.Printf("创建文件 %s 失败: %v", file.name, err)
		}
	}

	return testDir
}

// cleanupTestEnvironment 清理测试环境
func cleanupTestEnvironment(testDir string) {
	if err := os.RemoveAll(testDir); err != nil {
		log.Printf("清理测试目录失败: %v", err)
	}
}

// showBackupFiles 显示备份文件列表
func showBackupFiles(task BackupTask, backupFileExt string) {
	files, err := collectBackupFiles(task.GetStorageDir(), task.GetName(), backupFileExt)
	if err != nil {
		log.Printf("收集备份文件失败: %v", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("  (无备份文件)")
		return
	}

	// 按时间戳降序排序
	sortBackupFilesByTimestamp(files)

	for i, file := range files {
		age := time.Since(file.CreatedTime)
		fmt.Printf("  %d. %s\n", i+1, file.FileName)
		fmt.Printf("     创建时间: %s (%s前)\n",
			file.CreatedTime.Format("2006-01-02 15:04:05"),
			formatAge(age))
		fmt.Printf("     文件大小: %s\n", formatBytes(file.FileSize))
	}
}

// previewCleanup 预览清理操作
func previewCleanup(task BackupTask, backupFileExt string) {
	filesToDelete, err := GetCleanupPreview(task, backupFileExt)
	if err != nil {
		log.Printf("预览失败: %v", err)
		return
	}

	if len(filesToDelete) == 0 {
		fmt.Println("  无需清理任何文件")
		return
	}

	fmt.Printf("  将删除 %d 个文件:\n", len(filesToDelete))
	var totalSize int64
	for _, file := range filesToDelete {
		age := time.Since(file.CreatedTime)
		fmt.Printf("    - %s (%s前, %s)\n",
			file.FileName,
			formatAge(age),
			formatBytes(file.FileSize))
		totalSize += file.FileSize
	}
	fmt.Printf("  总计释放空间: %s\n", formatBytes(totalSize))
}

// formatAge 格式化时间间隔
func formatAge(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d天", days)
	}
	hours := int(duration.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d小时", hours)
	}
	minutes := int(duration.Minutes())
	return fmt.Sprintf("%d分钟", minutes)
}

// DemoIntegrationInRunCommand 演示在 run 命令中的集成方式
func DemoIntegrationInRunCommand() {
	fmt.Println(`
=== 在 run 命令中集成清理算法的示例代码 ===

// 在 cmd/subcmd/run/run.go 中修改 executeTask 函数：

func executeTask(task baktypes.BackupTask, db *sqlx.DB, cl *colorlib.ColorLib) error {
    // ... 现有的备份逻辑 ...
    
    // 8. 设置成功结果
    result.Success = true
    result.FileSize = size
    result.Checksum = checksum

    // 9. 清理历史备份
    cl.White("  → 清理历史备份...")
    
    // 创建任务适配器
    taskAdapter := cleanup.NewBackupTaskAdapter(
        task.ID,
        task.Name,
        task.StorageDir,
        task.RetainCount,
        task.RetainDays,
    )
    
    // 执行清理
    if err := cleanup.CleanupBackupFilesWithLogging(taskAdapter, baktypes.BackupFileExt, cl); err != nil {
        // 清理失败不影响备份成功，只记录警告
        cl.Yellowf("  → 清理警告: %v", err)
    }

    return nil
}

=== 使用说明 ===

1. 导入清理模块:
   import "gitee.com/MM-Q/bakctl/internal/cleanup"

2. 在备份成功后调用清理函数:
   - 使用 cleanup.NewBackupTaskAdapter() 创建适配器
   - 调用 cleanup.CleanupBackupFilesWithLogging() 执行清理

3. 清理策略:
   - retain_count > 0: 保留最新的 N 个备份文件
   - retain_days > 0: 保留最近 N 天内的备份文件
   - 两个策略可以同时生效，取并集
   - 两个策略都为 0 时不进行清理

4. 安全特性:
   - 只处理匹配格式的备份文件
   - 清理失败不影响备份任务的成功状态
   - 提供详细的操作日志和错误信息
`)
}

// TestDifferentScenarios 测试不同的清理场景
func TestDifferentScenarios() {
	scenarios := []struct {
		name        string
		retainCount int
		retainDays  int
		description string
	}{
		{"只按数量清理", 2, 0, "只保留最新的2个备份文件"},
		{"只按天数清理", 0, 3, "只保留3天内的备份文件"},
		{"数量和天数都限制", 3, 7, "保留最新3个备份且保留7天内的备份"},
		{"不清理", 0, 0, "不进行任何清理"},
		{"极端情况-保留1个", 1, 0, "只保留最新的1个备份文件"},
		{"极端情况-保留1天", 0, 1, "只保留1天内的备份文件"},
	}

	fmt.Println("=== 不同清理场景测试 ===\n")

	for i, scenario := range scenarios {
		fmt.Printf("%d. 场景: %s\n", i+1, scenario.name)
		fmt.Printf("   描述: %s\n", scenario.description)
		fmt.Printf("   参数: retainCount=%d, retainDays=%d\n", scenario.retainCount, scenario.retainDays)

		// 创建测试任务
		task := &MockBackupTask{
			ID:          int64(i + 1),
			Name:        fmt.Sprintf("test_task_%d", i+1),
			StorageDir:  "/tmp/test",
			RetainCount: scenario.retainCount,
			RetainDays:  scenario.retainDays,
		}

		// 验证参数
		if err := ValidateCleanupParams(task.GetStorageDir(), task.GetName(), task.GetRetainCount(), task.GetRetainDays()); err != nil {
			fmt.Printf("   结果: 参数验证失败 - %v\n", err)
		} else {
			fmt.Printf("   结果: 参数验证通过\n")
		}

		fmt.Println()
	}
}
