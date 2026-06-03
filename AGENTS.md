# BakCtl 项目分析报告

> **项目名称**: BakCtl - 跨平台备份管理工具
> **作者**: M乔木
> **许可证**: MIT (Copyright 2025)
> **仓库地址**: https://gitee.com/MM-Q/bakctl
> **分析日期**: 2026-06-03

---

## 一、项目概述

BakCtl 是一个使用 Go 语言开发的跨平台命令行备份管理工具，提供备份任务的全生命周期管理能力，包括任务创建、执行、监控、恢复和清理。项目采用子命令架构设计，使用 SQLite 数据库存储元数据，支持文件压缩、过滤规则、保留策略、完整性校验等功能。

**核心业务场景**: 本地文件/目录的备份管理，适用于开发者和运维人员的日常文件备份需求。

---

## 二、目录结构梳理

```
bakctl/
├── main.go                         # 程序入口，直接调用 bakctl.BakctlMain()
├── go.mod                          # Go模块定义 (module gitee.com/MM-Q/bakctl, go 1.25.0)
├── go.sum                          # 依赖校验文件
├── build.py                        # Python构建脚本（支持多平台编译、Git信息注入、ZIP打包）
├── .gitignore                      # Git忽略规则（二进制、测试产物等）
├── LICENSE                         # MIT许可证
├── README.md                       # 项目文档（结构完整、示例丰富）
│
├── cmd/                            # 命令行层（路由+子命令实现）
│   ├── bakctl/
│   │   └── bakctl.go               # 主命令路由中心：初始化所有子命令、解析参数、路由执行
│   └── subcmd/                     # 8个子命令实现（每个子命令一个独立包）
│       ├── add/                    # add命令：添加备份任务
│       │   ├── add.go              # 核心逻辑（支持命令行参数和TOML配置文件两种创建方式）
│       │   └── flags.go           # 命令行参数定义
│       ├── delete/                 # delete命令：删除备份任务
│       │   ├── delete.go           # 核心逻辑（支持单任务/批量/失败记录删除）
│       │   └── flags.go           # 命令行参数定义
│       ├── edit/                   # edit命令：编辑备份任务配置
│       │   ├── edit.go             # 核心逻辑（支持增量修改、批量编辑）
│       │   └── flags.go           # 命令行参数定义
│       ├── export/                 # export命令：导出任务配置
│       │   ├── export.go           # 核心逻辑（支持导出为命令/脚本，BAT/Shell格式）
│       │   └── flags.go           # 命令行参数定义
│       ├── list/                   # list命令：列出所有任务
│       │   ├── list.go             # 核心逻辑（表格渲染，支持简洁/完整模式）
│       │   └── flags.go           # 命令行参数定义
│       ├── log/                    # log命令：查看备份日志
│       │   ├── log.go              # 核心逻辑（支持过滤、分页、多种表格样式）
│       │   └── flags.go           # 命令行参数定义
│       ├── restore/                # restore命令：恢复备份
│       │   ├── restore.go          # 核心逻辑（校验+解压恢复）
│       │   └── flags.go           # 命令行参数定义
│       └── run/                    # run命令：执行备份任务
│           ├── run.go              # 核心逻辑（完整备份流程：验证→压缩→校验→记录→清理）
│           └── flags.go           # 命令行参数定义
│
├── internal/                       # 内部包（不对外暴露）
│   ├── cleanup/                    # 备份文件清理模块
│   │   ├── cleanup.go              # 清理核心算法（按数量/天数/组合策略）
│   │   ├── adapter.go              # 适配器模式实现（BackupTaskAdapter）
│   │   └── integration.go         # 集成层（BackupTask接口定义、带日志的清理封装）
│   ├── db/                         # 数据库操作模块
│   │   ├── db.go                   # 数据库初始化、建表脚本、INSERT/UPDATE操作
│   │   ├── get.go                  # 查询操作（单条/批量/过滤/分页）
│   │   ├── delete.go               # 删除操作（单条/批量/级联）
│   │   └── cleanup.go             # 孤儿记录清理（检测文件不存在的记录并删除）
│   ├── types/                      # 类型定义模块
│   │   ├── types.go                # 核心数据类型（BackupTask/BackupRecord/BackupResult/表格样式）
│   │   └── config.go              # 配置类型（AddTaskConfig/TaskConfig/RootConfig/验证逻辑）
│   └── utils/                      # 工具函数模块
│       ├── utils.go                # 通用工具（GetUserHomeDir/ConvertUTCToLocal）
│       ├── format.go              # 字节格式化（高性能位运算版本）
│       ├── hash.go                # 哈希计算（支持普通/进度条两种模式，动态缓冲区）
│       └── json_helpers.go        # JSON编解码辅助（规则字符串序列化/反序列化）
│
├── gobf/                           # 构建配置目录（gob构建工具）
│   ├── dev.toml                    # 开发环境配置（不注入Git信息、不安装）
│   ├── install.toml               # 安装环境配置（注入Git信息、自动安装到GOPATH/bin）
│   └── release.toml               # 发布环境配置（批量编译、ZIP打包、全平台）
│
└── vendor/                         # Go vendor依赖目录
    ├── gitee.com/MM-Q/            # 作者自研库
    │   ├── colorlib/              # 彩色输出库
    │   ├── comprx/                # 压缩处理库
    │   ├── go-kit/                # 工具集（hash/id/pool/str/fuzzy等）
    │   ├── qflag/                 # 命令行参数解析库
    │   └── verman/                # 版本管理库
    └── github.com/                # 第三方库
        ├── jmoiron/sqlx/          # SQL扩展库
        ├── jedib0t/go-pretty/     # 表格渲染库
        ├── pelletier/go-toml/     # TOML解析库
        ├── schollz/progressbar/   # 进度条库
        └── ...                    # 其他间接依赖
```

