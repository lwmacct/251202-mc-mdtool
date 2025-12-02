package toc

import (
	"bufio"
	"context"
	"fmt"
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
	showPath := cmd.Bool("path")
	globalMode := cmd.Bool("global")

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
		// 无文件时显示帮助
		return cli.ShowSubcommandHelp(cmd)
	}

	// 创建基础选项
	// 默认启用章节模式 (SectionTOC=true)，只有指定 --global 才使用全局模式
	baseOpts := mdtoc.Options{
		MinLevel:   int(minLevel),
		MaxLevel:   int(maxLevel),
		Ordered:    ordered,
		LineNumber: lineNumber,
		ShowPath:   showPath,
		SectionTOC: !globalMode,
	}

	// 根据模式执行不同操作
	switch {
	case diff:
		return processDiff(mdtoc.New(baseOpts), files)
	case inPlace:
		return processInPlace(mdtoc.New(baseOpts), files)
	default:
		return processStdout(baseOpts, files)
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
func processStdout(baseOpts mdtoc.Options, files []string) error {
	for i, file := range files {
		if err := checkFileExists(file); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		// 为每个文件创建带有文件路径的 TOC 实例
		opts := baseOpts
		opts.FilePath = file
		toc := mdtoc.New(opts)

		var tocStr string
		var err error

		if opts.SectionTOC {
			// 章节模式：预览每个 H1 的子目录
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", file, readErr)
				continue
			}
			tocStr, err = toc.GenerateSectionTOCsPreview(content)
		} else {
			tocStr, err = toc.GenerateFromFile(file)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", file, err)
			continue
		}

		// 跳过空的 TOC
		if strings.TrimSpace(tocStr) == "" {
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
