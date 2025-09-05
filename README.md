# 🗂️ BakCtl - 跨平台备份管理工具

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge)](https://github.com/golang/go)
[![Gitee](https://img.shields.io/badge/Gitee-Repository-red?style=for-the-badge&logo=gitee)](https://gitee.com/MM-Q/bakctl)

**一个功能强大、易于使用的跨平台备份管理工具**

[🚀 快速开始](#-安装指南) • [📖 使用文档](#-使用示例) • [🔧 配置说明](#-配置选项) • [🤝 贡献指南](#-贡献指南)

</div>

---

## 📋 项目简介

BakCtl 是一个现代化的备份管理工具，专为简化文件和目录的备份操作而设计。它提供了完整的备份生命周期管理，包括任务创建、执行、监控、恢复和清理功能。

### 🎯 设计理念

- **简单易用**：直观的命令行界面，支持中文帮助信息
- **功能完整**：涵盖备份管理的全生命周期
- **高度可配置**：灵活的过滤规则和保留策略
- **跨平台支持**：Windows、Linux、macOS 全平台兼容
- **数据安全**：SQLite 数据库存储，支持完整性校验

## ✨ 核心特性

### 🔧 任务管理
- ➕ **任务创建**：支持交互式和配置文件两种方式创建备份任务
- ✏️ **任务编辑**：灵活修改现有备份任务的各项配置
- 📋 **任务列表**：美观的表格展示所有备份任务信息
- 🗑️ **任务删除**：安全删除备份任务及相关数据

### 🚀 备份执行
- ⚡ **快速备份**：高效的文件压缩和存储机制
- 🎯 **智能过滤**：支持包含/排除规则，精确控制备份内容
- 📊 **实时进度**：彩色进度条显示备份进度
- 🔒 **完整性校验**：自动生成和验证文件哈希值

### 📈 监控与日志
- 📝 **详细日志**：完整记录每次备份操作的详细信息
- 📊 **状态监控**：实时查看备份任务的执行状态
- 🔍 **历史查询**：支持按任务、时间等条件查询备份历史

### 🔄 恢复与清理
- 🔄 **一键恢复**：快速恢复指定版本的备份文件
- 🧹 **自动清理**：基于保留策略自动清理过期备份
- 🗂️ **孤儿清理**：自动清理数据库中的无效记录

### 📤 导入导出
- 📤 **配置导出**：支持导出备份任务配置
- 📥 **批量导入**：支持从配置文件批量创建任务

## 🛠️ 安装指南

### 📦 从源码安装

```bash
# 克隆仓库
git clone https://gitee.com/MM-Q/bakctl.git
cd bakctl

# 编译安装
go build -o bakctl cmd/bakctl/main.go

# 通过build.py一键编译安装
python build.py -s -ai -f
```

### 🔧 系统要求

- **Go 版本**：1.25.0 或更高版本
- **操作系统**：Windows 10+、Linux、macOS
- **磁盘空间**：至少 8MB 可用空间

## 📚 使用示例

### 🚀 基础用法

```bash
# 查看帮助信息
bakctl --help

# 创建新的备份任务
bakctl add --name "我的文档备份" --backup-dir "/home/user/documents" --storage-dir "/backup/docs"

# 列出所有备份任务
bakctl list

# 执行指定任务的备份
bakctl run -id 1

# 查看备份日志
bakctl log -id 1
```

### 🔧 高级用法

```bash
# 创建带过滤规则的备份任务
bakctl add \
  --name "项目备份" \
  --backup-dir "/home/user/projects" \
  --storage-dir "/backup/projects" \
  --include "*.go,*.md,*.json" \
  --exclude "node_modules,*.log,*.tmp" \
  --compress \
  --retain-count 10 \
  --retain-days 30

# 批量执行多个任务
bakctl run -ids 1,2,3

# 执行所有任务
bakctl run --all

# 恢复指定版本的备份
bakctl restore -id 1 -vid "abc123" -d "/restore/path"

# 删除任务及其所有备份数据
bakctl delete -id 1 
```

### 📋 配置文件示例

创建 `backup_config.toml` 文件：

```toml
[task]
name = "重要文档备份"
backup_dir = "/home/user/important_docs"
storage_dir = "/backup/important"
compress = true
retain_count = 5
retain_days = 30
max_file_size = '100MB'  
min_file_size = '10KB'

[rules]
include = ["*.pdf", "*.docx", "*.xlsx", "*.txt"]
exclude = ["*.tmp", "*.log", "~*"]
```

然后使用配置文件创建任务：

```bash
bakctl add --config backup_config.toml
```

## 📖 API 文档概述

### 🎯 命令结构

BakCtl 采用子命令架构，每个功能模块对应一个子命令：

| 命令 | 简写 | 功能描述 |
|------|------|----------|
| `add` | `a` | 创建新的备份任务 |
| `edit` | `e` | 编辑现有备份任务 |
| `list` | `l` | 列出所有备份任务 |
| `run` | `r` | 执行备份任务 |
| `log` | `lg` | 查看备份日志 |
| `restore` | `rs` | 恢复备份文件 |
| `delete` | `d` | 删除备份任务 |
| `export` | `ex` | 导出任务配置 |

### 🔧 全局选项

| 选项 | 简写 | 描述 |
|------|------|------|
| `--no-color` | `-nc` | 禁用彩色输出 |
| `--help` | `-h` | 显示帮助信息 |
| `--version` | `-v` | 显示版本信息 |

## 🎛️ 支持的功能特性

### 📁 文件格式支持
- ✅ **压缩格式**：ZIP（默认）
- ✅ **文件类型**：支持所有文件类型
- ✅ **大文件**：支持大文件备份（>4GB）
- ✅ **符号链接**：智能处理符号链接

### 🔍 过滤规则
- 🎯 **包含规则**：支持通配符模式匹配
- 🚫 **排除规则**：灵活的排除模式
- 📏 **文件大小**：基于文件大小的过滤
- 📅 **时间过滤**：基于修改时间的过滤

### 🗄️ 存储特性
- 💾 **本地存储**：支持本地文件系统存储
- 🔐 **数据完整性**：SHA-256 哈希校验
- 📊 **元数据管理**：SQLite 数据库存储任务信息
- 🧹 **自动清理**：智能的备份保留策略

## ⚙️ 配置选项

### 📋 任务配置参数

| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `name` | string | ✅ | - | 备份任务名称 |
| `backup_dir` | string | ✅ | - | 源目录路径 |
| `storage_dir` | string | ✅ | `~/.bakctl/bak` | 备份存储目录 |
| `compress` | bool | ❌ | `false` | 是否启用压缩 |
| `retain_count` | int | ❌ | `0` | 保留备份数量（0=无限制） |
| `retain_days` | int | ❌ | `0` | 保留天数（0=无限制） |
| `max_file_size` | string | ❌ | `0` | 最大文件大小 |
| `min_file_size` | string | ❌ | `0` | 最小文件大小 |

### 🎯 过滤规则配置

```toml
[rules]
# 包含规则 - 只备份匹配的文件
include = [
    "*.go",           # Go 源文件
    "*.md",           # Markdown 文件
    "*.json",         # JSON 配置文件
    "docs/**/*.pdf"   # docs 目录下的所有 PDF 文件
]

# 排除规则 - 跳过匹配的文件/目录
exclude = [
    "node_modules",   # Node.js 依赖目录
    "*.log",          # 日志文件
    "*.tmp",          # 临时文件
    ".git",           # Git 仓库目录
    "**/.DS_Store"    # macOS 系统文件
]
```

### 🗂️ 目录结构配置

```
~/.bakctl/                    # 默认数据目录
├── bakctl.db3               # SQLite 数据库文件
├── bak/                     # 默认备份存储目录
│   ├── task1_20250903_143022.zip
│   └── task2_20250903_143045.zip
└── config/                  # 配置文件目录（可选）
    └── tasks.toml
```

## 🏗️ 项目结构

```
bakctl/
├── cmd/                     # 命令行入口
│   ├── bakctl/             # 主程序
│   │   └── main.go
│   └── subcmd/             # 子命令实现
│       ├── add/            # 添加任务命令
│       ├── delete/         # 删除任务命令
│       ├── edit/           # 编辑任务命令
│       ├── export/         # 导出配置命令
│       ├── list/           # 列表显示命令
│       ├── log/            # 日志查看命令
│       ├── restore/        # 恢复备份命令
│       └── run/            # 执行备份命令
├── internal/               # 内部包
│   ├── cleanup/            # 清理功能
│   ├── db/                 # 数据库操作
│   ├── types/              # 类型定义
│   └── utils/              # 工具函数
├── gobf/                   # 构建脚本
├── go.mod                  # Go 模块文件
├── go.sum                  # 依赖校验文件
├── LICENSE                 # MIT 许可证
└── README.md               # 项目文档
```

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。

```
MIT License

Copyright (c) 2025 M乔木

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

## 🤝 贡献指南

我们欢迎所有形式的贡献！请遵循以下步骤：

### 🔧 开发环境设置

1. **Fork 项目**到你的 Gitee 账户
2. **克隆**你的 fork：
   ```bash
   git clone https://gitee.com/your-username/bakctl.git
   ```
3. **创建特性分支**：
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. **提交更改**：
   ```bash
   git commit -m 'Add some amazing feature'
   ```
5. **推送到分支**：
   ```bash
   git push origin feature/amazing-feature
   ```
6. **创建 Pull Request**

### 📝 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档
- 编写单元测试
- 确保所有测试通过

### 🐛 问题报告

如果你发现了 bug 或有功能建议，请：

1. 检查是否已有相关 issue
2. 创建新的 issue，详细描述问题
3. 提供复现步骤和环境信息
4. 如果可能，提供修复建议

## 📞 联系方式

- **项目仓库**：[https://gitee.com/MM-Q/bakctl](https://gitee.com/MM-Q/bakctl)
- **作者**：M乔木
- **问题反馈**：[提交 Issue](https://gitee.com/MM-Q/bakctl/issues)

## 🔗 相关链接

- 📚 **Go 官方文档**：[https://golang.org/doc/](https://golang.org/doc/)
- 🛠️ **依赖库**：
  - [colorlib](https://gitee.com/MM-Q/colorlib) - 彩色输出库
  - [comprx](https://gitee.com/MM-Q/comprx) - 压缩处理库
  - [qflag](https://gitee.com/MM-Q/qflag) - 命令行参数解析
  - [verman](https://gitee.com/MM-Q/verman) - 版本管理库

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给它一个 Star！**

[![Gitee Stars](https://gitee.com/MM-Q/bakctl/badge/star.svg?theme=dark)](https://gitee.com/MM-Q/bakctl)

</div>