### 规范度评估

| 维度 | 评价 | 说明 |
|------|------|------|
| 目录结构 | **优秀** | 遵循 Go 项目标准布局（cmd/internal/vendor），职责清晰 |
| 文件命名 | **优秀** | 小写下划线命名，与功能一一对应 |
| 包粒度 | **优秀** | 每个子命令独立包，internal下按功能模块拆分 |
| 冗余目录 | **无** | 无多余/不合理目录 |
| 冗余文件 | **无** | 所有文件均有明确用途 |

---

## 三、核心功能模块识别

### 3.1 业务核心模块

| 模块名称 | 核心功能 | 对应代码文件 | 核心输入/输出 | 核心依赖 |
|----------|---------|-------------|-------------|---------|
| **任务管理模块** | CRUD操作备份任务 | `cmd/subcmd/add/`, `cmd/subcmd/edit/`, `cmd/subcmd/delete/`, `cmd/subcmd/list/` | 输入：命令行参数/TOML配置；输出：数据库记录+终端表格 | db模块, types模块 |
| **备份执行模块** | 执行文件压缩备份 | `cmd/subcmd/run/run.go` | 输入：任务ID；输出：ZIP备份文件+数据库记录 | comprx, hash, cleanup, db |
| **恢复模块** | 从备份文件恢复数据 | `cmd/subcmd/restore/restore.go` | 输入：任务ID+版本ID；输出：解压后的文件 | comprx, hash, db |
| **日志查看模块** | 展示备份历史记录 | `cmd/subcmd/log/log.go` | 输入：过滤条件；输出：格式化表格 | go-pretty, db |
| **导出模块** | 导出任务配置/脚本 | `cmd/subcmd/export/export.go` | 输入：任务ID列表；输出：命令/脚本文本 | db |
| **清理模块** | 按策略清理过期备份 | `internal/cleanup/cleanup.go` | 输入：存储目录+保留策略；输出：删除结果 | 文件系统操作 |

### 3.2 基础支撑模块

| 模块名称 | 核心功能 | 对应代码文件 | 说明 |
|----------|---------|-------------|------|
| **命令路由模块** | 初始化命令、解析参数、路由到子命令 | `cmd/bakctl/bakctl.go` | 程序总入口，包含panic恢复 |
| **数据库模块** | SQLite连接管理、表结构初始化、CRUD操作 | `internal/db/` | 4个文件，按操作类型拆分 |
| **类型定义模块** | 核心数据结构、常量定义、验证逻辑 | `internal/types/` | 2个文件：业务类型+配置类型 |
| **工具函数模块** | 格式化、哈希、JSON、路径工具 | `internal/utils/` | 4个文件，按功能拆分 |
| **命令行参数模块** | 参数解析和验证 | 各子命令的 `flags.go` | 基于qflag库封装 |
| **构建配置模块** | 多环境构建配置 | `gobf/*.toml` | dev/install/release三套配置 |

