# fmt 子命令

<!--TOC-->

- [命令行接口](#命令行接口) `:18:33`
- [格式化规则](#格式化规则) `:34:44`
- [proseWrap 选项](#prosewrap-选项) `:45:52`
- [配置文件](#配置文件) `:53:71`
- [技术难点](#技术难点) `:72:80`
- [参考项目](#参考项目) `:81:87`

<!--TOC-->

> **状态**: 📋 计划中 (P2)

格式化 Markdown 文件，参考 [Prettier](https://prettier.io/) 设计理念。

## 命令行接口

```shell
mc-mdtool fmt [options] <file>

Options:
  -i, --in-place       原地更新文件
  -w, --print-width    目标行宽 (默认 80, 0=不限制)
  --prose-wrap         段落换行: always|never|preserve (默认 preserve)
  --tab-width          缩进空格数 (默认 2)
  --use-tabs           使用 Tab 缩进
  --end-of-line        行尾符: lf|crlf|auto (默认 lf)
  --code               格式化代码块 (需要外部格式化器)
  -c, --config         配置文件路径
```

## 格式化规则

| 元素   | 规则                            |
| ------ | ------------------------------- |
| 标题   | `#` 后必须有且只有一个空格      |
| 列表   | 列表标记后一个空格，嵌套 2 空格 |
| 表格   | 单元格内容两侧加空格            |
| 链接   | 移除 URL 前后空格               |
| 空行   | 标题后一个空行，段落间一个空行  |
| 代码块 | 可选调用外部格式化器            |

## proseWrap 选项

| 值         | 说明                     |
| ---------- | ------------------------ |
| `preserve` | 保持原有换行 (默认)      |
| `always`   | 超过 printWidth 自动换行 |
| `never`    | 移除段落内换行，每段一行 |

## 配置文件

支持项目配置文件 `.mdtool.yaml`：

```yaml
fmt:
  print-width: 80
  prose-wrap: preserve
  tab-width: 2
  use-tabs: false
  end-of-line: lf
  code-formatters:
    go: gofmt
    js: prettier --parser babel
    python: black -
```

**优先级**：命令行参数 > 项目配置 > 用户配置 > 默认值

## 技术难点

| 难点         | 解决方案               |
| ------------ | ---------------------- |
| 保留注释     | 预处理标记             |
| 代码块格式化 | 类似 mdsf 调用外部工具 |
| 表格对齐     | runewidth 库处理 CJK   |
| 原文保留     | preserve 模式          |

## 参考项目

| 项目     | 语言   | 特点                    |
| -------- | ------ | ----------------------- |
| Prettier | JS     | 业界标准，opinionated   |
| mdformat | Python | CommonMark 兼容，可扩展 |
| mdsf     | Rust   | 专注代码块格式化        |
