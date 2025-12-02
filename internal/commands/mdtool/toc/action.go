package toc

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

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
	lineNumber := cmd.Bool("line-number")

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

	// 收集要处理的文件
	files := collectFiles(cmd.Args().Slice())
	if len(files) == 0 {
		return fmt.Errorf("请指定要处理的 Markdown 文件")
	}

	slog.Debug("处理 Markdown 文件",
		"files", files,
		"count", len(files),
		"min_level", minLevel,
		"max_level", maxLevel,
		"in_place", inPlace,
		"diff", diff,
		"ordered", ordered,
		"line_number", lineNumber,
	)

	// 创建 TOC 实例
	toc := mdtoc.New(mdtoc.Options{
		MinLevel:   int(minLevel),
		MaxLevel:   int(maxLevel),
		Ordered:    ordered,
		LineNumber: lineNumber,
	})

	// 根据模式执行不同操作
	switch {
	case diff:
		return processDiff(toc, files)
	case inPlace:
		return processInPlace(toc, files)
	default:
		return processStdout(toc, files)
	}
}

// collectFiles 收集要处理的文件列表
// 优先从命令行参数获取，如果没有则尝试从 stdin 读取
func collectFiles(args []string) []string {
	var files []string

	// 从命令行参数收集
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg != "" {
			files = append(files, arg)
		}
	}

	// 如果没有参数，尝试从 stdin 读取
	if len(files) == 0 && !isTerminal(os.Stdin) {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				files = append(files, line)
			}
		}
	}

	return files
}

// isTerminal 检查文件是否是终端
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return true
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// processDiff 检查差异模式 - 任一文件有差异返回非零
func processDiff(toc *mdtoc.TOC, files []string) error {
	hasAnyDiff := false

	for _, file := range files {
		if err := checkFileExists(file); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		hasDiff, err := toc.CheckDiff(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		if hasDiff {
			fmt.Fprintf(os.Stderr, "%s: TOC 需要更新\n", file)
			hasAnyDiff = true
		} else {
			fmt.Printf("%s: TOC 已是最新\n", file)
		}
	}

	if hasAnyDiff {
		os.Exit(ExitCodeDiff)
	}
	return nil
}

// processInPlace 原地更新模式
// 如果文件没有 TOC 标记，会自动在第一个标题后插入
func processInPlace(toc *mdtoc.TOC, files []string) error {
	var errors []string

	for _, file := range files {
		if err := checkFileExists(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		hasMarker, _ := toc.HasMarker(file)

		if err := toc.UpdateFile(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		if hasMarker {
			fmt.Printf("%s: 已更新\n", file)
		} else {
			fmt.Printf("%s: 已插入 (在第一个标题后)\n", file)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("部分文件处理失败:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}

// processStdout 输出到 stdout 模式
func processStdout(toc *mdtoc.TOC, files []string) error {
	for i, file := range files {
		if err := checkFileExists(file); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		tocStr, err := toc.GenerateFromFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		// 多文件时添加文件名标题
		if len(files) > 1 {
			fmt.Printf("## %s\n\n", file)
		}

		fmt.Println(tocStr)

		// 多文件时添加分隔
		if len(files) > 1 && i < len(files)-1 {
			fmt.Println()
		}
	}

	return nil
}

// checkFileExists 检查文件是否存在
func checkFileExists(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在")
	}
	return nil
}
