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

// Parse 解析 Markdown 内容，返回标题列表
func (p *Parser) Parse(content []byte) ([]*Header, error) {
	// 重置锚点生成器
	p.anchor.Reset()

	// 解析为 AST
	reader := text.NewReader(content)
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

		// 检查层级范围
		if heading.Level < p.options.MinLevel || heading.Level > p.options.MaxLevel {
			return ast.WalkSkipChildren, nil
		}

		// 提取标题文本
		text := extractText(content, heading)

		// 生成 anchor link
		anchor := p.anchor.Generate(text)

		headers = append(headers, &Header{
			Level:      heading.Level,
			Text:       text,
			AnchorLink: anchor,
		})

		return ast.WalkSkipChildren, nil
	})

	return headers, err
}

// extractText 从标题节点提取纯文本内容
func extractText(src []byte, n ast.Node) string {
	var buf bytes.Buffer
	writeNodeText(src, &buf, n)
	return buf.String()
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
