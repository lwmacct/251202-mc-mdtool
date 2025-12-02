# mc-mdtool

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
go install github.com/lwmacct/251202-mc-mdtool/cmd/mc-mdtool@latest
```

## 使用示例

```shell
# 查看帮助
mc-mdtool --help
mc-mdtool toc --help

# 生成 TOC 到 stdout
mc-mdtool toc README.md

# 原地更新文件 (在 <!--TOC--> 标记处插入)
mc-mdtool toc -i README.md

# 检查 TOC 是否需要更新 (CI 场景)
mc-mdtool toc -d README.md
```

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

- [开发计划](docs/content/mdtoc-design.md)

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
