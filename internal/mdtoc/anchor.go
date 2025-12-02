package mdtoc

import (
	"regexp"
	"strings"
	"unicode"
)

// AnchorGenerator 生成 GitHub 风格的 anchor link
type AnchorGenerator struct {
	counter map[string]int // 重复标题计数器
}

// NewAnchorGenerator 创建新的锚点生成器
func NewAnchorGenerator() *AnchorGenerator {
	return &AnchorGenerator{
		counter: make(map[string]int),
	}
}

// Reset 重置计数器
func (g *AnchorGenerator) Reset() {
	g.counter = make(map[string]int)
}

// Generate 生成 anchor link
// 规则 (参考 GitHub):
// 1. 转小写
// 2. 移除 HTML 标签
// 3. 移除强调符号 (*, _, ~)
// 4. 保留 Unicode 字母、数字、连字符、空格
// 5. 空格转连字符
// 6. 处理重复标题 (添加 -1, -2, ...)
func (g *AnchorGenerator) Generate(text string) string {
	// 1. 转小写
	anchor := strings.ToLower(text)

	// 2. 移除 HTML 标签
	anchor = removeHTMLTags(anchor)

	// 3. 移除强调符号和代码标记
	anchor = removeEmphasis(anchor)

	// 4. 保留 Unicode 字母、数字、连字符、空格
	anchor = filterCharacters(anchor)

	// 5. 空格转连字符，合并多个连字符
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = mergeHyphens(anchor)
	anchor = strings.Trim(anchor, "-")

	// 6. 处理重复标题
	anchor = g.handleDuplicate(anchor)

	return anchor
}

// removeHTMLTags 移除 HTML 标签
func removeHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// removeEmphasis 移除 Markdown 强调符号
func removeEmphasis(s string) string {
	// 移除 **, *, __, ~~, `
	// 注意: 下划线斜体 _text_ 只在下划线位于单词边界时生效
	patterns := []string{
		`\*\*(.+?)\*\*`,            // **bold**
		`\*(.+?)\*`,                // *italic*
		`__(.+?)__`,                // __bold__
		`(?:^|[\s])_([^_]+?)_(?:[\s]|$)`, // _italic_ (只匹配单词边界的下划线)
		`~~(.+?)~~`,                // ~~strikethrough~~
		"`(.+?)`",                  // `code`
	}

	result := s
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "$1")
	}

	// 移除链接语法 [text](url) -> text
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	result = linkRe.ReplaceAllString(result, "$1")

	// 移除图片语法 ![alt](url) -> alt
	imgRe := regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	result = imgRe.ReplaceAllString(result, "$1")

	return result
}

// filterCharacters 保留 Unicode 字母、数字、连字符、下划线、空格
func filterCharacters(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// mergeHyphens 合并多个连续的连字符为一个
func mergeHyphens(s string) string {
	re := regexp.MustCompile(`-+`)
	return re.ReplaceAllString(s, "-")
}

// handleDuplicate 处理重复标题
func (g *AnchorGenerator) handleDuplicate(anchor string) string {
	count, exists := g.counter[anchor]
	g.counter[anchor] = count + 1

	if exists {
		return anchor + "-" + itoa(count)
	}
	return anchor
}

// itoa 简单的整数转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
