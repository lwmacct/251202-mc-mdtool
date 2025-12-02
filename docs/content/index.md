---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "mc-mdtool"
  text: "Markdown CLI 工具集"
  tagline: 目录生成、格式化、检查等功能
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: 设计文档
      link: /design/cmd-toc

features:
  - title: toc - 目录生成
    details: 生成 GitHub 风格的 Table of Contents，支持原地更新
  - title: fmt - 格式化
    details: 参考 Prettier 设计，统一 Markdown 代码风格
  - title: lint - 规范检查
    details: 检查 Markdown 文件是否符合规范
  - title: links - 链接检查
    details: 检查文档中的链接是否有效
---

<!--@include: ./readme.md-->
