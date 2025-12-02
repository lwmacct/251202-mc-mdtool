package toc

import "github.com/urfave/cli/v3"

// Command 返回 toc 子命令
func Command() *cli.Command {
	return &cli.Command{
		Name:    "toc",
		Usage:   "生成 Markdown 目录 (Table of Contents)",
		UsageText: `mc-mdtool toc [options] <file>...
   fd -e md | mc-mdtool toc`,
		Aliases: []string{"t"},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "min-level",
				Aliases: []string{"m"},
				Value:   1,
				Usage:   "最小标题层级 (1-6)",
			},
			&cli.IntFlag{
				Name:    "max-level",
				Aliases: []string{"M"},
				Value:   3,
				Usage:   "最大标题层级 (1-6)",
			},
			&cli.BoolFlag{
				Name:    "in-place",
				Aliases: []string{"i"},
				Usage:   "原地更新文件 (在 <!--TOC--> 标记处插入)",
			},
			&cli.BoolFlag{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "检查 TOC 是否需要更新 (返回码 128 表示有差异)",
			},
			&cli.BoolFlag{
				Name:    "ordered",
				Aliases: []string{"o"},
				Usage:   "使用有序列表 (1. 2. 3.)",
			},
			&cli.BoolFlag{
				Name:    "line-number",
				Aliases: []string{"L"},
				Value:   true,
				Usage:   "显示行号范围 (:start:end)",
			},
			&cli.BoolFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "显示文件路径 (path:start:end)",
			},
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Usage:   "全局模式: 生成完整文档的单一目录 (默认为章节模式)",
			},
		},
		Action: action,
	}
}
