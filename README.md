# BakCtl - 跨平台备份管理工具

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/golang/go)
[![Gitee](https://img.shields.io/badge/Gitee-Repository-red?style=flat-square&logo=gitee)](https://gitee.com/MM-Q/bakctl)

</div>

一个轻量级的跨平台命令行备份管理工具，提供任务创建、执行、恢复、清理等完整生命周期管理。

## 核心特性

- **任务管理** - 支持交互式和配置文件两种方式创建/编辑/删除备份任务，美观的表格展示
- **智能备份** - 高效 ZIP 压缩、包含/排除过滤规则、实时彩色进度条、SHA-1 完整性校验
- **保留策略** - 按数量/天数/组合策略自动清理过期备份，孤儿记录自动回收
- **一键恢复** - 支持指定版本或最新备份恢复，自动校验文件完整性
- **操作日志** - 完整记录每次备份的详细信息，支持按任务/状态过滤查询
- **导入导出** - 支持导出为 add 命令或一键备份脚本（BAT/Shell）
- **跨平台** - Windows / Linux / macOS 全平台兼容，中文帮助信息开箱即用

## 安装

```bash
git clone https://gitee.com/MM-Q/bakctl.git
cd bakctl
go build -o bakctl .
# 或通过构建脚本一键编译安装
python build.py -s -ai -f
```

要求 Go 1.25.0+。

## 快速开始

```bash
# 创建备份任务
bakctl add --name "文档备份" --backup-dir "/home/user/docs" --storage-dir "/backup/docs"

# 执行备份
bakctl run -id 1

# 查看备份日志
bakctl log -id 1

# 恢复备份
bakctl restore -id 1 --latest -d "/restore/path"

# 列出所有任务
bakctl list

# 删除任务
bakctl delete -id 1
```

详细参数请运行 `bakctl <命令> --help` 查看。

## 许可证

[MIT](LICENSE) - Copyright (c) 2025 M乔木

## 链接

- 仓库: https://gitee.com/MM-Q/bakctl
- 问题反馈: https://gitee.com/MM-Q/bakctl/issues
- 依赖库: [colorlib](https://gitee.com/MM-Q/colorlib) | [comprx](https://gitee.com/MM-Q/comprx) | [qflag](https://gitee.com/MM-Q/qflag) | [verman](https://gitee.com/MM-Q/verman)
