# web-v2

Bedrock 2.0 frontend (Vue 3 + Veltra + CatKit), built with [Vite+](https://viteplus.dev/guide) (`vp`).

## Prerequisites

Install the global `vp` CLI: https://viteplus.dev/guide

## Setup

```sh
vp install
```

### Develop

```sh
vp dev
```

Dev server proxies `/api` and `/ws` to the Go backend on `:8080` (see `vite.config.ts`).

### Check (format + lint + typecheck)

```sh
vp check
```

### Build

```sh
vp build
```

From the repo root, `make build-frontend` / `make build` run `vp install && vp build` and embed `dist/` into the Go binary.
