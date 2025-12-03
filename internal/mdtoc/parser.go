package mdtoc

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser 解析 Markdown 文档并提取标题
type Parser struct {
	md      goldmark.Markdown
	anchor  *AnchorGenerator
	options Options
}

// NewParser 创建新的解析器
func NewParser(opts Options) *Parser {
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // 自动生成标题 ID
		),
	)
	return &Parser{
		md:      md,
		anchor:  NewAnchorGenerator(),
		options: opts,
	}
}

// Parse 解析 Markdown 内容，返回标题列表 (受 MinLevel/MaxLevel 限制)
func (p *Parser) Parse(content []byte) ([]*Header, error) {
	return p.parseHeaders(content, true)
}

// ParseAllHeaders 解析所有标题 (不受 MinLevel/MaxLevel 限制)
// 用于章节模式，需要完整的标题层级信息
func (p *Parser) ParseAllHeaders(content []byte) ([]*Header, error) {
	return p.parseHeaders(content, false)
}

// parseHeaders 解析标题的内部实现
// filterLevel 控制是否按 MinLevel/MaxLevel 过滤标题
func (p *Parser) parseHeaders(content []byte, filterLevel bool) ([]*Header, error) {
	// 重置锚点生成器
	p.anchor.Reset()

	// 检测并跳过 frontmatter
	lines := bytes.Split(content, []byte("\n"))
	frontmatterEnd := FindFrontmatterEnd(lines)
	lineOffset := 0
	parseContent := content

	if frontmatterEnd >= 0 {
		lineOffset = frontmatterEnd + 1
		parseContent = bytes.Join(lines[lineOffset:], []byte("\n"))
	}

	// 预计算行号映射 (byte offset -> line number)
	lineMap := buildLineMap(parseContent)
	totalLines := countLines(content)

	// 解析为 AST
	reader := text.NewReader(parseContent)
	doc := p.md.Parser().Parse(reader)

	var headers []*Header

	// 遍历 AST 提取标题
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		// 检查层级范围（仅在 filterLevel 为 true 时）
		if filterLevel && (heading.Level < p.options.MinLevel || heading.Level > p.options.MaxLevel) {
			return ast.WalkSkipChildren, nil
		}

		// 提取标题文本
		text := extractText(parseContent, heading)

		// 生成 anchor link
		anchor := p.anchor.Generate(text)

		// 获取行号（需要加上 frontmatter 的偏移）
		line := getNodeLine(heading, lineMap) + lineOffset

		headers = append(headers, &Header{
			Level:      heading.Level,
			Text:       text,
			AnchorLink: anchor,
			Line:       line,
		})

		return ast.WalkSkipChildren, nil
	})

	// 计算每个标题的结束行
	calculateEndLines(headers, totalLines)

	return headers, err
}

// buildLineMap 构建 byte offset 到行号的映射
func buildLineMap(content []byte) []int {
	// lineMap[i] = 第 i 个字节所在的行号 (1-based)
	lineMap := make([]int, len(content)+1)
	line := 1
	for i, b := range content {
		lineMap[i] = line
		if b == '\n' {
			line++
		}
	}
	lineMap[len(content)] = line
	return lineMap
}

// countLines 计算总行数
func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := 1
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}
	// 如果最后一个字符是换行，不额外计数
	if len(content) > 0 && content[len(content)-1] == '\n' {
		lines--
	}
	return lines
}

// getNodeLine 获取节点所在行号
func getNodeLine(n ast.Node, lineMap []int) int {
	if n.Lines().Len() > 0 {
		start := n.Lines().At(0).Start
		if start < len(lineMap) {
			return lineMap[start]
		}
	}
	return 0
}

// calculateEndLines 计算每个标题的结束行
// 规则：标题的结束行是下一个同级或更高级标题的前一行
// 这样父级标题会包含其所有子级内容
func calculateEndLines(headers []*Header, totalLines int) {
	for i, h := range headers {
		// 查找下一个同级或更高级的标题
		endLine := totalLines
		for j := i + 1; j < len(headers); j++ {
			if headers[j].Level <= h.Level {
				// 找到同级或更高级标题，结束行是其前一行
				endLine = headers[j].Line - 1
				break
			}
		}
		h.EndLine = max(endLine, h.Line)
	}
}

// extractText 从标题节点提取纯文本内容
func extractText(src []byte, n ast.Node) string {
	var buf bytes.Buffer
	writeNodeText(src, &buf, n)
	return buf.String()
}

// SplitSections 将标题列表按 H1 分割成章节
// 每个章节包含一个 H1 和其后续的子标题 (H2-H6)
func SplitSections(headers []*Header) []*Section {
	var sections []*Section
	var currentSection *Section

	for _, h := range headers {
		if h.Level == 1 {
			// 遇到新的 H1，创建新章节
			if currentSection != nil {
				sections = append(sections, currentSection)
			}
			currentSection = &Section{
				Title:      h,
				SubHeaders: []*Header{},
			}
		} else if currentSection != nil {
			// 当前在某个章节内，添加子标题
			currentSection.SubHeaders = append(currentSection.SubHeaders, h)
		}
		// 如果 currentSection == nil 且 h.Level != 1，
		// 说明在第一个 H1 之前有其他标题，跳过这些标题
	}

	// 添加最后一个章节
	if currentSection != nil {
		sections = append(sections, currentSection)
	}

	return sections
}

// writeNodeText 递归写入节点文本
func writeNodeText(src []byte, buf *bytes.Buffer, n ast.Node) {
	switch node := n.(type) {
	case *ast.Text:
		buf.Write(node.Segment.Value(src))
	case *ast.String:
		buf.Write(node.Value)
	case *ast.CodeSpan:
		// 对于行内代码，提取内容
		for c := node.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(src, buf, c)
		}
	default:
		// 递归处理子节点
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(src, buf, c)
		}
	}
}
