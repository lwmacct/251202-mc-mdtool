package mdtoc

import (
	"bytes"
	"strings"
)

// MarkerHandler 处理 <!--TOC--> 标记
type MarkerHandler struct {
	marker string
}

// NewMarkerHandler 创建新的标记处理器
func NewMarkerHandler(marker string) *MarkerHandler {
	if marker == "" {
		marker = DefaultMarker
	}
	return &MarkerHandler{marker: marker}
}

// FindMarkers 查找 TOC 标记位置
func (h *MarkerHandler) FindMarkers(content []byte) *TOCMarker {
	lines := bytes.Split(content, []byte("\n"))
	markerBytes := []byte(h.marker)

	var positions []int
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.Equal(trimmed, markerBytes) {
			positions = append(positions, i)
		}
	}

	result := &TOCMarker{
		StartLine: -1,
		EndLine:   -1,
		Found:     false,
	}

	if len(positions) >= 1 {
		result.StartLine = positions[0]
		result.Found = true
	}
	if len(positions) >= 2 {
		result.EndLine = positions[1]
	}

	return result
}

// InsertTOC 在标记位置插入或替换 TOC
func (h *MarkerHandler) InsertTOC(content []byte, toc string) []byte {
	markers := h.FindMarkers(content)

	// 没有找到标记，返回原内容
	if !markers.Found {
		return content
	}

	lines := bytes.Split(content, []byte("\n"))
	var result [][]byte

	if markers.EndLine == -1 {
		// 只有一个标记：在标记后插入 TOC，并添加结束标记
		for i, line := range lines {
			result = append(result, line)
			if i == markers.StartLine {
				// 添加空行 + TOC + 空行 + 结束标记
				result = append(result, []byte(""))
				result = append(result, []byte(toc))
				result = append(result, []byte(""))
				result = append(result, []byte(h.marker))
			}
		}
	} else {
		// 两个标记：替换两个标记之间的内容
		for i, line := range lines {
			if i < markers.StartLine {
				result = append(result, line)
			} else if i == markers.StartLine {
				result = append(result, line)
				result = append(result, []byte(""))
				result = append(result, []byte(toc))
				result = append(result, []byte(""))
			} else if i == markers.EndLine {
				result = append(result, line)
			} else if i > markers.EndLine {
				result = append(result, line)
			}
			// 跳过 StartLine+1 到 EndLine-1 之间的内容
		}
	}

	return bytes.Join(result, []byte("\n"))
}

// ExtractExistingTOC 提取现有的 TOC 内容 (两个标记之间的内容)
func (h *MarkerHandler) ExtractExistingTOC(content []byte) string {
	markers := h.FindMarkers(content)

	if !markers.Found || markers.EndLine == -1 {
		return ""
	}

	lines := bytes.Split(content, []byte("\n"))

	// 提取两个标记之间的内容
	var tocLines []string
	for i := markers.StartLine + 1; i < markers.EndLine; i++ {
		tocLines = append(tocLines, string(lines[i]))
	}

	// 去除首尾空行
	result := strings.TrimSpace(strings.Join(tocLines, "\n"))
	return result
}

// FindFirstHeading 查找第一个标题所在行 (0-based)
// 返回 -1 表示未找到标题
func (h *MarkerHandler) FindFirstHeading(content []byte) int {
	lines := bytes.Split(content, []byte("\n"))
	inCodeBlock := false

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)

		// 检测代码块
		if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// 检测 ATX 标题 (# ~ ######)
		if len(trimmed) > 0 && trimmed[0] == '#' {
			// 确保是有效标题 (# 后有空格或直接结束)
			for j := 0; j < len(trimmed) && j < 6; j++ {
				if trimmed[j] != '#' {
					break
				}
				if j+1 < len(trimmed) && (trimmed[j+1] == ' ' || trimmed[j+1] == '\t') {
					return i
				}
				if j+1 == len(trimmed) {
					return i // 只有 # 的行也算标题
				}
			}
		}
	}

	return -1
}

// InsertTOCAfterFirstHeading 在第一个标题后插入 TOC
func (h *MarkerHandler) InsertTOCAfterFirstHeading(content []byte, toc string) []byte {
	firstHeading := h.FindFirstHeading(content)
	if firstHeading == -1 {
		// 没有标题，在文件开头插入
		firstHeading = -1
	}

	lines := bytes.Split(content, []byte("\n"))
	var result [][]byte

	insertLine := firstHeading // 在标题行后插入

	for i, line := range lines {
		result = append(result, line)
		if i == insertLine {
			// 在标题后插入空行 + 标记 + TOC + 标记
			result = append(result, []byte(""))
			result = append(result, []byte(h.marker))
			result = append(result, []byte(""))
			result = append(result, []byte(toc))
			result = append(result, []byte(""))
			result = append(result, []byte(h.marker))
		}
	}

	// 如果没有找到标题 (firstHeading == -1)，在开头插入
	if firstHeading == -1 {
		var newResult [][]byte
		newResult = append(newResult, []byte(h.marker))
		newResult = append(newResult, []byte(""))
		newResult = append(newResult, []byte(toc))
		newResult = append(newResult, []byte(""))
		newResult = append(newResult, []byte(h.marker))
		newResult = append(newResult, []byte(""))
		newResult = append(newResult, result...)
		result = newResult
	}

	return bytes.Join(result, []byte("\n"))
}
