package mdtoc

import (
	"strconv"
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
	return g.generateTOC(headers, g.options.MinLevel)
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

	// 找到最小层级作为基准 (章节模式下通常是 H2)
	minLevel := 6
	for _, h := range filteredHeaders {
		if h.Level < minLevel {
			minLevel = h.Level
		}
	}

	return g.generateTOC(filteredHeaders, minLevel)
}

// generateTOC 生成 TOC 字符串的内部实现
// baseLevel 用于计算缩进的基准层级
func (g *Generator) generateTOC(headers []*Header, baseLevel int) string {
	var sb strings.Builder
	orderedCounters := make(map[int]int)

	for i, h := range headers {
		// 计算缩进 (相对于基准层级)
		indent := (h.Level - baseLevel) * 2
		indentStr := strings.Repeat(" ", indent)

		// 生成列表标记
		var marker string
		if g.options.Ordered {
			orderedCounters[h.Level]++
			// 重置更深层级的计数器
			for level := h.Level + 1; level <= 6; level++ {
				orderedCounters[level] = 0
			}
			marker = strconv.Itoa(orderedCounters[h.Level]) + "."
		} else {
			marker = "-"
		}

		// 生成链接：ShowAnchor 控制是否包含 (#anchor) 部分
		var link string
		if g.options.ShowAnchor {
			link = "[" + h.Text + "](#" + h.AnchorLink + ")"
		} else {
			link = "[" + h.Text + "]"
		}

		// 添加行号范围 (LLM 友好格式: :start+count)
		if g.options.LineNumber && h.Line > 0 {
			count := h.EndLine - h.Line + 1
			if g.options.ShowPath && g.options.FilePath != "" {
				link += " `" + g.options.FilePath + ":" + strconv.Itoa(h.Line) + "+" + strconv.Itoa(count) + "`"
			} else {
				link += " `:" + strconv.Itoa(h.Line) + "+" + strconv.Itoa(count) + "`"
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
