// Package add 实现了 bakctl 的 add 子命令功能。
//
// 该包提供了创建新备份任务的完整功能，支持两种创建方式：
//   - 交互式创建：通过命令行参数直接指定任务配置
//   - 配置文件创建：从 TOML 配置文件中读取任务配置
//
// 主要功能包括：
//   - 解析和验证命令行参数
//   - 读取和解析 TOML 配置文件
//   - 验证备份任务配置的有效性
//   - 将任务配置保存到数据库
//   - 提供用户友好的错误提示和成功反馈
//
// 支持的配置选项包括任务名称、备份目录、存储目录、压缩设置、
// 保留策略、文件过滤规则等。
package add

import (
	"fmt"
	"os"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"gitee.com/MM-Q/colorlib"
	"github.com/jmoiron/sqlx"
	"github.com/pelletier/go-toml/v2"
)

// addCmdMain 添加任务的主函数
//
// 参数:
//   - db: 数据库连接
//   - cl: 颜色库
//
// 返回值:
//   - error: 错误信息
func AddCmdMain(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 如果指定了生成配置文件的选项，则生成配置文件
	if genF.Get() {
		if err := GenerateConfigFile(); err != nil {
			return err
		}

		cl.Green("已生成配置文件:", types.AddTaskFilename)
		return nil
	}

	// 获取配置文件路径
	configPath := configF.Get()

	// 优先使用配置文件，如果没有配置文件则尝试使用命令行标志
	if configPath != "" {
		return addTaskFromConfigFile(db, configPath, cl)
	}

	// 尝试从命令行标志创建任务
	return addTaskFromFlags(db, cl)
}

// addTaskFromConfigFile 从配置文件添加任务
//
// 参数:
//   - db: 数据库连接
//   - configPath: 配置文件路径
//   - cl: 颜色库
//
// 返回值:
//   - error: 错误信息
func addTaskFromConfigFile(db *sqlx.DB, configPath string, cl *colorlib.ColorLib) error {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// 解析toml配置文件
	var config types.RootConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}

	// 检查任务名是否已存在
	taskID, err := DB.GetTaskIDByName(db, config.AddTaskConfig.Name)
	if err == nil && taskID != 0 {
		return fmt.Errorf("任务名称 '%s' 已存在，请使用其他名称", config.AddTaskConfig.Name)
	}

	// 获取最大文件大小
	var maxFileSize int64
	if config.AddTaskConfig.MaxFileSize != "" {
		if err := maxSizeF.Set(config.AddTaskConfig.MaxFileSize); err != nil {
			return fmt.Errorf("无效的最大文件大小: %v", err)
		} else {
			maxFileSize = maxSizeF.Get() // 获取转为字节的最大值
		}
	}

	// 转换最大文件大小
	var minFileSize int64
	if config.AddTaskConfig.MinFileSize != "" {
		if err := minSizeF.Set(config.AddTaskConfig.MinFileSize); err != nil {
			return fmt.Errorf("无效的最小文件大小: %v", err)
		} else {
			minFileSize = minSizeF.Get() // 获取转为字节的最小值
		}
	}

	// 转换为任务配置
	taskConfig := &types.TaskConfig{
		Name:         config.AddTaskConfig.Name,         // 任务名称
		BackupDir:    config.AddTaskConfig.BackupDir,    // 备份目录
		StorageDir:   config.AddTaskConfig.StorageDir,   // 存储目录
		Compress:     config.AddTaskConfig.Compress,     // 是否压缩
		RetainCount:  config.AddTaskConfig.RetainCount,  // 保留数量
		RetainDays:   config.AddTaskConfig.RetainDays,   // 保留天数
		IncludeRules: config.AddTaskConfig.IncludeRules, // 包含规则
		ExcludeRules: config.AddTaskConfig.ExcludeRules, // 排除规则
		MaxFileSize:  maxFileSize,                       // 最大文件大小
		MinFileSize:  minFileSize,                       // 最小文件大小
	}

	// 将配置文件中的内容保存到数据库中
	if err := DB.InsertAddTaskConfig(db, taskConfig); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	cl.Greenf("任务 '%s' 添加成功!", config.AddTaskConfig.Name)

	return nil
}

// addTaskFromFlags 从命令行标志添加任务
//
// 参数:
//   - db: 数据库连接
//   - cl: 颜色库
//
// 返回值:
//   - error: 错误信息
func addTaskFromFlags(db *sqlx.DB, cl *colorlib.ColorLib) error {
	// 构建任务配置
	config := &types.TaskConfig{
		Name:         nameF.Get(),        // 任务名称
		RetainCount:  retainCountF.Get(), // 保留备份数量
		RetainDays:   retainDaysF.Get(),  // 保留天数
		BackupDir:    backupDirF.Get(),   // 备份源目录
		StorageDir:   storageDirF.Get(),  // 存储目录
		Compress:     compressF.Get(),    // 是否压缩
		IncludeRules: includeF.Get(),     // 包含规则
		ExcludeRules: excludeF.Get(),     // 排除规则
		MaxFileSize:  maxSizeF.Get(),     // 最大文件大小
		MinFileSize:  minSizeF.Get(),     // 最小文件大小
	}

	// 检查必须参数
	if err := config.Validate(); err != nil {
		return fmt.Errorf("参数验证失败: %v", err)
	}

	// 检查任务名是否已存在
	taskID, err := DB.GetTaskIDByName(db, config.Name)
	if err == nil && taskID != 0 {
		return fmt.Errorf("任务名称 '%s' 已存在，请使用其他名称", config.Name)
	}

	// 保存到数据库
	if err := DB.InsertAddTaskConfig(db, config); err != nil {
		return fmt.Errorf("保存任务失败: %v", err)
	}

	cl.Greenf("任务 '%s' 添加成功!", config.Name)
	return nil
}

// GenerateConfigFile 生成配置文件
//
// 返回值:
//   - error: 错误信息
func GenerateConfigFile() error {
	// 检查当前目录下是否存在配置文件
	if _, err := os.Stat(types.AddTaskFilename); err == nil {
		return fmt.Errorf("当前目录下已存在 %s, 请勿重复生成", types.AddTaskFilename)
	}

	// 创建配置文件
	f, err := os.Create(types.AddTaskFilename)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// 通过toml包生成配置文件
	data, err := toml.Marshal(types.RootConfig{})
	if err != nil {
		return fmt.Errorf("生成配置文件失败: %v", err)
	}

	// 写入配置文件
	_, err = f.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}
