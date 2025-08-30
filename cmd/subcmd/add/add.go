package add

import (
	"fmt"
	"os"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
	"github.com/pelletier/go-toml/v2"
)

// addCmdMain 添加任务的主函数
func AddCmdMain(db *sqlx.DB) error {
	// 如果指定了生成配置文件的选项，则生成配置文件
	if genF.Get() {
		if err := GenerateConfigFile(); err != nil {
			return err
		}

		fmt.Println("已生成配置文件:", types.AddTaskFilename)
		return nil
	}

	// 获取配置文件路径
	configPath := configF.Get()

	// 优先使用配置文件，如果没有配置文件则尝试使用命令行标志
	if configPath != "" {
		return addTaskFromConfigFile(db, configPath)
	}

	// 尝试从命令行标志创建任务
	return addTaskFromFlags(db)
}

// addTaskFromConfigFile 从配置文件添加任务
//
// 参数:
//   - db: 数据库连接
//   - configPath: 配置文件路径
//
// 返回值:
//   - error: 错误信息
func addTaskFromConfigFile(db *sqlx.DB, configPath string) error {
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

	// 检查必须参数
	if err := config.AddTaskConfig.Validate(); err != nil {
		return fmt.Errorf("配置文件验证失败: %v", err)
	}

	// 检查任务名是否已存在
	taskID, err := DB.GetTaskIDByName(db, config.AddTaskConfig.Name)
	if err == nil && taskID != 0 {
		return fmt.Errorf("任务名称 '%s' 已存在，请使用其他名称", config.AddTaskConfig.Name)
	}

	// 将配置文件中的内容保存到数据库中
	if err := DB.InsertAddTaskConfig(db, &config.AddTaskConfig); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}
	fmt.Println("添加任务成功!")

	return nil
}

// addTaskFromFlags 从命令行标志添加任务
//
// 参数:
//   - db: 数据库连接
//
// 返回值:
//   - error: 错误信息
func addTaskFromFlags(db *sqlx.DB) error {
	// 构建任务配置
	config := &types.AddTaskConfig{
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

	fmt.Println("添加任务成功!")
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
