# GoStack CLI

> Modern Go Project Scaffolding — inspired by Vite.

[![CI](https://github.com/alifkhasan01/gostack-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/alifkhasan01/gostack-cli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alifkhasan01/gostack-cli)](https://goreportcard.com/report/github.com/alifkhasan01/gostack-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/alifkhasan01/gostack-cli)](https://github.com/alifkhasan01/gostack-cli/releases)

GoStack CLI membuat project Go backend siap pakai dalam hitungan detik. Pilih framework, arsitektur, database, dan ORM — CLI akan mengunduh starter template, mengganti placeholder, lalu menjalankan `go mod tidy` secara otomatis.

---

## Installation

```bash
go install github.com/alifkhasan01/gostack-cli/cmd/gostack@latest
```

Atau download binary langsung dari [Releases](https://github.com/alifkhasan01/gostack-cli/releases).

---

## Quick Start

```bash
# Interactive wizard
gostack create

# Non-interactive (CI-friendly)
gostack create my-api \
  --framework gin \
  --arch clean \
  --database postgres \
  --orm gorm \
  --auth jwt \
  --docker \
  --swagger
```

### Wizard

```
Project Name   › my-api
Module Name    › github.com/yourname/my-api

Framework      › Gin | Fiber | Echo | Chi
Architecture   › Standard | Clean | Hexagonal | DDD
Database       › PostgreSQL | MySQL | SQLite
ORM            › GORM | Bun | SQLX
Authentication › JWT | Session | None
Docker         › Yes / No
Swagger        › Yes / No
```

---

## Commands

### `gostack create`

Buat project baru dari starter template.

```bash
gostack create [project-name] [flags]

Flags:
  --module     Go module path (e.g. github.com/you/my-api)
  --framework  gin | fiber | echo | chi
  --arch       standard | clean | hexagonal | ddd
  --database   postgres | mysql | sqlite
  --orm        gorm | bun | sqlx
  --auth       jwt | session | none
  --docker     Include Docker files
  --swagger    Include Swagger setup
```

### `gostack generate`

Generate file di dalam project yang sudah ada.

```bash
# Single file
gostack generate handler  User
gostack generate service   User
gostack generate repository User
gostack generate migration create_users

# Semua sekaligus (handler + service + repository)
gostack generate module User

# CRUD lengkap (entity + repo + service + handler + migration + route injection)
gostack generate crud Product

# Alias singkat
gostack g crud Order
```

### `gostack update`

Self-update ke versi terbaru.

```bash
gostack update           # download dan install
gostack update --check   # cek saja tanpa install
```

### `gostack version`

```bash
gostack version
# GoStack CLI v0.1.0
#   Commit    : abc1234
#   Built     : 2025-01-01T00:00:00Z
#   Go        : go1.22.0
#   Platform  : linux/amd64
```

---

## How It Works

```
gostack create
     │
     ▼
Interactive Wizard
     │
     ▼
Download Template (github.com/gostack-templates/*)
     │  (fallback: built-in scaffold)
     ▼
Replace Placeholders ({{MODULE_NAME}}, {{PROJECT_NAME}})
     │
     ▼
git init  →  go mod tidy
     │
     ▼
Done ✅
```

---

## Generated Project Structure

```
my-api/
├── cmd/api/            # entry point
├── internal/
│   ├── config/         # env config
│   ├── database/       # db connection
│   ├── entity/         # domain models
│   ├── handler/        # HTTP handlers
│   ├── middleware/      # CORS, JWT
│   ├── repository/     # data access layer
│   ├── routes/         # route registration
│   └── service/        # business logic
├── migrations/         # SQL files
├── docs/               # Swagger output
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── gostack.json        # project metadata
└── go.mod
```

---

## Template Repository

GoStack tidak menyimpan template di dalam binary. Semua starter berada di repository terpisah di bawah organisasi [gostack-templates](https://github.com/gostack-templates).

```
gostack-templates/
├── template-gin-clean
├── template-gin-standard
├── template-fiber-clean
├── template-fiber-standard
├── template-echo-clean
└── template-chi-clean
```

Keuntungan:
- Template bisa diperbarui tanpa rilis CLI baru
- Bug pada starter cukup di-fix di repo template
- Komunitas bisa membuat template sendiri

Setiap template repo memiliki struktur:

```
template-gin-clean/
├── template/       # file yang akan di-copy ke project
├── manifest.json   # metadata template
├── version.json
└── README.md
```

---

## Roadmap

| Version | Features |
|---------|----------|
| **v0.1** ✅ | Create project, remote template, placeholder replacement, built-in scaffold |
| **v0.2** ✅ | Generate handler/service/repository/migration/module, CRUD generator, route injection |
| **v0.3** 🔜 | Swagger integration, gRPC starter, GraphQL starter |
| **v0.4** 🔜 | Worker starter, background job scaffold |
| **v1.0** 🔜 | Plugin system, community templates, template marketplace |

---

## Philosophy

GoStack mengikuti filosofi yang sama seperti Vite:

- CLI tetap kecil dan ringan
- Starter project berada di repository terpisah
- Template dapat diperbarui kapan saja tanpa update CLI
- Framework baru dapat ditambahkan tanpa mengubah kode CLI

---

## Contributing

Lihat [CONTRIBUTING.md](CONTRIBUTING.md) untuk panduan kontribusi.

## License

[MIT](LICENSE) © 2025 GoStack
