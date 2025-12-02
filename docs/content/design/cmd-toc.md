# toc 子命令

> **状态**: ✅ 已完成 (Phase 1-2)

为 Markdown 文件自动生成符合规范的目录（Table of Contents）。

## 命令行接口

```shell
mc-mdtool toc [options] <file>...
   fd -e md | mc-mdtool toc

Options:
  -m, --min-level    最小标题层级 (默认 1)
  -M, --max-level    最大标题层级 (默认 3)
  -i, --in-place     原地更新文件
  -d, --diff         检查差异 (返回码 128 表示有差异)
  -o, --ordered      有序列表
  -L, --line-number  显示行号范围 :start:end (默认启用)
  -p, --path         显示文件路径 path:start:end
  -s, --section      章节模式
```

## 功能特性

| 功能         | 说明                               | 状态      |
| ------------ | ---------------------------------- | --------- |
| 标题解析     | 解析 ATX 风格标题 (`# ~ ######`)   | ✅ 已完成 |
| 锚点生成     | GitHub 规范 anchor link            | ✅ 已完成 |
| TOC 标记     | 支持 `<!--TOC-->` 标记定位         | ✅ 已完成 |
| 原地更新     | `-i` 直接修改文件                  | ✅ 已完成 |
| 差异检测     | `-d` 检查 TOC 是否需要更新         | ✅ 已完成 |
| 有序列表     | `-o` 生成 `1. 2. 3.` 格式          | ✅ 已完成 |
| 行号范围     | `-L` 显示 `:start:end`             | ✅ 已完成 |
| 文件路径     | `-p` 显示 `path:start:end`         | ✅ 已完成 |
| 章节模式     | `-s` 每个 H1 后生成独立子目录      | ✅ 已完成 |
| 多文件处理   | 支持多文件和管道输入               | ✅ 已完成 |
| 多框架支持   | VitePress、Hugo 等                 | 📋 计划中 |

## 输出格式

```shell
# 默认输出 (行号范围默认启用)
mc-mdtool toc README.md
# - [标题](#标题) `:1:10`

# 带文件路径
mc-mdtool toc -p README.md
# - [标题](#标题) `README.md:1:10`

# 禁用行号
mc-mdtool toc -L=false README.md
# - [标题](#标题)
```

## TOC 标记规范

使用 HTML 注释作为标记，渲染后不可见：

```markdown
<!--TOC-->

- [toc 子命令](#toc-子命令) `:1:6`
  - [命令行接口](#命令行接口) `:7:23`
  - [功能特性](#功能特性) `:24:39`
  - [输出格式](#输出格式) `:40:55`
  - [TOC 标记规范](#toc-标记规范) `:56:79`
  - [技术实现](#技术实现) `:80:93`
  - [参考项目](#参考项目) `:94:100`

<!--TOC-->
```

**更新逻辑**：
1. 查找第一个 `<!--TOC-->` 标记
2. 查找第二个 `<!--TOC-->` 标记（可选）
3. 替换两个标记之间的内容
4. 如果没有标记，在第一个标题后自动插入

## 技术实现

基于 [goldmark](https://github.com/yuin/goldmark) CommonMark 解析器。

**核心模块**：

| 文件             | 职责                          |
| ---------------- | ----------------------------- |
| `types.go`       | Header/Options 类型定义       |
| `parser.go`      | 解析 Markdown，提取标题       |
| `anchor.go`      | GitHub 风格 anchor link 生成  |
| `generator.go`   | TOC 字符串生成                |
| `marker.go`      | `<!--TOC-->` 标记处理         |

## 参考项目

| 项目         | 说明                  |
| ------------ | --------------------- |
| md-toc       | Python TOC 生成器     |
| gh-md-toc-go | Go GitHub TOC 生成器  |
| goldmark-toc | goldmark TOC 扩展     |
