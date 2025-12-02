# mc-mdtool

<!--TOC-->

- [mc-mdtool](#mc-mdtool) `:1:20`
  - [功能特性](#功能特性) `:21:29`
  - [安装](#安装) `:30:39`
  - [使用示例](#使用示例) `:40:78`
    - [toc 命令选项](#toc-命令选项) `:79:91`
  - [开发](#开发) `:92:93`
    - [环境准备](#环境准备) `:94:103`
    - [构建](#构建) `:104:109`
  - [设计文档](#设计文档) `:110:113`
  - [参考项目](#参考项目) `:114:124`
  - [相关链接](#相关链接) `:125:129`

<!--TOC-->

Markdown CLI 工具集，提供目录生成、格式化、检查等功能。

## 功能特性

| 子命令  | 说明                   | 状态      |
| ------- | ---------------------- | --------- |
| `toc`   | 生成 Table of Contents | ✅ 已完成 |
| `fmt`   | 格式化 Markdown        | 📋 计划中 |
| `lint`  | 检查 Markdown 规范     | 📋 计划中 |
| `links` | 检查链接有效性         | 📋 计划中 |

## 安装

```shell
# 从 GitHub 安装 (需要先发布)
go install github.com/lwmacct/251202-mc-mdtool/cmd/mc-mdtool@latest

# 本地构建安装
go install ./cmd/mc-mdtool
```

## 使用示例

```shell
# 查看帮助
mc-mdtool --help
mc-mdtool toc --help

# 生成 TOC 到 stdout
mc-mdtool toc README.md

# 显示行号范围 (默认启用, VS Code 兼容格式)
mc-mdtool toc README.md
# 输出: - [标题](#标题) `:1:10`

# 显示文件路径 + 行号范围
mc-mdtool toc -p README.md
# 输出: - [标题](#标题) `README.md:1:10`

# 禁用行号范围
mc-mdtool toc -L=false README.md

# 原地更新文件 (在 <!--TOC--> 标记处插入)
mc-mdtool toc -i README.md

# 检查 TOC 是否需要更新 (CI 场景)
mc-mdtool toc -d README.md

# 使用有序列表 + 指定层级
mc-mdtool toc -o -m 2 -M 4 README.md

# 多文件处理
mc-mdtool toc file1.md file2.md file3.md
mc-mdtool toc -i docs/*.md

# 管道输入 (从 stdin 读取文件列表)
find . -name "*.md" | mc-mdtool toc
find . -name "*.md" | mc-mdtool toc -i
```

### toc 命令选项

| 选项            | 短选项 | 说明                                 |
| --------------- | ------ | ------------------------------------ |
| `--min-level`   | `-m`   | 最小标题层级 (默认 1)                |
| `--max-level`   | `-M`   | 最大标题层级 (默认 3)                |
| `--in-place`    | `-i`   | 原地更新文件                         |
| `--diff`        | `-d`   | 检查是否需要更新                     |
| `--ordered`     | `-o`   | 使用有序列表                         |
| `--line-number` | `-L`   | 显示行号范围 `:start:end` (默认启用) |
| `--path`        | `-p`   | 显示文件路径 `path:start:end`        |
| `--section`     | `-s`   | 章节模式: 每个 H1 后生成独立子目录   |

## 开发

### 环境准备

```shell
# 安装 pre-commit hooks
pre-commit install

# 查看可用任务
task -a
```

### 构建

```shell
go build ./cmd/mc-mdtool/
```

## 设计文档

- [toc 命令设计](design/cmd-toc.md)
- [fmt 命令设计](design/cmd-fmt.md)
- [解析器参考](design/ref-parsers.md)

## 参考项目

| 项目                                                       | 语言    | 说明              |
| ---------------------------------------------------------- | ------- | ----------------- |
| [md-toc](https://github.com/frnmst/md-toc)                 | Python  | TOC 生成          |
| [goldmark](https://github.com/yuin/goldmark)               | Go      | CommonMark 解析器 |
| [glamour](https://github.com/charmbracelet/glamour)        | Go      | Markdown 渲染     |
| [mdsf](https://github.com/hougesen/mdsf)                   | Rust    | 代码块格式化      |
| [markdownlint](https://github.com/DavidAnson/markdownlint) | Node.js | Markdown 检查     |
| [lychee](https://github.com/lycheeverse/lychee)            | Rust    | 链接检查          |

## 相关链接

- [Taskfile](https://taskfile.dev) - 任务管理
- [Pre-commit](https://pre-commit.com/) - Git hooks 管理
- [CommonMark Spec](https://spec.commonmark.org/0.31.2/) - Markdown 规范
