# Wavelet Frontend

Modern frontend application for the Wavelet scaffold framework.

[中文](./README_zh.md) | English

## Tech Stack

- **Framework**: [Next.js 16](https://nextjs.org/)
- **Styling**: [Tailwind CSS 4](https://tailwindcss.com/)
- **UI Components**: [Radix UI](https://www.radix-ui.com/)
- **Icons**: [Lucide React](https://lucide.dev/)
- **Package Manager**: [Bun](https://bun.sh/)

## Getting Started

### Prerequisites

- Bun >= 1.2

### Installation

1. Install dependencies:

   ```bash
   bun install
   ```

2. Run the development server:

   ```bash
   bun dev
   ```

   Open [http://localhost:3000](http://localhost:3000) (or the port shown in your terminal) with your browser to see the result.

## Project Structure

- `app/`: Next.js App Router pages and layouts
- `components/`: Reusable UI components
  - `ui/`: Base UI components (buttons, inputs, etc.)
  - `common/`: Shared business components
- `lib/`: Utility functions and service definitions
- `public/`: Static assets

## Scripts

- `bun dev`: Start development server with Turbopack
- `bun run build`: Build the application for production
- `bun run build:embed`: Build with static export for embedding into the Go binary
- `bun start`: Start production server
- `bun run lint`: Run ESLint
- `bun run format`: Format code with Biome
