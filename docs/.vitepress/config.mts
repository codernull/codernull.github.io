import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'codernull',
  description: '结论型技术文章：分布式系统 · 低延迟 · Go 实战 · 工程方法',
  lang: 'zh-CN',
  cleanUrls: true,
  lastUpdated: true,

  themeConfig: {
    nav: [
      { text: '首页', link: '/' },
      { text: '分布式系统', link: '/分布式系统/' },
      { text: '低延迟系统', link: '/低延迟系统/' },
      { text: 'Go 实战', link: '/Go实战/' },
      { text: '工程方法', link: '/工程方法/' },
      { text: 'GitHub', link: 'https://github.com/codernull/codernull.github.io' },
    ],

    sidebar: {
      '/分布式系统/': [
        {
          text: '分布式系统',
          items: [
            { text: '概览', link: '/分布式系统/' },
            {
              text: '四种语言 Raft 实现横向对比',
              link: '/分布式系统/raft-实现对比',
            },
          ],
        },
      ],
      '/低延迟系统/': [
        {
          text: '低延迟系统',
          items: [{ text: '概览', link: '/低延迟系统/' }],
        },
      ],
      '/Go实战/': [
        {
          text: 'Go 实战',
          items: [{ text: '概览', link: '/Go实战/' }],
        },
      ],
      '/工程方法/': [
        {
          text: '工程方法',
          items: [{ text: '概览', link: '/工程方法/' }],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/codernull' },
    ],

    search: {
      provider: 'local',
    },

    outline: {
      label: '本页目录',
      level: [2, 3],
    },

    docFooter: {
      prev: '上一篇',
      next: '下一篇',
    },

    lastUpdated: {
      text: '最后更新',
    },

    footer: {
      message: '结论型技术文章 · 踩坑之后的判断',
      copyright: 'Copyright © 2026 codernull',
    },
  },
})
