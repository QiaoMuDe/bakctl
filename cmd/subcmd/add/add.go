package add

import (
	"fmt"
	"os"

	DB "gitee.com/MM-Q/bakctl/internal/db"
	"gitee.com/MM-Q/bakctl/internal/types"
	"github.com/jmoiron/sqlx"
	"github.com/pelletier/go-toml/v2"
)

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

	// 如果配置文件路径为空，则返回nil
	if configPath == "" {
		return nil
	}

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

	// 将配置文件中的内容保存到数据库中
	if err := DB.InsertAddTaskConfig(db, &config.AddTask); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}
	fmt.Println("添加任务成功!")

	return nil
}

// 生成配置文件
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
