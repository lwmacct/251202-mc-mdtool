package mdtoc_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lwmacct/251202-mc-mdtool/internal/mdtoc"
)

// TestTOC_GenerateFromContent 测试从内容生成 TOC 的核心功能
func TestTOC_GenerateFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		opts     mdtoc.Options
		contains []string // TOC 应包含的内容
		excludes []string // TOC 不应包含的内容
	}{
		{
			name: "basic markdown",
			content: `# 项目简介

## 功能特性

### 特性一

### 特性二

## 安装说明

## 使用方法`,
			opts: mdtoc.Options{MinLevel: 1, MaxLevel: 3},
			contains: []string{
				"[项目简介](#项目简介)",
				"[功能特性](#功能特性)",
				"[特性一](#特性一)",
				"[特性二](#特性二)",
				"[安装说明](#安装说明)",
				"[使用方法](#使用方法)",
			},
		},
		{
			name: "filter by level",
			content: `# Title
## Section 1
### Subsection
#### Deep
## Section 2`,
			opts:     mdtoc.Options{MinLevel: 2, MaxLevel: 2},
			contains: []string{"[Section 1]", "[Section 2]"},
			excludes: []string{"[Title]", "[Subsection]", "[Deep]"},
		},
		{
			name: "ordered list",
			content: `# Title
## Section 1
## Section 2`,
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2, Ordered: true},
			contains: []string{"1.", "2."},
			excludes: []string{"- ["},
		},
		{
			name: "with line numbers",
			content: `# Title

## Section 1

Content here

## Section 2`,
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2, LineNumber: true},
			contains: []string{"`:", "-"},
		},
		{
			name: "code block headers ignored",
			content: "# Real Header\n```markdown\n# Fake Header\n```\n## Another Real",
			opts:     mdtoc.Options{MinLevel: 1, MaxLevel: 2},
			contains: []string{"[Real Header]", "[Another Real]"},
			excludes: []string{"[Fake Header]"},
		},
		{
			name: "duplicate header handling",
			content: `# API
## GET
## POST
## GET
## GET`,
			opts: mdtoc.Options{MinLevel: 1, MaxLevel: 2},
			contains: []string{
				"(#get)",
				"(#get-1)",
				"(#get-2)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := mdtoc.New(tt.opts)
			got, err := toc.GenerateFromContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("GenerateFromContent() error = %v", err)
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("TOC should contain %q, got:\n%s", s, got)
				}
			}

			for _, s := range tt.excludes {
				if strings.Contains(got, s) {
					t.Errorf("TOC should NOT contain %q, got:\n%s", s, got)
				}
			}
		})
	}
}

// TestTOC_SectionMode 测试章节 TOC 模式
func TestTOC_SectionMode(t *testing.T) {
	content := `# 第一章

## 1.1 概述

## 1.2 详情

# 第二章

## 2.1 功能

### 2.1.1 子功能

# 第三章

只有介绍，没有子标题。
`
	toc := mdtoc.New(mdtoc.Options{
		MinLevel:   2,
		MaxLevel:   3,
		SectionTOC: true,
	})

	preview, err := toc.GenerateSectionTOCsPreview([]byte(content))
	if err != nil {
		t.Fatalf("GenerateSectionTOCsPreview() error = %v", err)
	}

	// 验证章节标题存在
	if !strings.Contains(preview, "第一章") {
		t.Error("Preview should contain chapter 1 title")
	}
	if !strings.Contains(preview, "第二章") {
		t.Error("Preview should contain chapter 2 title")
	}

	// 验证子标题链接
	if !strings.Contains(preview, "[1.1 概述]") {
		t.Error("Preview should contain section 1.1 link")
	}
	if !strings.Contains(preview, "[2.1.1 子功能]") {
		t.Error("Preview should contain subsection 2.1.1 link")
	}

	// 第三章无子标题，不应出现在预览中
	if strings.Contains(preview, "第三章") {
		t.Error("Preview should NOT contain chapter 3 (no sub-headers)")
	}
}

