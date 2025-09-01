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

func LogCmdMain(db *sqlx.DB) error {
	// 创建表格
	t := table.NewWriter()

	// 设置表格样式
	if style, ok := types.TableStyle[logCmdTableStyle.Get()]; ok {
		t.SetStyle(style)
	} else {
		return fmt.Errorf("表格样式不存在: %s, 可选样式: %v", logCmdTableStyle.Get(), types.TableStyleList)
	}

	// 查询备份记录列表
	data, err := DB.GetAllBackupRecords(db)
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
