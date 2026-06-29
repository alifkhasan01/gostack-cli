# Contributing to GoStack CLI

Terima kasih sudah meluangkan waktu untuk berkontribusi! 🎉

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Submitting a Pull Request](#submitting-a-pull-request)
- [Development Guide](#development-guide)
- [Commit Style](#commit-style)
- [Project Structure](#project-structure)

---

## Code of Conduct

Proyek ini mengikuti prinsip sederhana: **bersikap baik dan profesional**. Diskusi teknis yang sehat selalu disambut. Komentar yang bersifat personal, diskriminatif, atau tidak relevan tidak akan ditoleransi.

---

## Getting Started

### Prerequisites

- Go 1.22+
- Git
- `gh` CLI (opsional, untuk PR)

### Setup lokal

```bash
git clone https://github.com/gostack/cli.git
cd cli

go mod tidy
make build

# Verifikasi
./bin/gostack --help
```

### Menjalankan test

```bash
make test

# Atau langsung
go test ./... -v
```

---

## How to Contribute

### Reporting Bugs

Sebelum membuat issue baru:

1. Pastikan menggunakan versi terbaru (`gostack update`)
2. Cek apakah issue serupa sudah ada di [Issues](https://github.com/gostack/cli/issues)

Ketika membuat bug report, sertakan:

- **Versi GoStack CLI**: output dari `gostack version`
- **OS dan arsitektur**: e.g. `linux/amd64`, `darwin/arm64`
- **Langkah reproduksi** yang jelas
- **Output aktual** vs **output yang diharapkan**

> ⚠️ Untuk kerentanan keamanan, **jangan buat issue publik**. Kirim email ke maintainer secara langsung.

### Suggesting Features

Buka [Issue](https://github.com/gostack/cli/issues/new) dengan label `enhancement` dan jelaskan:

- Masalah yang ingin diselesaikan
- Solusi yang kamu bayangkan
- Apakah ada alternatif yang sudah kamu pertimbangkan

### Submitting a Pull Request

1. **Fork** repository ini
2. Buat branch baru dari `main`:
   ```bash
   git checkout -b feat/nama-fitur
   ```
3. Buat perubahan dan **tulis test** jika relevan
4. Pastikan semua test pass:
   ```bash
   make test
   make vet
   ```
5. Commit dengan format yang sesuai (lihat [Commit Style](#commit-style))
6. Push dan buat PR ke branch `main`

---

## Development Guide

### Struktur Package

```
internal/
├── cli/        # cobra commands (create, generate, update, version)
├── generator/  # code generator (handler, service, repo, crud, ...)
├── printer/    # styled terminal output (lipgloss)
├── project/    # gostack.json read/write
├── replacer/   # placeholder {{MODULE_NAME}} replacement
├── runner/     # shell runner (go mod tidy, git init)
├── scaffold/   # built-in project generator (fallback dari remote template)
├── template/   # remote template downloader (GitHub zip)
├── updater/    # self-update dari GitHub releases
└── wizard/     # interactive TUI wizard (huh)
```

### Menambah Framework Baru

1. Tambah opsi di `internal/wizard/wizard.go`
2. Tambah template routes di `internal/scaffold/scaffold.go` (`routesGoTmpl`)
3. Tambah template CORS di `corsGoTmpl`
4. Buat template repo baru di `github.com/gostack-templates/template-<fw>-<arch>`

### Menambah Generator Baru

1. Tambah `Kind` constant di `internal/generator/generator.go`
2. Implementasikan fungsi generate
3. Daftarkan di `kindMap` di `internal/cli/generate.go`
4. Tulis test di `internal/generator/generator_test.go`

### Menambah Template Repo Baru

Setiap template repo mengikuti struktur:

```
template-<framework>-<arch>/
├── template/       # semua file yang akan di-copy
│   ├── cmd/api/main.go
│   ├── internal/...
│   ├── go.mod      # gunakan {{MODULE_NAME}}
│   └── .env.example # gunakan {{PROJECT_NAME}}
├── manifest.json
├── version.json
└── README.md
```

`manifest.json`:
```json
{
  "name": "Gin Clean Architecture",
  "framework": "gin",
  "architecture": "clean",
  "version": "1.0.0",
  "go": "1.22"
}
```

---

## Commit Style

Gunakan format [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <deskripsi singkat>
```

| Type | Kapan digunakan |
|------|----------------|
| `feat` | Fitur baru |
| `fix` | Bug fix |
| `docs` | Perubahan dokumentasi |
| `test` | Tambah atau perbaiki test |
| `refactor` | Refactoring tanpa perubahan perilaku |
| `chore` | Update dependency, config, tooling |
| `ci` | Perubahan CI/CD |

Contoh:

```
feat(generator): add crud generator with route injection
fix(updater): use semver comparison instead of lexicographic
docs(readme): update installation instructions
test(scaffold): add test for fiber framework scaffold
```

---

## Project Structure

```
go-cli/
├── cmd/gostack/        # binary entry point
├── internal/           # semua logic internal
├── templates/          # built-in template reference (gin-clean)
├── .github/
│   └── workflows/
│       ├── ci.yml      # test + build pada setiap push
│       └── release.yml # goreleaser pada push tag v*
├── .goreleaser.yml     # konfigurasi release multi-platform
├── Makefile
├── go.mod
└── README.md
```

---

Kalau ada pertanyaan, buka [Discussion](https://github.com/gostack/cli/discussions) atau buat [Issue](https://github.com/gostack/cli/issues). Selamat berkontribusi! 🚀
