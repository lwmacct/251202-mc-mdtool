package mdtoc

import (
	"os"
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

// UpdateFile 原地更新文件中的 TOC
func (t *TOC) UpdateFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	toc, err := t.GenerateFromContent(content)
	if err != nil {
		return err
	}

	newContent := t.marker.InsertTOC(content, toc)
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
