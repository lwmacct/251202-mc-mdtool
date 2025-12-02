package toc

import "github.com/urfave/cli/v3"

// Command 返回 toc 子命令
func Command() *cli.Command {
	return &cli.Command{
		Name:    "toc",
		Usage:   "生成 Markdown 目录 (Table of Contents)",
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
				Usage:   "显示行号范围 (:start-end)",
			},
			&cli.BoolFlag{
				Name:    "section",
				Aliases: []string{"s"},
				Usage:   "章节模式: 在每个 H1 后生成独立的子目录",
			},
		},
		Action: action,
	}
}
