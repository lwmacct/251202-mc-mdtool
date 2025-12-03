package mdtoc

import (
	"bytes"
	"strings"
)

// FindFrontmatterEnd 查找 YAML frontmatter 的结束位置
// 返回 frontmatter 结束行号 (0-based)，如果没有 frontmatter 则返回 -1
// YAML frontmatter 必须从文件第一行开始，以 "---" 标记
func FindFrontmatterEnd(lines [][]byte) int {
	if len(lines) == 0 {
		return -1
	}

	// 检查第一行是否为 frontmatter 开始标记
	firstLine := bytes.TrimSpace(lines[0])
	if !bytes.Equal(firstLine, []byte("---")) {
		return -1
	}

	// 查找结束标记 (第二个 "---" 或 "...")
	for i := 1; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
		if bytes.Equal(trimmed, []byte("---")) || bytes.Equal(trimmed, []byte("...")) {
			return i
		}
	}

	// 没有找到结束标记，说明 frontmatter 未闭合
	return -1
}

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
// 注意：会跳过 YAML frontmatter 内部的标记
func (h *MarkerHandler) FindMarkers(content []byte) *TOCMarker {
	lines := bytes.Split(content, []byte("\n"))
	markerBytes := []byte(h.marker)

	// 找到 frontmatter 结束位置
	frontmatterEnd := FindFrontmatterEnd(lines)
	startLine := 0
	if frontmatterEnd >= 0 {
		startLine = frontmatterEnd + 1 // 从 frontmatter 之后开始搜索
	}

	var positions []int
	for i := startLine; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])
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
		skipNextEmpty := false
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
				skipNextEmpty = true // 标记：跳过结束标记后的第一个空行
			} else if i > markers.EndLine {
				// 跳过结束标记后紧跟的空行，避免空行累积
				if skipNextEmpty && len(bytes.TrimSpace(line)) == 0 {
					skipNextEmpty = false
					continue
				}
				skipNextEmpty = false
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
// 注意：会跳过 YAML frontmatter 内部的内容
func (h *MarkerHandler) FindFirstHeading(content []byte) int {
	lines := bytes.Split(content, []byte("\n"))
	inCodeBlock := false

	// 找到 frontmatter 结束位置
	frontmatterEnd := FindFrontmatterEnd(lines)
	startLine := 0
	if frontmatterEnd >= 0 {
		startLine = frontmatterEnd + 1 // 从 frontmatter 之后开始搜索
	}

	for i := startLine; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])

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

// SectionTOC 表示一个章节的 TOC 信息
type SectionTOC struct {
	H1Line int    // H1 所在行号 (0-based)
	TOC    string // 该章节的 TOC 内容
}

// FindH1Lines 查找所有 H1 标题的行号 (0-based)
// 注意：会跳过 YAML frontmatter 内部的内容
func (h *MarkerHandler) FindH1Lines(content []byte) []int {
	lines := bytes.Split(content, []byte("\n"))
	inCodeBlock := false
	var h1Lines []int

	// 找到 frontmatter 结束位置
	frontmatterEnd := FindFrontmatterEnd(lines)
	startLine := 0
	if frontmatterEnd >= 0 {
		startLine = frontmatterEnd + 1 // 从 frontmatter 之后开始搜索
	}

	for i := startLine; i < len(lines); i++ {
		trimmed := bytes.TrimSpace(lines[i])

		// 检测代码块
		if bytes.HasPrefix(trimmed, []byte("```")) || bytes.HasPrefix(trimmed, []byte("~~~")) {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// 检测 H1 标题 (只匹配单个 #)
		if len(trimmed) > 1 && trimmed[0] == '#' && trimmed[1] != '#' {
			if trimmed[1] == ' ' || trimmed[1] == '\t' {
				h1Lines = append(h1Lines, i)
			}
		}
	}

	return h1Lines
}

// InsertSectionTOCs 在每个 H1 后插入对应章节的 TOC
// sectionTOCs 是一个按 H1Line 排序的切片
func (h *MarkerHandler) InsertSectionTOCs(content []byte, sectionTOCs []SectionTOC) []byte {
	if len(sectionTOCs) == 0 {
		return content
	}

	lines := bytes.Split(content, []byte("\n"))
	var result [][]byte

	// 创建 H1 行号到 TOC 的映射
	h1ToTOC := make(map[int]string)
	for _, st := range sectionTOCs {
		if st.TOC != "" {
			h1ToTOC[st.H1Line] = st.TOC
		}
	}

	// 标记需要跳过的行（H1 后的空行，因为 TOC 块会自己添加空行）
	skipLines := make(map[int]bool)
	for h1Line := range h1ToTOC {
		// 检查 H1 后面的行是否为空行
		if h1Line+1 < len(lines) && len(bytes.TrimSpace(lines[h1Line+1])) == 0 {
			skipLines[h1Line+1] = true
		}
	}

	for i, line := range lines {
		// 跳过 H1 后原有的空行（TOC 块会自己添加）
		if skipLines[i] {
			continue
		}

		result = append(result, line)

		// 检查是否需要在此行后插入 TOC
		if toc, ok := h1ToTOC[i]; ok {
			result = append(result, []byte(""))           // 空行（开始标记前）
			result = append(result, []byte(h.marker))    // <!--TOC-->
			result = append(result, []byte(""))           // 空行（开始标记后）
			result = append(result, []byte(toc))          // TOC 内容
			result = append(result, []byte(""))           // 空行（结束标记前）
			result = append(result, []byte(h.marker))    // <!--TOC-->
			result = append(result, []byte(""))           // 空行（结束标记后）
		}
	}

	return bytes.Join(result, []byte("\n"))
}

// UpdateSectionTOCs 更新现有的章节 TOC (替换每个 H1 后的 <!--TOC--> 区块)
// 返回更新后的内容
func (h *MarkerHandler) UpdateSectionTOCs(content []byte, sectionTOCs []SectionTOC) []byte {
	// 如果文件中没有 TOC 标记，使用插入模式
	markers := h.FindMarkers(content)
	if !markers.Found {
		return h.InsertSectionTOCs(content, sectionTOCs)
	}

	lines := bytes.Split(content, []byte("\n"))
	markerBytes := []byte(h.marker)

	// 找到所有现有的 TOC 区块 (成对的 <!--TOC-->)
	type tocBlock struct {
		startLine int // 开始标记行
		endLine   int // 结束标记行
	}
	var existingBlocks []tocBlock

	var pendingStart = -1
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.Equal(trimmed, markerBytes) {
			if pendingStart == -1 {
				pendingStart = i
			} else {
				existingBlocks = append(existingBlocks, tocBlock{pendingStart, i})
				pendingStart = -1
			}
		}
	}

	// 如果没有成对的标记，使用插入模式
	if len(existingBlocks) == 0 {
		return h.InsertSectionTOCs(content, sectionTOCs)
	}

	// 创建新内容，删除所有现有 TOC 块
	var cleanedLines [][]byte
	skipUntil := -1
	skipNextEmpty := false
	for i, line := range lines {
		if i <= skipUntil {
			continue
		}

		// 跳过 TOC 块结束后紧跟的空行，避免空行累积
		if skipNextEmpty {
			skipNextEmpty = false
			if len(bytes.TrimSpace(line)) == 0 {
				continue
			}
		}

		// 检查是否是某个 TOC 块的开始
		isBlockStart := false
		for _, block := range existingBlocks {
			if i == block.startLine {
				skipUntil = block.endLine
				skipNextEmpty = true // 标记：跳过结束标记后的空行
				isBlockStart = true
				break
			}
		}

		if !isBlockStart {
			cleanedLines = append(cleanedLines, line)
		}
	}

	// 在清理后的内容上重新插入章节 TOC
	cleanedContent := bytes.Join(cleanedLines, []byte("\n"))

	// 重新计算 H1 行号 (因为删除了旧 TOC 后行号可能变化)
	// 我们需要根据 H1 的文本内容来匹配
	return h.InsertSectionTOCs(cleanedContent, sectionTOCs)
}