---

## 四、模块间依赖关系分析

### 4.1 依赖关系图（文字描述）

```
层级依赖关系（自上而下）：

┌─────────────────────────────────────────────────────┐
│                    main.go                           │
│                   (程序入口)                          │
└───────────────────────┬─────────────────────────────┘
                        │ 调用
┌───────────────────────▼─────────────────────────────┐
│              cmd/bakctl/bakctl.go                    │
│           (主命令路由中心)                             │
│  初始化8个子命令 → 解析参数 → 路由执行                  │
└───┬───┬───┬───┬───┬───┬───┬───┬────────────────────┘
    │   │   │   │   │   │   │   │  注册并调用
┌───▼───▼───▼───▼───▼───▼───▼───▼────────────────────┐
│              cmd/subcmd/* (8个子命令)                 │
│  add | edit | delete | list | run | log | restore | export │
└───┬───┬───┬───┬───┬───┬───┬────────────────────────┘
    │   │   │   │   │   │   │  调用
┌───▼───▼───▼───▼───▼───▼───────────────────────────┐
│              internal/* (内部模块)                    │
│  ┌─────┐ ┌──────┐ ┌───────┐ ┌──────┐              │
│  │ db  │ │types │ │cleanup│ │utils │              │
│  └──┬──┘ └──┬───┘ └──┬────┘ └──────┘              │
└─────┼───────┼────────┼─────────────────────────────┘
      │       │        │
┌─────▼───────▼────────▼─────────────────────────────┐
│              第三方/自研依赖库                         │
│  comprx | colorlib | qflag | verman | go-kit        │
│  go-pretty | sqlx | go-toml | progressbar | sqlite  │
└─────────────────────────────────────────────────────┘
```

### 4.2 详细依赖关系

| 依赖方 | 被依赖方 | 依赖内容 |
|--------|---------|---------|
| `cmd/bakctl` | `cmd/subcmd/*` (8个) | 调用各子命令的 `Init*Cmd()` 和 `*CmdMain()` |
| `cmd/bakctl` | `internal/db` | 调用 `db.InitSQLite()` 初始化数据库 |
| `cmd/bakctl` | `internal/types` | 引用 `types.DBFilename`, `types.DataDirPath` |
| `cmd/bakctl` | `qflag`, `colorlib`, `verman` | 参数解析、彩色输出、版本信息 |
| `cmd/subcmd/add` | `internal/db`, `internal/types` | 插入任务、类型定义 |
| `cmd/subcmd/run` | `internal/db`, `internal/cleanup`, `internal/types`, `internal/utils` | 查询任务、清理备份、类型定义、JSON解析 |
| `cmd/subcmd/run` | `comprx`, `go-kit/hash`, `go-kit/id` | 压缩、哈希计算、ID生成 |
| `cmd/subcmd/delete` | `internal/db`, `internal/types` | 查询/删除记录、类型定义 |
| `cmd/subcmd/edit` | `internal/db`, `internal/types`, `internal/utils` | 查询/更新任务、类型定义、JSON编解码 |
| `cmd/subcmd/list` | `internal/db`, `internal/types`, `internal/utils` | 查询任务、格式化字节 |
| `cmd/subcmd/log` | `internal/db`, `internal/types`, `internal/utils` | 查询记录、格式化、时间转换 |
| `cmd/subcmd/restore` | `internal/db`, `internal/types` | 查询记录、类型定义 |
| `cmd/subcmd/restore` | `comprx`, `go-kit/hash` | 解压、哈希校验 |
| `cmd/subcmd/export` | `internal/db`, `internal/types`, `internal/utils` | 查询任务、类型定义、JSON解析 |
| `internal/db` | `internal/types`, `internal/utils` | 类型定义、JSON编解码 |
| `internal/db` | `sqlx`, `modernc.org/sqlite` | 数据库操作、SQLite驱动 |
| `internal/cleanup` | `colorlib` | 彩色日志输出 |
| `internal/types` | `internal/utils` | 获取用户目录 |
| `internal/types` | `go-pretty/table` | 表格样式定义 |

### 4.3 潜在依赖问题

