package cleanup

import (
	"fmt"

	"gitee.com/MM-Q/colorlib"
)

// ExampleIntegration 演示如何在备份执行后集成清理算法
func ExampleIntegration() {
	// 模拟一个备份任务
	task := NewBackupTaskAdapter(
		1,                  // 任务ID
		"database_backup",  // 任务名称
		"/backup/database", // 存储目录
		5,                  // 保留最新5个备份
		30,                 // 保留30天内的备份
	)

	// 创建颜色库
	cl := colorlib.New()

	// 备份文件扩展名
	backupFileExt := ".zip"

	fmt.Println("=== 备份任务执行后的清理集成示例 ===")

	// 1. 预览将要清理的文件
	fmt.Println("\n1. 清理预览:")
	filesToDelete, err := GetCleanupPreview(task, backupFileExt)
	if err != nil {
		fmt.Printf("预览失败: %v\n", err)
		return
	}

	if len(filesToDelete) == 0 {
		fmt.Println("  无需清理任何文件")
	} else {
		fmt.Printf("  将删除 %d 个文件:\n", len(filesToDelete))
		for _, file := range filesToDelete {
			fmt.Printf("    - %s (创建时间: %s, 大小: %s)\n",
				file.FileName,
				file.CreatedTime.Format("2006-01-02 15:04:05"),
				formatBytes(file.FileSize))
		}
	}

	// 2. 执行清理（带日志输出）
	fmt.Println("\n2. 执行清理:")
	if err := CleanupBackupFilesWithLogging(task, backupFileExt, cl); err != nil {
		fmt.Printf("清理失败: %v\n", err)
		return
	}

	fmt.Println("\n清理完成！")
}

// ExampleRunCommandIntegration 演示在 run 命令中的集成方式
func ExampleRunCommandIntegration() {
	fmt.Println(`
// 在 cmd/subcmd/run/run.go 的 executeTask 函数中集成清理算法：

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
`)
}

// ExampleCleanupCommand 演示独立的清理命令实现
func ExampleCleanupCommand() {
	fmt.Println(`
// 可以创建独立的清理命令 cmd/subcmd/cleanup/cleanup.go：

func CleanupCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
    // 获取要清理的任务
    tasks, err := getTasksToCleanup(db)
    if err != nil {
        return err
    }

    for _, task := range tasks {
        cl.Bluef("清理任务: %s (ID: %d)", task.Name, task.ID)
        
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
            cl.Redf("任务 %s 清理失败: %v", task.Name, err)
            continue
        }
        
        cl.Green("任务清理完成")
    }
    
    return nil
}
`)
}
