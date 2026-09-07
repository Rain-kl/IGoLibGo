# Wavelet Frontend

Wavelet 脚手架系统的现代化前端应用。

[English](./README.md) | 中文

## 技术栈

- **框架**: [Next.js 16](https://nextjs.org/)
- **样式**: [Tailwind CSS 4](https://tailwindcss.com/)
- **UI 组件**: [Radix UI](https://www.radix-ui.com/)
- **图标**: [Lucide React](https://lucide.dev/)
- **包管理器**: [Bun](https://bun.sh/)

## 快速开始

### 前置要求

- Bun >= 1.2

### 安装

1. 安装依赖:

   ```bash
   bun install
   ```

2. 运行开发服务器:

   ```bash
   bun dev
   ```

   在浏览器中打开 [http://localhost:3000](http://localhost:3000) (或终端中显示的端口) 查看结果。

## 项目结构

- `app/`: Next.js App Router 页面和布局
- `components/`: 可复用 UI 组件
  - `ui/`: 基础 UI 组件 (按钮, 输入框等)
  - `common/`: 共享业务组件
- `lib/`: 工具函数和服务定义
- `public/`: 静态资源

## 脚本

- `bun dev`: 启动开发服务器 (使用 Turbopack)
- `bun run build`: 构建生产版本
- `bun start`: 启动生产服务器
- `bun run lint`: 运行 ESLint