| 问题类型 | 说明 | 严重程度 |
|----------|------|---------|
| **无循环依赖** | cleanup包通过`BackupTask`接口解耦，避免了与types包的循环依赖 | - |
| **无过度依赖** | 各子命令仅依赖必要的内部模块，耦合度合理 | - |
| **export模块未使用colorlib** | `ExportCmdMain`签名中未传入colorlib，与其他子命令风格不一致 | 低 |

---

## 五、设计模式与实现逻辑

### 5.1 设计模式识别

| 模式名称 | 应用位置 | 应用场景 | 说明 |
|----------|---------|---------|------|
| **子命令模式 (Subcommand)** | `cmd/bakctl/bakctl.go` L116-L176 | 命令路由 | 经典CLI设计模式，每个子命令独立包，通过switch-case路由 |
| **适配器模式 (Adapter)** | `internal/cleanup/adapter.go` | 解耦cleanup与types | `BackupTaskAdapter`将`types.BackupTask`适配为`cleanup.BackupTask`接口 |
| **策略模式 (Strategy)** | `internal/cleanup/cleanup.go` L185-L236 | 保留策略选择 | 三种策略：按数量、按天数、组合策略，通过条件分支选择 |
| **工厂函数 (Factory)** | `internal/cleanup/adapter.go` L60 | 创建适配器 | `NewBackupTaskAdapter()` 封装适配器创建 |
| **适配器模式 (Interface)** | `internal/cleanup/integration.go` L41-L47 | 定义清理接口 | `BackupTask`接口统一任务信息访问方式 |
| **模板方法 (Template)** | `cmd/subcmd/run/run.go` L81-L167 | 备份执行流程 | 固定步骤：验证→解析→压缩→校验→记录→清理 |
| **延迟执行 (defer)** | `cmd/subcmd/run/run.go` L89-L94 | 结果记录 | 使用defer确保无论成功失败都记录到数据库 |

### 5.2 核心业务流程

#### 5.2.1 备份执行流程 (`run` 命令)

```
用户输入 `bakctl run -id 1`
        │
        ▼
┌─ validateFlags() ─────────────┐  参数互斥检查（-id/-ids/-all三选一）
└──────────────┬────────────────┘
               ▼
┌─ selectTasks(db) ─────────────┐  根据参数从数据库查询任务
│  ├─ 单ID → GetTaskByID()      │
│  ├─ 多ID → GetTasksByIDs()    │
│  └─ 全部 → GetAllTasks()      │
└──────────────┬────────────────┘
               ▼
┌─ executeTasks(tasks, db, cl) ─┐  遍历任务列表，逐个执行
└──────────────┬────────────────┘
               ▼
┌─ executeTask(task, db, cl) ───┐  单个任务执行（核心）
│  │                             │
│  ├─ defer: recordBackupResult()│  确保记录到数据库
│  ├─ 1. validateSourceDir()     │  验证源目录存在
│  ├─ 2. parseFilterRules()      │  解析包含/排除规则（JSON→[]string）
│  ├─ 3. 构建FilterOptions       │  配置过滤器
│  ├─ 4. 设置压缩等级            │
│  ├─ 5. comprx.PackOptions()    │  执行ZIP压缩（带进度条）
│  ├─ 6. collectBackupInfo()     │  获取文件大小 + SHA1哈希
│  ├─ 7. 设置result为成功        │
│  ├─ 8. cleanup.CleanupBackupFilesWithLogging() │  清理历史备份
│  └─ 9. db.CleanupOrphanRecords() │  清理孤儿记录
└────────────────────────────────┘
```

#### 5.2.2 备份清理策略 (`cleanup` 模块)

```
CleanupBackupFiles(storageDir, taskName, retainCount, retainDays)
    │
    ├─ 两者都为0 → 跳过清理
    │
    ├─ collectBackupFiles()        正则匹配备份文件名，提取时间戳
    ├─ sort.Slice()                按时间降序排序
    └─ determineFilesToDelete()    根据策略确定删除列表
        │
        ├─ 仅retainCount > 0      保留最新N个，删除其余
        ├─ 仅retainDays > 0       删除超过N天的文件
        └─ 两者都 > 0             先按天数过滤 → 再每天保留N个
```

#### 5.2.3 任务添加流程 (`add` 命令)

