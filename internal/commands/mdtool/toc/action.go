package toc

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lwmacct/251202-mc-mdtool/internal/mdtoc"
	"github.com/urfave/cli/v3"
)

// ExitCodeDiff 是 --diff 检测到差异时的退出码
const ExitCodeDiff = 128

func action(ctx context.Context, cmd *cli.Command) error {
	// 解析命令行参数
	minLevel := cmd.Int("min-level")
	maxLevel := cmd.Int("max-level")
	inPlace := cmd.Bool("in-place")
	diff := cmd.Bool("diff")
	ordered := cmd.Bool("ordered")

	// 获取文件参数
	file := cmd.Args().First()
	if file == "" {
		return fmt.Errorf("请指定要处理的 Markdown 文件")
	}

	// 检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", file)
	}

	// 验证层级参数
	if minLevel < 1 || minLevel > 6 {
		return fmt.Errorf("min-level 必须在 1-6 之间")
	}
	if maxLevel < 1 || maxLevel > 6 {
		return fmt.Errorf("max-level 必须在 1-6 之间")
	}
	if minLevel > maxLevel {
		return fmt.Errorf("min-level 不能大于 max-level")
	}

	slog.Debug("处理 Markdown 文件",
		"file", file,
		"min_level", minLevel,
		"max_level", maxLevel,
		"in_place", inPlace,
		"diff", diff,
		"ordered", ordered,
	)

	// 创建 TOC 实例
	toc := mdtoc.New(mdtoc.Options{
		MinLevel: int(minLevel),
		MaxLevel: int(maxLevel),
		Ordered:  ordered,
	})

	// 根据模式执行不同操作
	switch {
	case diff:
		// 检查差异模式
		hasDiff, err := toc.CheckDiff(file)
		if err != nil {
			return fmt.Errorf("检查差异失败: %w", err)
		}
		if hasDiff {
			fmt.Fprintln(os.Stderr, "TOC 需要更新")
			os.Exit(ExitCodeDiff)
		}
		fmt.Println("TOC 已是最新")
		return nil

	case inPlace:
		// 原地更新模式
		hasMarker, err := toc.HasMarker(file)
		if err != nil {
			return fmt.Errorf("检查标记失败: %w", err)
		}
		if !hasMarker {
			return fmt.Errorf("文件中未找到 %s 标记", mdtoc.DefaultMarker)
		}

		if err := toc.UpdateFile(file); err != nil {
			return fmt.Errorf("更新文件失败: %w", err)
		}
		fmt.Printf("已更新 %s 的目录\n", file)
		return nil

	default:
		// 输出到 stdout 模式
		tocStr, err := toc.GenerateFromFile(file)
		if err != nil {
			return fmt.Errorf("生成 TOC 失败: %w", err)
		}
		fmt.Println(tocStr)
		return nil
	}
}
