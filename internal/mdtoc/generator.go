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

		// 生成 TOC 行
		line := indentStr + marker + " [" + h.Text + "](#" + h.AnchorLink + ")"
		sb.WriteString(line)

		// 除最后一行外添加换行符
		if i < len(headers)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