```
用户输入 bakctl add
    │
    ├─ -g/--generate-template     生成TOML模板文件
    ├─ -C/--config <path>         从TOML配置文件读取
    │   ├─ toml.Unmarshal()       解析配置
    │   ├─ 验证必填字段            name, backup_dir
    │   ├─ 路径标准化              相对→绝对路径
    │   ├─ 检查任务名唯一性        GetTaskIDByName()
    │   └─ InsertAddTaskConfig()   插入数据库
    │
    └─ 无配置文件                   从命令行标志构建
        ├─ config.Validate()       验证必填字段
        ├─ 路径标准化
        ├─ 检查任务名唯一性
        └─ InsertAddTaskConfig()   插入数据库
```

### 5.3 代码逻辑评估

| 维度 | 评价 | 说明 |
|------|------|------|
| 流程清晰度 | **优秀** | 每个子命令主函数流程清晰，步骤注释明确 |
| 代码冗余 | **低** | 存在少量重复的参数验证模式（但可接受） |
| 硬编码 | **极少** | 关键常量已提取到types包（DBFilename, HashAlgorithm等） |
| 错误处理 | **完善** | 使用`fmt.Errorf` + `%w`包装错误，调用链完整 |

---

## 六、技术栈评估

### 6.1 核心技术栈

| 技术组件 | 版本 | 用途 | 选型评价 |
|----------|------|------|---------|
| **Go** | 1.25.0+ | 主要开发语言 | 最新版，非常新 |
| **modernc.org/sqlite** | v1.51.0 | 纯Go SQLite驱动 | 无需CGO，跨平台编译友好 |
| **jmoiron/sqlx** | v1.4.0 | SQL扩展库 | 简化SQLite操作，社区成熟 |
| **pelletier/go-toml/v2** | v2.3.1 | TOML配置解析 | 配置文件解析，性能优秀 |
| **jedib0t/go-pretty/v6** | v6.7.10 | 表格渲染 | 功能丰富，样式多样 |
| **schollz/progressbar/v3** | v3.19.0 | 进度条 | 终端进度展示 |
| **gitee.com/MM-Q/qflag** | v0.5.20 | CLI参数解析 | 作者自研，支持中文帮助、自动补全 |
| **gitee.com/MM-Q/comprx** | v0.1.9 | 压缩处理 | 作者自研，封装ZIP操作 |
| **gitee.com/MM-Q/colorlib** | v1.3.2 | 彩色输出 | 作者自研，终端彩色渲染 |
| **gitee.com/MM-Q/go-kit** | v0.0.23 | 工具集 | 作者自研，提供hash/id等 |
| **gitee.com/MM-Q/verman** | v0.0.19 | 版本管理 | 作者自研，Git信息注入 |
| **Python** (build.py) | - | 构建脚本 | 多平台编译、Git信息注入 |

### 6.2 技术栈评估

| 维度 | 评价 |
|------|------|
| **适配性** | **优秀** - 技术栈与项目场景高度匹配。CLI工具选择Go非常合适；SQLite作为嵌入式数据库免运维；pure Go的sqlite驱动确保跨平台编译无CGO依赖 |
| **复杂度** | **适中** - 没有过度使用复杂框架，核心依赖控制在合理范围 |
| **自研库** | **合理** - 作者的qflag/comprx/colorlib/go-kit/verman均为轻量级库，解决特定问题，与项目高度适配 |
| **社区活跃度** | **良好** - 主要依赖库均为活跃维护的开源项目 |
| **版本兼容** | **无风险** - 所有依赖均为较新版本，Go 1.25.0为最新稳定版 |

### 6.3 依赖关系图（直接依赖）

```
bakctl
├── gitee.com/MM-Q/qflag         (CLI解析)
├── gitee.com/MM-Q/colorlib      (彩色输出)
├── gitee.com/MM-Q/verman        (版本管理)
├── gitee.com/MM-Q/comprx        (压缩处理)
├── gitee.com/MM-Q/go-kit        (工具集)
├── github.com/jmoiron/sqlx      (SQL操作)
├── modernc.org/sqlite           (SQLite驱动)
├── github.com/pelletier/go-toml/v2  (TOML解析)
├── github.com/jedib0t/go-pretty/v6  (表格渲染)
└── github.com/schollz/progressbar/v3 (进度条)
```

---

## 七、补充分析项

### 7.1 代码规范