// CleanTOCBlocks 删除所有 TOC 块，返回干净的内容
// 删除的内容包括：TOC 块本身 + 块前的一个空行 + 块后的一个空行
// 确保 H1 和 H2 之间只保留原始的一个空行（如果有）
func (h *MarkerHandler) CleanTOCBlocks(content []byte) ([]byte, []TOCBlockInfo) {
	lines := bytes.Split(content, []byte("\n"))
	markerBytes := []byte(h.marker)

	// 找到所有现有的 TOC 区块 (成对的 <!--TOC-->)
	type tocBlock struct {
		startLine int
		endLine   int
	}
	var existingBlocks []tocBlock

	var pendingStart = -1
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.Equal(trimmed, markerBytes) {
			if pendingStart == -1 {
				pendingStart = i
			} else {
				existingBlocks = append(existingBlocks, tocBlock{pendingStart, i})
				pendingStart = -1
			}
		}
	}

	// 如果没有 TOC 块，返回原内容
	if len(existingBlocks) == 0 {
		return content, nil
	}

	// 收集 TOC 块信息并标记要删除的行
	deleteLines := make(map[int]bool)
	var blockInfos []TOCBlockInfo

	for _, block := range existingBlocks {
		// 标记 TOC 块内所有行需要删除
		for i := block.startLine; i <= block.endLine; i++ {
			deleteLines[i] = true
		}

		// 检查块前是否有空行需要删除（开始标记前的空行）
		if block.startLine > 0 && len(bytes.TrimSpace(lines[block.startLine-1])) == 0 {
			deleteLines[block.startLine-1] = true
		}

		// 检查块后是否有空行需要删除（结束标记后的空行）
		// 删除结束标记后连续的所有空行，只保留一个
		afterEnd := block.endLine + 1
		emptyCount := 0
		for afterEnd < len(lines) && len(bytes.TrimSpace(lines[afterEnd])) == 0 {
			emptyCount++
			if emptyCount > 1 {
				// 只保留一个空行，多余的删除
				deleteLines[afterEnd] = true
			}
			afterEnd++
		}
		// 如果只有一个空行，也删除它（因为 InsertSectionTOCs 会添加）
		if emptyCount == 1 {
			deleteLines[block.endLine+1] = true
		}

		blockInfos = append(blockInfos, TOCBlockInfo{
			StartLine: block.startLine,
			EndLine:   block.endLine,
		})
	}

	// 构建干净的内容
	var cleanedLines [][]byte
	for i, line := range lines {
		if !deleteLines[i] {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return bytes.Join(cleanedLines, []byte("\n")), blockInfos
}

// TOCBlockInfo 记录 TOC 块的位置信息
type TOCBlockInfo struct {
	StartLine int // 开始行 (0-based)
	EndLine   int // 结束行 (0-based)
}

// CalcTOCBlockLines 计算插入一个 TOC 块会增加多少行
// 格式：空行 + <!--TOC--> + 空行 + TOC内容 + 空行 + <!--TOC--> + 空行
// 返回：6 + TOC内容行数
func CalcTOCBlockLines(tocContent string) int {
	if tocContent == "" {
		return 0
	}
	contentLines := strings.Count(tocContent, "\n") + 1
	return 6 + contentLines // 1(空行) + 1(开始标记) + 1(空行) + N(内容) + 1(空行) + 1(结束标记) + 1(空行)
}