// TestTOC_EmptyAndEdgeCases 测试边缘情况
func TestTOC_EmptyAndEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		opts        mdtoc.Options
		expectEmpty bool
	}{
		{
			name:        "empty content",
			content:     "",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "no headers",
			content:     "Just some text\n\nMore text",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "only code blocks",
			content:     "```\n# Not a header\n## Also not\n```",
			opts:        mdtoc.DefaultOptions(),
			expectEmpty: true,
		},
		{
			name:        "headers below min level",
			content:     "#### Deep header\n##### Deeper",
			opts:        mdtoc.Options{MinLevel: 1, MaxLevel: 2},
			expectEmpty: true,
		},
		{
			name:        "headers above max level",
			content:     "# Title\n## Section",
			opts:        mdtoc.Options{MinLevel: 3, MaxLevel: 6},
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := mdtoc.New(tt.opts)
			got, err := toc.GenerateFromContent([]byte(tt.content))
			if err != nil {
				t.Fatalf("GenerateFromContent() error = %v", err)
			}

			isEmpty := got == ""
			if isEmpty != tt.expectEmpty {
				if tt.expectEmpty {
					t.Errorf("Expected empty TOC, got: %q", got)
				} else {
					t.Error("Expected non-empty TOC, got empty")
				}
			}
		})
	}
}

