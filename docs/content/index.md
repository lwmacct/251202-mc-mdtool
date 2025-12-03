---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "mc-mdtoc"
  text: "Markdown TOC 生成工具"
  tagline: 为 Markdown 文件自动生成符合规范的目录
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: 设计文档
      link: /design/cmd-toc

features:
  - title: GitHub 风格锚点
    details: 生成符合 GitHub 规范的 anchor link，支持中英文混合标题
  - title: 章节模式
    details: 每个 H1 后生成独立子目录，适合长文档组织
  - title: Frontmatter 支持
    details: 自动跳过 YAML frontmatter，兼容 VitePress、Hugo 等
  - title: 批量处理
    details: 支持多文件和管道输入，轻松集成到 CI/CD 流程
---

<!--@include: ./readme.md-->
