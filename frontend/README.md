# VDA + IM Frontend

pnpm monorepo with web (Vite + React) and mobile (Expo + React Native) apps, sharing protocol, UI components, config, and types packages.

## Structure

```
apps/
  web/          @game/web      - Web SPA (Vite, React Router, Tailwind)
  mobile/       @game/mobile   - React Native app (Expo)
packages/
  api/          @game/api      - IM protocol codec + HTTP API clients
  config/       @game/config   - Service configuration types
  shared/       @game/shared   - Shared types and mock data
  ui/           @game/ui       - UI component library (React)
```

## Quick Start

```bash
pnpm install
pnpm dev          # start web dev server
pnpm typecheck    # type-check all packages
pnpm lint         # lint all packages
pnpm test         # run unit tests
```

### Web App

```bash
cd apps/web
pnpm dev                  # http://localhost:5173
pnpm test:e2e             # Playwright E2E tests
```

### Mobile App

```bash
cd apps/mobile
pnpm dev                  # Expo dev server
pnpm ios                  # iOS simulator
pnpm android              # Android emulator
```

## Tech Stack

| Area | Choice |
|------|--------|
| State | Zustand |
| Routing (web) | React Router v7 |
| Styling (web) | Tailwind CSS v4 |
| HTTP | fetch wrapper (packages/api) |
| IM | Binary WebSocket protocol |
| Voice | LiveKit WebRTC |
| Test | Vitest + React Testing Library + Playwright |

## IM Protocol

WebSocket binary frames: 16-byte header (`totalLen | headerLen | version | op | seq`) followed by JSON payload.

| Opcode | Name | Direction |
|--------|------|-----------|
| 2 | Heartbeat | C → S |
| 3 | HeartbeatReply | S → C |
| 4 | Send | C ↔ S |
| 7 | Auth | C → S |
| 8 | AuthReply | S → C |

See `packages/api/README.md` for protocol details.