| 维度 | 评价 | 具体表现 |
|------|------|---------|
| **命名规范** | **优秀** | 严格遵循Go官方命名规范，包名小写无下划线，导出函数PascalCase |
| **注释规范** | **优秀** | 每个包有完整的包注释，每个导出函数有参数/返回值说明，关键常量有行内注释 |
| **代码风格** | **优秀** | 统一使用gofmt格式化，错误处理一致（fmt.Errorf + %w） |
| **文件组织** | **优秀** | flags.go与逻辑.go分离，职责清晰 |

### 7.2 异常处理

| 场景 | 处理方式 | 评价 |
|------|---------|------|
| 程序级panic | `bakctl.go` L48-L53 defer recover捕获 | 完善 |
| 数据库操作错误 | 每个DB函数返回error，调用方用fmt.Errorf包装 | 完善 |
| 文件操作错误 | 检查os.Stat/os.IsNotExist，提供明确错误信息 | 完善 |
| 用户输入验证 | 每个子命令有validateFlags()函数 | 完善 |
| 备份执行失败 | defer记录失败结果到数据库 | 完善 |
| 清理操作失败 | 不影响主流程，仅记录错误 | 合理 |

### 7.3 扩展性评估

| 维度 | 评价 | 说明 |
|------|------|------|
| **新子命令** | **易扩展** | 只需在`cmd/subcmd/`下新建包，实现Init*Cmd()和*CmdMain()，在bakctl.go中注册 |
| **新数据库** | **需重构** | 数据库层通过sqlx抽象，但SQLite DDL脚本硬编码，迁移需要修改db.go |
| **新清理策略** | **易扩展** | cleanup包基于策略模式，新增策略只需在determineFilesToDelete中添加分支 |
| **新文件格式** | **易扩展** | 压缩通过comprx库封装，支持新格式只需修改comprx配置 |

### 7.4 性能关键点

| 关注点 | 位置 | 分析 |
|--------|------|------|
| **哈希计算缓冲区** | `internal/utils/hash.go` L166-L184 | 已优化：根据文件大小动态选择32KB-4MB缓冲区 |
| **字节格式化** | `internal/utils/format.go` | 已优化：使用位运算、预定义常量，避免循环 |
| **批量查询** | `internal/db/get.go` | 使用`sqlx.In`展开IN查询，避免N+1问题 |
| **孤儿记录清理** | `internal/db/cleanup.go` | 每次备份执行后全量扫描，任务量大时可能有性能影响（待确认） |
| **备份文件正则匹配** | `internal/cleanup/cleanup.go` | 每次清理重新编译正则，可预编译优化（影响小） |

---

## 八、数据库设计

### 8.1 表结构

**backup_tasks 表**（备份任务配置）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ID | INTEGER | PRIMARY KEY AUTOINCREMENT | 任务唯一标识 |
| name | TEXT | NOT NULL UNIQUE | 任务名称 |
| retain_count | INTEGER | DEFAULT 3 | 保留备份数量 |
| retain_days | INTEGER | DEFAULT 7 | 保留天数 |
| backup_dir | TEXT | NOT NULL | 备份源目录 |
| storage_dir | TEXT | NOT NULL | 存储目录 |
| compress | BOOLEAN | DEFAULT FALSE | 是否压缩 |
| include_rules | TEXT | | 包含规则（JSON数组） |
| exclude_rules | TEXT | | 排除规则（JSON数组） |
| max_file_size | INTEGER | | 最大文件大小（字节） |
| min_file_size | INTEGER | | 最小文件大小（字节） |
| created_at | TEXT | DEFAULT CURRENT_TIMESTAMP | 创建时间 |
| updated_at | TEXT | DEFAULT CURRENT_TIMESTAMP | 更新时间 |

**backup_records 表**（备份记录）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ID | INTEGER | PRIMARY KEY AUTOINCREMENT | 记录唯一标识 |
| task_id | INTEGER | NOT NULL | 关联任务ID |
| task_name | TEXT | NOT NULL | 任务名称（冗余） |
| version_id | TEXT | NOT NULL UNIQUE | 备份版本ID |
| backup_filename | TEXT | NOT NULL | 备份文件名 |
| backup_size | INTEGER | NOT NULL | 文件大小（字节） |
| status | BOOLEAN | NOT NULL | 备份状态 |
| failure_message | TEXT | | 失败信息 |
| checksum | TEXT | | 校验码 |
| storage_path | TEXT | NOT NULL | 文件完整路径 |
| created_at | TEXT | DEFAULT CURRENT_TIMESTAMP | 创建时间 |

