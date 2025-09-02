package log

import (
	"fmt"
	"os"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/bakctl/internal/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jmoiron/sqlx"
)

// LogCmdMain 日志命令主函数
func LogCmdMain(db *sqlx.DB) error {
	// 参数验证
	if err := validateLogFlags(); err != nil {
		return err
	}

	// 创建表格
	t := table.NewWriter()

	// 设置表格样式
	if style, ok := types.TableStyle[logCmdTableStyle.Get()]; ok {
		t.SetStyle(style)
	} else {
		return fmt.Errorf("表格样式不存在: %s, 可选样式: %v", logCmdTableStyle.Get(), types.TableStyleList)
	}

	// 查询备份记录列表
	data, err := getBackupRecords(db)
	if err != nil {
		return fmt.Errorf("查询备份记录失败: %w", err)
	}

	// 提前检查是否有备份记录
	if len(data) == 0 {
		fmt.Println("没有备份记录")
		return nil
	}

	// 使用标准输出作为输出目标
	t.SetOutputMirror(os.Stdout)

	// 设置表头
	t.AppendHeader(table.Row{"任务ID", "任务名", "版本ID", "备份文件名", "文件大小", "存储路径", "状态", "失败信息", "校验码", "创建时间"})

	// 设置列配置
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "任务ID", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
		{Name: "任务名", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "版本ID", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "备份文件名", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "文件大小", Align: text.AlignRight, WidthMaxEnforcer: text.WrapHard},
		{Name: "存储路径", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "状态", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
		{Name: "失败信息", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "校验码", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		{Name: "创建时间", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
	})

	// 添加数据行
	for _, record := range data {
		t.AppendRow(table.Row{
			record.TaskID,    // 任务ID
			record.TaskName,  // 任务名
			record.VersionID, // 版本ID
			func() string {
				// 如果备份文件名不为空，则返回备份文件名，否则返回 "-"
				if record.BackupFilename != "" {
					return record.BackupFilename
				}
				return "---"
			}(), // 备份文件名
			utils.FormatBytes(record.BackupSize), // 文件大小
			record.StoragePath,                   // 存储路径
			record.Status,                        // 状态
			func() string {
				// 如果失败信息不为空，则返回失败信息，否则返回 "-"
				if record.FailureMessage != "" {
					return record.FailureMessage
				}
				return "---"
			}(), // 失败信息
			func() string {
				// 如果校验码不为空，则返回校验码，否则返回 "-"
				if record.Checksum != "" {
					return record.Checksum
				}
				return "---"
			}(), // 校验码
			record.CreatedAt, // 创建时间
		})
	}

	// 渲染表格
	t.Render()

	return nil
}

// validateLogFlags 验证log命令的标志参数
func validateLogFlags() error {
	// 验证任务选择
	hasID := logCmdTaskID.Get() > 0       // 任务ID必须大于0
	hasName := logCmdTaskName.Get() != "" // 任务名称不能为空

	if hasID && hasName {
		return fmt.Errorf("--id 和 --name/-n 不能同时使用")
	}

	// 验证limit参数
	if logCmdLimit.Get() < 0 {
		return fmt.Errorf("--limit/-l 必须大于等于0")
	}

	return nil
}

// getBackupRecords 根据标志参数获取备份记录
func getBackupRecords(db *sqlx.DB) ([]types.BackupRecord, error) {
	taskID := logCmdTaskID.Get()     // 任务ID
	taskName := logCmdTaskName.Get() // 任务名称
	limit := logCmdLimit.Get()       // 限制条数

	// 如果指定了任务ID
	if taskID > 0 {
		return DB.GetBackupRecordsByTaskIDWithLimit(db, int64(taskID), limit)
	}

	// 如果指定了任务名称
	if taskName != "" {
		// 先根据任务名称获取任务ID
		id, err := DB.GetTaskIDByName(db, taskName)
		if err != nil {
			return nil, fmt.Errorf("根据任务名称获取任务ID失败: %w", err)
		}
		return DB.GetBackupRecordsByTaskIDWithLimit(db, id, limit)
	}

	// 默认获取所有备份记录
	return DB.GetAllBackupRecordsWithLimit(db, limit)
}
