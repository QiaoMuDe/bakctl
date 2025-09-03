package list

import (
	"fmt"
	"os"
	"path/filepath"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/colorlib"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/jmoiron/sqlx"
)

// listCmdMain 列出所有备份任务
//
// 参数:
//   - db: 数据库连接
//   - cl: 颜色库
//
// 返回:
//   - error: 错误信息
func ListCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 静默清理孤儿记录（处理错误但不输出）
	if _, err := DB.CleanupOrphanRecords(db, 0); err != nil {
		return fmt.Errorf("清理孤儿记录失败: %w", err)
	}

	// 创建表格
	t := table.NewWriter()

	// 设置表格样式
	if style, ok := types.TableStyle[listCmdTableStyle.Get()]; ok {
		t.SetStyle(style)
	} else {
		return fmt.Errorf("表格样式不存在: %s, 可选样式: %v", listCmdTableStyle.Get(), types.TableStyleList)
	}

	// 查询任务列表
	data, err := DB.GetAllTasks(db)
	if err != nil {
		return fmt.Errorf("查询任务列表失败: %w", err)
	}

	// 检查是否有任务
	if len(data) == 0 {
		cl.Yellow("当前没有备份任务")
		cl.Whitef("提示: 使用 '%s add' 命令添加新的备份任务", filepath.Base(os.Args[0]))
		return nil
	}

	// 使用标准输出作为输出目标
	t.SetOutputMirror(os.Stdout)

	// 根据简洁模式设置不同的表头和列配置
	if listCmdSimple.Get() {
		// 简洁模式：只显示ID、任务名、备份源目录、备份存储目录
		t.AppendHeader(table.Row{"ID", "任务名", "备份源目录", "备份存储目录"})

		t.SetColumnConfigs([]table.ColumnConfig{
			{Name: "ID", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
			{Name: "任务名", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "备份源目录", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "备份存储目录", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
		})

		// 添加简洁模式数据行
		for _, task := range data {
			t.AppendRow(table.Row{
				task.ID,         // ID
				task.Name,       // 任务名
				task.BackupDir,  // 备份源目录
				task.StorageDir, // 备份存储目录
			})
		}
	} else {
		// 完整模式：显示所有信息
		t.AppendHeader(table.Row{"ID", "任务名", "保留数量", "保留天数", "备份源目录", "备份存储目录", "是否压缩", "包含规则", "排除规则", "最大文件大小(字节)", "最小文件大小(字节)"})

		t.SetColumnConfigs([]table.ColumnConfig{
			{Name: "ID", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
			{Name: "任务名", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "保留数量", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
			{Name: "保留天数", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
			{Name: "备份源目录", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "备份存储目录", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "是否压缩", Align: text.AlignCenter, WidthMaxEnforcer: text.WrapHard},
			{Name: "包含规则", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "排除规则", Align: text.AlignLeft, WidthMaxEnforcer: text.WrapHard},
			{Name: "最大文件大小(字节)", Align: text.AlignRight, WidthMaxEnforcer: text.WrapHard},
			{Name: "最小文件大小(字节)", Align: text.AlignRight, WidthMaxEnforcer: text.WrapHard},
		})

		// 添加完整模式数据行
		for _, task := range data {
			t.AppendRow(table.Row{
				task.ID,           // ID
				task.Name,         // 任务名
				task.RetainCount,  // 保留数量
				task.RetainDays,   // 保留天数
				task.BackupDir,    // 备份源目录
				task.StorageDir,   // 备份存储目录
				task.Compress,     // 是否压缩
				task.IncludeRules, // 包含规则
				task.ExcludeRules, // 排除规则
				task.MaxFileSize,  // 最大文件大小
				task.MinFileSize,  // 最小文件大小
			})
		}
	}

	// 渲染表格
	t.Render()

	return nil
}
