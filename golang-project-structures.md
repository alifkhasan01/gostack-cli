# 📦 Struktur Project Golang — Semua Bidang

> Panduan lengkap struktur direktori project Go berdasarkan bidang/domain.  
> Setiap bidang memiliki konvensi dan kebutuhan yang berbeda.

---

## Daftar Isi

1. [REST API / Backend Service](#1-rest-api--backend-service)
2. [CLI Tool](#2-cli-tool)
3. [Microservices](#3-microservices)
4. [gRPC Service](#4-grpc-service)
5. [Full-Stack Web App](#5-full-stack-web-app)
6. [Library / Package](#6-library--package)
7. [Worker / Background Job](#7-worker--background-job)
8. [GraphQL API](#8-graphql-api)
9. [Monorepo / Multi-Module](#9-monorepo--multi-module)
10. [Wayland Compositor / System Tool (CGo)](#10-wayland-compositor--system-tool-cgo)

---

## 1. REST API / Backend Service

Cocok untuk: API server, CRUD app, backend service dengan HTTP.

```
myapp/
├── cmd/
│   └── server/
│       └── main.go             # Entry point
├── internal/
│   ├── handler/                # HTTP handler (controller)
│   │   ├── user.go
│   │   └── product.go
│   ├── service/                # Business logic
│   │   ├── user.go
│   │   └── product.go
│   ├── repository/             # Data access layer (DB)
│   │   ├── user.go
│   │   └── product.go
│   ├── model/                  # Struct / domain model
│   │   ├── user.go
│   │   └── product.go
│   ├── middleware/             # Auth, logging, CORS, dll
│   │   ├── auth.go
│   │   └── logger.go
│   └── config/                 # Konfigurasi app
│       └── config.go
├── pkg/
│   ├── response/               # Format response JSON
│   │   └── response.go
│   ├── validator/              # Input validation helper
│   │   └── validator.go
│   └── jwt/                    # JWT helper
│       └── jwt.go
├── migrations/                 # SQL migration files
│   ├── 001_create_users.sql
│   └── 002_create_products.sql
├── docs/                       # Swagger / API docs
│   └── swagger.yaml
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

**Key points:**
- `internal/` — kode yang tidak boleh diimport package lain di luar project
- `pkg/` — kode reusable yang boleh diimport dari luar
- Pisahkan `handler → service → repository` untuk clean architecture

---

## 2. CLI Tool

Cocok untuk: developer tools, automation script, terminal utilities.

```
mycli/
├── cmd/
│   ├── root.go                 # Root command (cobra)
│   ├── init.go                 # Subcommand: mycli init
│   ├── build.go                # Subcommand: mycli build
│   └── run.go                  # Subcommand: mycli run
├── internal/
│   ├── config/
│   │   └── config.go           # Load config file (~/.myclirc)
│   ├── runner/
│   │   └── runner.go           # Core logic eksekusi
│   └── output/
│       └── printer.go          # Colored terminal output
├── pkg/
│   └── util/
│       └── file.go             # File utilities
├── main.go                     # Entry point (panggil cmd/root.go)
├── .goreleaser.yml             # Config untuk build & release
├── Makefile
├── go.mod
└── go.sum
```

**Key points:**
- Gunakan [Cobra](https://github.com/spf13/cobra) untuk subcommand
- Gunakan [Viper](https://github.com/spf13/viper) untuk config management
- `main.go` di root, delegasi ke `cmd/root.go`

---

## 3. Microservices

Cocok untuk: sistem terdistribusi, multiple service independen.

```
microservices/
├── services/
│   ├── user-service/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   ├── proto/              # Protobuf definitions
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   ├── order-service/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   ├── proto/
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   └── notification-service/
│       ├── cmd/main.go
│       ├── internal/
│       ├── Dockerfile
│       └── go.mod
│
├── shared/                     # Shared library (go module tersendiri)
│   ├── proto/                  # Shared protobuf
│   ├── middleware/
│   ├── logger/
│   └── go.mod
│
├── gateway/                    # API Gateway
│   ├── cmd/main.go
│   ├── internal/
│   │   └── proxy/
│   └── go.mod
│
├── infra/
│   ├── docker-compose.yml
│   ├── k8s/                    # Kubernetes manifests
│   │   ├── user-deployment.yaml
│   │   └── order-deployment.yaml
│   └── nginx/
│       └── nginx.conf
│
└── Makefile                    # make run-all, make build-all, dll
```

**Key points:**
- Setiap service punya `go.mod` sendiri
- Komunikasi antar service via gRPC atau message broker (NATS, Kafka)
- Shared code diletakkan di `shared/` sebagai module terpisah

---

## 4. gRPC Service

Cocok untuk: komunikasi antar service, high-performance API, streaming.

```
grpc-service/
├── cmd/
│   ├── server/
│   │   └── main.go             # gRPC server entry point
│   └── client/
│       └── main.go             # gRPC client contoh
├── internal/
│   ├── server/
│   │   └── server.go           # Implementasi gRPC server
│   ├── service/
│   │   └── user_service.go     # Business logic
│   └── repository/
│       └── user_repo.go
├── proto/
│   └── user/
│       ├── user.proto           # Definisi protobuf
│       └── user.pb.go          # Generated (jangan edit manual)
│       └── user_grpc.pb.go     # Generated gRPC code
├── pkg/
│   └── interceptor/            # gRPC interceptors (auth, logging)
│       ├── auth.go
│       └── logger.go
├── buf.gen.yaml                # Buf tool config (protobuf generator)
├── buf.yaml
├── Makefile                    # make proto, make run
├── go.mod
└── go.sum
```

**Key points:**
- Generate code dari `.proto` dengan `protoc` atau `buf`
- Pisahkan `server/` (wiring) dan `service/` (logic)
- Interceptor = middleware untuk gRPC

---

## 5. Full-Stack Web App

Cocok untuk: web app dengan server-side rendering atau embedded frontend.

```
webapp/
├── cmd/
│   └── web/
│       └── main.go
├── internal/
│   ├── handler/                # HTTP handler
│   ├── service/
│   ├── repository/
│   └── template/               # Template rendering logic
├── web/
│   ├── static/                 # CSS, JS, images
│   │   ├── css/
│   │   ├── js/
│   │   └── img/
│   └── templates/              # HTML templates
│       ├── layout/
│       │   └── base.html
│       ├── pages/
│       │   ├── home.html
│       │   └── dashboard.html
│       └── components/
│           └── navbar.html
├── migrations/
├── .env.example
├── Dockerfile
├── go.mod
└── go.sum
```

**Key points:**
- `web/` berisi semua aset frontend
- Gunakan `embed.FS` (Go 1.16+) untuk embed static files ke binary
- Template menggunakan `html/template` standar library Go

---

## 6. Library / Package

Cocok untuk: bikin library yang bisa diimport orang lain.

```
mylib/
├── mylib.go                    # Public API utama
├── mylib_test.go               # Unit test
├── option.go                   # Options pattern
├── option_test.go
├── internal/
│   └── core/                   # Implementasi internal
│       └── engine.go
├── examples/
│   ├── basic/
│   │   └── main.go             # Contoh pemakaian basic
│   └── advanced/
│       └── main.go
├── testdata/                   # Test fixtures
│   └── sample.json
├── CHANGELOG.md
├── LICENSE
├── README.md
├── go.mod
└── go.sum
```

**Key points:**
- Semua public API ada di root package (bukan `internal/`)
- Selalu sertakan `examples/` agar mudah dipahami user
- Ikuti semantic versioning (`v1.x.x`, `v2.x.x`)
- Gunakan `internal/` untuk implementasi yang tidak perlu diekspos

---

## 7. Worker / Background Job

Cocok untuk: job queue, cron job, event consumer (Kafka, RabbitMQ, NATS).

```
worker/
├── cmd/
│   └── worker/
│       └── main.go
├── internal/
│   ├── worker/
│   │   ├── worker.go           # Worker pool / orchestrator
│   │   ├── email_worker.go     # Worker spesifik task
│   │   └── report_worker.go
│   ├── job/
│   │   ├── job.go              # Job definition / interface
│   │   └── scheduler.go        # Cron scheduler
│   ├── queue/
│   │   ├── consumer.go         # Message consumer
│   │   └── producer.go         # Message producer
│   └── handler/
│       └── webhook.go          # HTTP endpoint untuk trigger job
├── pkg/
│   └── retry/                  # Retry logic dengan backoff
│       └── retry.go
├── config/
│   └── config.go
├── Dockerfile
├── go.mod
└── go.sum
```

**Key points:**
- Gunakan goroutine + channel untuk concurrency
- Implementasi graceful shutdown (context cancellation)
- Pisahkan `job/` (definisi task) dari `worker/` (eksekutor)

---

## 8. GraphQL API

Cocok untuk: API fleksibel, query client-driven.

```
graphql-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── resolver/               # GraphQL resolvers
│   │   ├── query.go
│   │   ├── mutation.go
│   │   └── subscription.go
│   ├── model/                  # Generated + custom models
│   │   └── models_gen.go       # Generated by gqlgen
│   ├── service/
│   │   └── user_service.go
│   ├── repository/
│   │   └── user_repo.go
│   └── loader/                 # DataLoader (N+1 problem)
│       └── user_loader.go
├── graph/
│   ├── schema.graphqls         # Schema definition
│   └── schema.resolvers.go     # Implementasi resolver
├── gqlgen.yml                  # Config gqlgen
├── pkg/
│   └── middleware/
│       └── auth.go
├── go.mod
└── go.sum
```

**Key points:**
- Gunakan [gqlgen](https://github.com/99designs/gqlgen) untuk code generation dari schema
- `DataLoader` wajib untuk menghindari N+1 query
- Schema-first approach: tulis `.graphqls` dulu, baru generate code

---

## 9. Monorepo / Multi-Module

Cocok untuk: project besar dengan banyak komponen terkait.

```
monorepo/
├── apps/
│   ├── api/                    # REST API service
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   └── go.mod
│   ├── worker/                 # Background worker
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   └── go.mod
│   └── cli/                    # CLI tool
│       ├── main.go
│       └── go.mod
│
├── packages/                   # Shared internal packages
│   ├── database/
│   │   ├── db.go
│   │   └── go.mod
│   ├── logger/
│   │   ├── logger.go
│   │   └── go.mod
│   ├── config/
│   │   ├── config.go
│   │   └── go.mod
│   └── auth/
│       ├── jwt.go
│       └── go.mod
│
├── proto/                      # Shared protobuf definitions
│   └── v1/
│       └── user.proto
│
├── scripts/
│   ├── build.sh
│   └── test.sh
│
├── go.work                     # Go workspace file
├── go.work.sum
├── Makefile
└── .github/
    └── workflows/
        └── ci.yml
```

**Key points:**
- Gunakan `go.work` (Go 1.18+) untuk Go Workspace — semua module bisa saling import lokal
- `apps/` = executable services, `packages/` = shared libraries
- Satu CI/CD pipeline bisa build semua apps sekaligus

---

## 10. Wayland Compositor / System Tool (CGo)

Cocok untuk: system-level tool, binding ke C library, Wayland compositor.

```
wm-persone/                     # Contoh: custom Wayland compositor
├── cmd/
│   └── wm-persone/
│       └── main.go             # Entry point compositor
├── internal/
│   ├── compositor/             # Core compositor logic
│   │   ├── compositor.go
│   │   ├── output.go           # Monitor/output management
│   │   └── seat.go             # Input (keyboard, mouse)
│   ├── layout/                 # Tiling layout engine
│   │   ├── layout.go
│   │   ├── tiling.go
│   │   └── floating.go
│   ├── config/                 # Lua config loader
│   │   ├── config.go
│   │   └── lua.go              # Lua binding (gopher-lua)
│   ├── ipc/                    # IPC socket
│   │   └── socket.go
│   └── xwayland/               # XWayland support
│       └── xwayland.go
├── pkg/
│   ├── log/
│   │   └── log.go
│   └── event/                  # Event bus internal
│       └── bus.go
├── bindings/                   # CGo bindings ke C library
│   ├── wlroots/
│   │   ├── wlroots.go          # Go wrapper
│   │   └── wlroots.h           # C header (copy dari sistem)
│   └── xkb/
│       └── xkb.go
├── config/
│   └── default.lua             # Default config file
├── docs/
│   ├── architecture.md
│   └── config-reference.md
├── scripts/
│   └── install.sh
├── Makefile                    # make build, make install, make run
├── go.mod
└── go.sum
```

**Key points:**
- CGo digunakan untuk binding ke `wlroots`, `libxkbcommon`, dsb
- Pisahkan `bindings/` dari `internal/` agar wrapping C jelas
- Lua config dimuat runtime, bukan dikompilasi
- `Makefile` sangat penting untuk manage build flags CGo

---

## Konvensi Umum Go (Berlaku di Semua Bidang)

| Direktori | Fungsi |
|-----------|--------|
| `cmd/` | Entry point executable, boleh lebih dari satu |
| `internal/` | Kode privat project, tidak bisa diimport dari luar |
| `pkg/` | Kode reusable yang boleh diimport dari luar |
| `api/` | OpenAPI/Swagger spec, protobuf definitions |
| `configs/` | Config file template |
| `scripts/` | Build, install, CI scripts |
| `test/` | Integration test, test fixtures |
| `docs/` | Dokumentasi tambahan |
| `third_party/` | External tools, vendored code |

### Tips Umum

- **Jangan buat `utils/` atau `helpers/`** — terlalu generik, pecah jadi package spesifik
- **Gunakan `internal/`** sebisa mungkin untuk mencegah coupling tidak sengaja
- **Satu `main.go` per binary** — jika butuh banyak binary, pakai `cmd/<nama>/main.go`
- **Ikuti [Standard Go Project Layout](https://github.com/golang-standards/project-layout)** sebagai referensi, bukan aturan wajib
- **Package name = nama direktori** — hindari `package util`, pakai `package fileutil`, `package httputil`, dll

---

*Generated for Go project structure reference — semua bidang*