### 8.2 索引

| 索引名 | 目标表 | 索引列 | 用途 |
|--------|--------|--------|------|
| idx_backup_tasks_name | backup_tasks | name | 按名称查询任务 |
| idx_backup_records_created_at | backup_records | created_at | 按时间排序/过滤 |
| idx_backup_records_task_id | backup_records | task_id | 按任务查询记录 |
| idx_backup_records_task_name | backup_records | task_name | 按任务名查询记录 |

---

## 九、命令行接口一览

### 9.1 命令列表

| 命令 | 简写 | 功能 | 关键参数 |
|------|------|------|---------|
| `bakctl add` | `a` | 添加备份任务 | `--name`, `--backup-dir`, `--storage-dir`, `--config`, `--compress` |
| `bakctl edit` | `e` | 编辑任务配置 | `--id`, `--ids`, `--retain-count`, `--compress` |
| `bakctl list` | `ls` | 列出所有任务 | `--table-style`, `--simple` |
| `bakctl run` | `r` | 执行备份任务 | `--id`, `--ids`, `--all` |
| `bakctl log` | `lg` | 查看备份日志 | `--id`, `--name`, `--limit`, `--failed` |
| `bakctl delete` | `del` | 删除备份任务 | `--id`, `--ids`, `--force`, `--keep-files`, `--failed` |
| `bakctl restore` | `rs` | 恢复备份 | `--id`, `--vid`, `--latest`, `-d` |
| `bakctl export` | `exp` | 导出配置 | `--id`, `--ids`, `--all`, `--cmd`, `--script`, `--bat`, `--sh` |

### 9.2 全局选项

| 选项 | 简写 | 说明 |
|------|------|------|
| `--no-color` | `-nc` | 禁用彩色输出 |
| `--version` | `-v` | 显示版本信息 |
| `--help` | `-h` | 显示帮助信息 |

---

## 十、项目总结

### 10.1 项目核心特点

1. **架构规范**: 遵循Go项目标准布局（cmd/internal/vendor），职责划分清晰
2. **子命令模式**: 8个子命令各自独立包，扩展新命令只需新建包+注册
3. **自研生态**: 使用作者自研的qflag/comprx/colorlib/go-kit/verman库，形成完整工具链
4. **数据安全**: SHA-1哈希校验、数据库记录、孤儿清理等多重保障
5. **用户体验**: 中文帮助信息、彩色输出、进度条、多种表格样式
6. **跨平台**: pure Go SQLite驱动，无需CGO，支持Windows/Linux/macOS
7. **完整文档**: README内容丰富，代码注释规范，包注释完整

### 10.2 待优化点

| 优先级 | 优化项 | 说明 |
|--------|--------|------|
| **中** | 孤儿记录全量扫描 | `db.CleanupOrphanRecords`每次执行都遍历所有记录，大数据库时性能待优化 |
| **中** | 测试覆盖 | 项目无单元测试文件（`*_test.go`），建议为核心模块添加测试 |
| **低** | export命令风格统一 | `ExportCmdMain`签名未传入colorlib，与其他子命令风格不一致 |
| **低** | 正则预编译 | cleanup模块的正则每次清理重新编译，可预编译提升微小性能 |
| **低** | 数据库迁移机制 | DDL脚本硬编码，缺少版本迁移机制 |
| **低** | 导入功能 | README提到支持批量导入，但代码中未实现独立的import子命令 |

### 10.3 关键记忆点

1. **项目定位**: BakCtl是一个Go语言CLI备份管理工具，使用SQLite存储元数据
2. **入口文件**: `main.go` → `cmd/bakctl/bakctl.go` → 各子命令
3. **数据库**: SQLite位于`~/.bakctl/bakctl.db3`，两张表backup_tasks和backup_records
4. **备份文件**: 默认存储在`~/.bakctl/bak/<任务名>/`目录，ZIP格式
5. **清理策略**: 支持按数量、按天数、组合三种保留策略，含安全检查（至少保留1个）
6. **作者自研库**: qflag(CLI解析), comprx(压缩), colorlib(彩色), go-kit(工具), verman(版本)
7. **构建方式**: `build.py` Python脚本或`gobf/*.toml`配置（gob构建工具）
8. **关键常量**: HashAlgorithm="sha1", BackupFileExt=".zip", DBFilename="bakctl.db3"
