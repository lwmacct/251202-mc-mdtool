package mdtoc

import (
	"os"
	"strings"
)

// TOC 是主要的门面结构，封装所有 TOC 生成功能
type TOC struct {
	parser    *Parser
	generator *Generator
	marker    *MarkerHandler
	options   Options
}

// New 创建新的 TOC 实例
func New(opts Options) *TOC {
	return &TOC{
		parser:    NewParser(opts),
		generator: NewGenerator(opts),
		marker:    NewMarkerHandler(DefaultMarker),
		options:   opts,
	}
}

// GenerateFromFile 从文件生成 TOC 字符串
func (t *TOC) GenerateFromFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return t.GenerateFromContent(content)
}

// GenerateFromContent 从内容生成 TOC 字符串
func (t *TOC) GenerateFromContent(content []byte) (string, error) {
	headers, err := t.parser.Parse(content)
	if err != nil {
		return "", err
	}
	return t.generator.Generate(headers), nil
}

// GenerateSectionTOCs 生成章节模式的 TOC (每个 H1 有独立的子目录)
func (t *TOC) GenerateSectionTOCs(content []byte) ([]SectionTOC, error) {
	// 解析所有标题
	headers, err := t.parser.ParseAllHeaders(content)
	if err != nil {
		return nil, err
	}

	// 按 H1 分割成章节
	sections := SplitSections(headers)

	// 为每个章节生成 TOC
	var sectionTOCs []SectionTOC
	for _, section := range sections {
		toc := t.generator.GenerateSection(section)
		if toc != "" {
			sectionTOCs = append(sectionTOCs, SectionTOC{
				H1Line: section.Title.Line - 1, // 转换为 0-based
				TOC:    toc,
			})
		}
	}

	return sectionTOCs, nil
}

// GenerateSectionTOCsPreview 生成章节模式的 TOC 预览 (用于 stdout 输出)
func (t *TOC) GenerateSectionTOCsPreview(content []byte) (string, error) {
	// 解析所有标题
	headers, err := t.parser.ParseAllHeaders(content)
	if err != nil {
		return "", err
	}

	// 按 H1 分割成章节
	sections := SplitSections(headers)

	var sb strings.Builder
	for i, section := range sections {
		toc := t.generator.GenerateSection(section)
		if toc != "" {
			sb.WriteString("### ")
			sb.WriteString(section.Title.Text)
			sb.WriteString("\n\n")
			sb.WriteString(toc)
			if i < len(sections)-1 {
				sb.WriteString("\n\n")
			}
		}
	}

	return sb.String(), nil
}

// UpdateFile 原地更新文件中的 TOC
// 如果文件没有 TOC 标记，会自动在第一个标题后插入
func (t *TOC) UpdateFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var newContent []byte

	if t.options.SectionTOC {
		// 章节模式：在每个 H1 后插入独立的子目录
		sectionTOCs, err := t.GenerateSectionTOCs(content)
		if err != nil {
			return err
		}
		newContent = t.marker.UpdateSectionTOCs(content, sectionTOCs)
	} else {
		// 普通模式：在 <!--TOC--> 标记处插入完整 TOC
		toc, err := t.GenerateFromContent(content)
		if err != nil {
			return err
		}

		markers := t.marker.FindMarkers(content)
		if markers.Found {
			newContent = t.marker.InsertTOC(content, toc)
		} else {
			newContent = t.marker.InsertTOCAfterFirstHeading(content, toc)
		}
	}

	return os.WriteFile(filename, newContent, 0644)
}

// CheckDiff 检查 TOC 是否需要更新
// 返回 true 表示需要更新 (有差异)
func (t *TOC) CheckDiff(filename string) (bool, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}

	// 生成新的 TOC
	newTOC, err := t.GenerateFromContent(content)
	if err != nil {
		return false, err
	}

	// 提取现有 TOC
	existingTOC := t.marker.ExtractExistingTOC(content)

	// 比较
	return newTOC != existingTOC, nil
}

// HasMarker 检查文件是否包含 TOC 标记
func (t *TOC) HasMarker(filename string) (bool, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return false, err
	}
	markers := t.marker.FindMarkers(content)
	return markers.Found, nil
}
