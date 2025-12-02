package mdtoc

import (
	"strings"
)

// Generator 生成 TOC 字符串
type Generator struct {
	options Options
}

// NewGenerator 创建新的生成器
func NewGenerator(opts Options) *Generator {
	return &Generator{
		options: opts,
	}
}

// Generate 从标题列表生成 TOC 字符串
func (g *Generator) Generate(headers []*Header) string {
	if len(headers) == 0 {
		return ""
	}

	var sb strings.Builder
	orderedCounters := make(map[int]int) // 每个层级的有序列表计数器

	for i, h := range headers {
		// 计算缩进 (相对于最小层级)
		indent := (h.Level - g.options.MinLevel) * 2
		indentStr := strings.Repeat(" ", indent)

		// 生成列表标记
		var marker string
		if g.options.Ordered {
			orderedCounters[h.Level]++
			// 重置更深层级的计数器
			for level := h.Level + 1; level <= 6; level++ {
				orderedCounters[level] = 0
			}
			marker = itoa(orderedCounters[h.Level]) + "."
		} else {
			marker = "-"
		}

		// 生成链接
		link := "[" + h.Text + "](#" + h.AnchorLink + ")"

		// 添加行号范围 (VS Code 兼容格式)
		if g.options.LineNumber && h.Line > 0 {
			if g.options.ShowPath && g.options.FilePath != "" {
				link += " `" + g.options.FilePath + ":" + itoa(h.Line) + ":" + itoa(h.EndLine) + "`"
			} else {
				link += " `:" + itoa(h.Line) + ":" + itoa(h.EndLine) + "`"
			}
		}

		// 生成 TOC 行
		line := indentStr + marker + " " + link
		sb.WriteString(line)

		// 除最后一行外添加换行符
		if i < len(headers)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// GenerateSection 为单个章节生成 TOC (只包含子标题)
// 章节模式下，每个 H1 后面只生成该章节的子目录
// 要求：章节内至少包含一个 H2 才会生成 TOC
func (g *Generator) GenerateSection(section *Section) string {
	if section == nil || len(section.SubHeaders) == 0 {
		return ""
	}

	// 检查是否至少有一个 H2 (章节必须包含 H2 才生成 TOC)
	hasH2 := false
	for _, h := range section.SubHeaders {
		if h.Level == 2 {
			hasH2 = true
			break
		}
	}
	if !hasH2 {
		return ""
	}

	// 筛选符合层级范围的子标题
	var filteredHeaders []*Header
	for _, h := range section.SubHeaders {
		if h.Level >= g.options.MinLevel && h.Level <= g.options.MaxLevel {
			filteredHeaders = append(filteredHeaders, h)
		}
	}

	if len(filteredHeaders) == 0 {
		return ""
	}

	var sb strings.Builder
	orderedCounters := make(map[int]int)

	// 找到最小层级作为基准 (章节模式下通常是 H2)
	minLevel := 6
	for _, h := range filteredHeaders {
		if h.Level < minLevel {
			minLevel = h.Level
		}
	}

	for i, h := range filteredHeaders {
		// 计算缩进 (相对于章节内最小层级)
		indent := (h.Level - minLevel) * 2
		indentStr := strings.Repeat(" ", indent)

		var marker string
		if g.options.Ordered {
			orderedCounters[h.Level]++
			for level := h.Level + 1; level <= 6; level++ {
				orderedCounters[level] = 0
			}
			marker = itoa(orderedCounters[h.Level]) + "."
		} else {
			marker = "-"
		}

		link := "[" + h.Text + "](#" + h.AnchorLink + ")"
		if g.options.LineNumber && h.Line > 0 {
			if g.options.ShowPath && g.options.FilePath != "" {
				link += " `" + g.options.FilePath + ":" + itoa(h.Line) + ":" + itoa(h.EndLine) + "`"
			} else {
				link += " `:" + itoa(h.Line) + ":" + itoa(h.EndLine) + "`"
			}
		}

		line := indentStr + marker + " " + link
		sb.WriteString(line)

		if i < len(filteredHeaders)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
