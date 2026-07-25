# Akari

<p align="left">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License" />
</p>

Another Minecraft Yggdrasil API-Compatible Server

## Build Environment

- **Node.js**: `^22.18.0` or `>=24.12.0`
- **Go**: `1.26`
- **Package Manager**: pnpm

## Build

```sh
pnpm install         # Install frontend dependencies
pnpm dev             # Start frontend dev server
make dev             # Run both frontend and Go backend concurrently
make build           # Build frontend and Go server binary
make run             # Run the Go server
pnpm lint            # Lint frontend
make lint            # Lint all (frontend + Go)
```

## License

[MIT](LICENSE)