// TestTOC_FileOperations 测试文件操作功能
func TestTOC_FileOperations(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	t.Run("GenerateFromFile", func(t *testing.T) {
		content := "# Title\n## Section 1\n## Section 2"
		filePath := filepath.Join(tmpDir, "test.md")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		toc := mdtoc.New(mdtoc.DefaultOptions())
		got, err := toc.GenerateFromFile(filePath)
		if err != nil {
			t.Fatalf("GenerateFromFile() error = %v", err)
		}

		if !strings.Contains(got, "[Title]") {
			t.Error("TOC should contain Title link")
		}
	})

	t.Run("GenerateFromFile_NotExists", func(t *testing.T) {
		toc := mdtoc.New(mdtoc.DefaultOptions())
		_, err := toc.GenerateFromFile(filepath.Join(tmpDir, "nonexistent.md"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("HasMarker", func(t *testing.T) {
		// 有标记的文件
		withMarker := "# Title\n<!--TOC-->\n## Section"
		withPath := filepath.Join(tmpDir, "with_marker.md")
		os.WriteFile(withPath, []byte(withMarker), 0644)

		// 无标记的文件
		withoutMarker := "# Title\n## Section"
		withoutPath := filepath.Join(tmpDir, "without_marker.md")
		os.WriteFile(withoutPath, []byte(withoutMarker), 0644)

		toc := mdtoc.New(mdtoc.DefaultOptions())

		has, err := toc.HasMarker(withPath)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Error("HasMarker() should return true for file with marker")
		}

		has, err = toc.HasMarker(withoutPath)
		if err != nil {
			t.Fatal(err)
		}
		if has {
			t.Error("HasMarker() should return false for file without marker")
		}
	})

	t.Run("UpdateFile_WithMarker", func(t *testing.T) {
		content := `# Title

<!--TOC-->

Old TOC content

<!--TOC-->

## Section 1

## Section 2
`
		filePath := filepath.Join(tmpDir, "update_with_marker.md")
		os.WriteFile(filePath, []byte(content), 0644)

		toc := mdtoc.New(mdtoc.DefaultOptions())
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath)
		updatedStr := string(updated)

		if !strings.Contains(updatedStr, "[Section 1]") {
			t.Error("Updated file should contain Section 1 link")
		}
		if strings.Contains(updatedStr, "Old TOC content") {
			t.Error("Updated file should NOT contain old TOC content")
		}
	})

	t.Run("UpdateFile_WithoutMarker", func(t *testing.T) {
		content := `# Title

## Section 1

## Section 2
`
		filePath := filepath.Join(tmpDir, "update_without_marker.md")
		os.WriteFile(filePath, []byte(content), 0644)

		toc := mdtoc.New(mdtoc.DefaultOptions())
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath)
		updatedStr := string(updated)

		// 应该自动插入 TOC 标记
		if !strings.Contains(updatedStr, "<!--TOC-->") {
			t.Error("Updated file should contain TOC markers")
		}
		if !strings.Contains(updatedStr, "[Section 1]") {
			t.Error("Updated file should contain Section 1 link")
		}
	})

	t.Run("UpdateFile_SectionMode", func(t *testing.T) {
		content := `# Chapter 1

## Section 1.1

# Chapter 2

## Section 2.1
`
		filePath := filepath.Join(tmpDir, "update_section_mode.md")
		os.WriteFile(filePath, []byte(content), 0644)

		toc := mdtoc.New(mdtoc.Options{
			MinLevel:   2,
			MaxLevel:   3,
			SectionTOC: true,
		})
		if err := toc.UpdateFile(filePath); err != nil {
			t.Fatal(err)
		}

		updated, _ := os.ReadFile(filePath)
		updatedStr := string(updated)

		// 每个章节后应该有独立的 TOC
		if strings.Count(updatedStr, "<!--TOC-->") < 4 {
			t.Error("Section mode should insert multiple TOC marker pairs")
		}
		if !strings.Contains(updatedStr, "[Section 1.1]") {
			t.Error("Updated file should contain Section 1.1 link")
		}
		if !strings.Contains(updatedStr, "[Section 2.1]") {
			t.Error("Updated file should contain Section 2.1 link")
		}
	})

	t.Run("CheckDiff_NoDiff", func(t *testing.T) {
		// 创建一个 TOC 已经是最新的文件
		toc := mdtoc.New(mdtoc.DefaultOptions())

		content := `# Title

<!--TOC-->

- [Title](#title)
  - [Section 1](#section-1)

<!--TOC-->

## Section 1
`
		filePath := filepath.Join(tmpDir, "no_diff.md")
		os.WriteFile(filePath, []byte(content), 0644)

		hasDiff, err := toc.CheckDiff(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if hasDiff {
			t.Error("CheckDiff() should return false when TOC is up to date")
		}
	})

	t.Run("CheckDiff_HasDiff", func(t *testing.T) {
		toc := mdtoc.New(mdtoc.DefaultOptions())

		content := `# Title

<!--TOC-->

- [Old Section](#old-section)

<!--TOC-->

## Section 1
`
		filePath := filepath.Join(tmpDir, "has_diff.md")
		os.WriteFile(filePath, []byte(content), 0644)

		hasDiff, err := toc.CheckDiff(filePath)
		if err != nil {
			t.Fatal(err)
		}
		if !hasDiff {
			t.Error("CheckDiff() should return true when TOC needs update")
		}
	})
}

// TestTOC_ChineseContent 测试中文内容处理
func TestTOC_ChineseContent(t *testing.T) {
	content := `# 项目介绍

## 功能特性

### 高性能

### 易扩展

## 快速开始

### 安装

### 配置

## 常见问题
`
	toc := mdtoc.New(mdtoc.DefaultOptions())
	got, err := toc.GenerateFromContent([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	// 验证中文标题保留
	expectedLinks := []string{
		"[项目介绍](#项目介绍)",
		"[功能特性](#功能特性)",
		"[高性能](#高性能)",
		"[易扩展](#易扩展)",
		"[快速开始](#快速开始)",
		"[安装](#安装)",
		"[配置](#配置)",
		"[常见问题](#常见问题)",
	}

	for _, link := range expectedLinks {
		if !strings.Contains(got, link) {
			t.Errorf("TOC should contain %q", link)
		}
	}
}

// TestTOC_SpecialCharacters 测试特殊字符处理
func TestTOC_SpecialCharacters(t *testing.T) {
	content := `# Hello, World!

## What's New?

## C++ Guide

## Node.js & npm

## 100% Complete

## Version 2.0.0
`
	toc := mdtoc.New(mdtoc.DefaultOptions())
	got, err := toc.GenerateFromContent([]byte(content))
	if err != nil {
		t.Fatal(err)
	}

	// 验证特殊字符被正确处理
	if !strings.Contains(got, "[Hello, World!]") {
		t.Error("Title with comma and exclamation should be preserved")
	}
	if !strings.Contains(got, "[What's New?]") {
		t.Error("Title with apostrophe and question mark should be preserved")
	}
	if !strings.Contains(got, "[C++ Guide]") {
		t.Error("Title with plus signs should be preserved")
	}
}

// TestDefaultOptions 测试默认配置
func TestDefaultOptions(t *testing.T) {
	opts := mdtoc.DefaultOptions()

	if opts.MinLevel != 1 {
		t.Errorf("DefaultOptions().MinLevel = %d, want 1", opts.MinLevel)
	}
	if opts.MaxLevel != 3 {
		t.Errorf("DefaultOptions().MaxLevel = %d, want 3", opts.MaxLevel)
	}
	if opts.Ordered {
		t.Error("DefaultOptions().Ordered should be false")
	}
	if opts.LineNumber {
		t.Error("DefaultOptions().LineNumber should be false")
	}
	if opts.SectionTOC {
		t.Error("DefaultOptions().SectionTOC should be false")
	}
}

// BenchmarkTOC_Generate 性能基准测试
func BenchmarkTOC_Generate(b *testing.B) {
	// 构造一个较大的文档
	var sb strings.Builder
	for i := 0; i < 100; i++ {
		sb.WriteString("# Chapter ")
		sb.WriteString(string(rune('A' + i%26)))
		sb.WriteString("\n\n")
		for j := 0; j < 10; j++ {
			sb.WriteString("## Section ")
			sb.WriteString(string(rune('0' + j)))
			sb.WriteString("\n\nContent here...\n\n")
		}
	}
	content := []byte(sb.String())

	toc := mdtoc.New(mdtoc.DefaultOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = toc.GenerateFromContent(content)
	}
}
