// Package mdtoc 提供 Markdown 目录生成功能
package mdtoc

// Header 表示一个 Markdown 标题
type Header struct {
	Level      int    // 标题层级 (1-6)
	Text       string // 标题文本 (原始文本，去除 # 和前后空格)
	AnchorLink string // 锚点链接 (GitHub 风格)
}

// Options 配置 TOC 生成选项
type Options struct {
	MinLevel int  // 最小标题层级 (默认 1)
	MaxLevel int  // 最大标题层级 (默认 3)
	Ordered  bool // 使用有序列表
}

// DefaultOptions 返回默认配置
func DefaultOptions() Options {
	return Options{
		MinLevel: 1,
		MaxLevel: 3,
		Ordered:  false,
	}
}

// TOCMarker 表示 TOC 标记位置
type TOCMarker struct {
	StartLine int // 第一个标记所在行号 (0-based)
	EndLine   int // 第二个标记所在行号 (0-based), -1 表示只有一个标记
	Found     bool
}

// DefaultMarker 是默认的 TOC 标记字符串
const DefaultMarker = "<!--TOC-->"
