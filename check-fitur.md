# GoStack REST API Foundation Feature Roadmap

> Draft v1.0

Dokumen ini berisi daftar fitur yang akan menjadi standar pada seluruh template **REST API Foundation** di GoStack.

Tujuan utama template adalah menyediakan **fondasi project** sehingga developer dapat langsung mulai mengembangkan business logic tanpa melakukan setup berulang.

---

# Philosophy

Template **bukan aplikasi**.

Template hanya bertanggung jawab terhadap:

* Struktur project
* Konfigurasi
* Dependency
* Setup framework
* Development tooling

Business logic tetap dibuat oleh developer.

---

# Current Foundation

Saat ini template menghasilkan struktur berikut.

```text
cmd/
└── api/
    └── main.go

internal/
├── config/
├── database/
├── entity/
├── middleware/
└── routes/

.env.example
Dockerfile
docker-compose.yml
go.mod
go.sum
Makefile
```

---

# Planned Foundation Structure

Seluruh template REST API akan menggunakan struktur berikut.

```text
cmd/
└── api/
    └── main.go

internal/
├── config/
├── database/
├── entity/
├── handler/
├── middleware/
├── repository/
├── routes/
├── service/

migrations/

.env.example
Dockerfile
docker-compose.yml
Makefile
go.mod
README.md
```

Semua folder tersebut bersifat generic dan tidak mengandung business logic.

---

# Planned Features

## 1. Handler Layer

Status

* Available

Folder

```text
internal/handler/
```

Purpose

* HTTP Handler
* Request Parsing
* Response Handling

---

## 2. Service Layer

Status

* Available

Folder

```text
internal/service/
```

Purpose

* Business Logic
* Validation
* Transaction Flow

---

## 3. Repository Layer

Status

* Available

Folder

```text
internal/repository/
```

Purpose

* Database Access
* Query Layer
* Repository Pattern

---

## 4. Migration Directory

Status

* Available

Folder

```text
migrations/
```

Purpose

* SQL Migration
* Future migration generator support

---

## 5. Configuration

Status

* Available

Folder

```text
internal/config/
```

Features

* Environment Loader
* Default Configuration
* Validation

---

## 6. Database

Status

* Available

Folder

```text
internal/database/
```

Supported

* PostgreSQL
* MySQL
* SQLite

---

## 7. Middleware

Status

* Available

Folder

```text
internal/middleware/
```

Current

* CORS
* JWT
* Request ID
* Logger
* Recovery
* Timeout
* Rate Limiter

---

## 8. Routing

Status

* Available

Folder

```text
internal/routes/
```

Purpose

* Route Registration
* API Versioning
* Grouping

---

## 9. Entity

Status

* Available

Folder

```text
internal/entity/
```

Purpose

* Domain Models
* Database Models

Template tidak menyediakan entity bawaan selain contoh sederhana.

---

## 10. Docker Support

Status

* Available

Files

```text
Dockerfile
docker-compose.yml
```

Purpose

* Local Development
* Container Deployment

---

## 11. Makefile

Status

* Available

Purpose

Shortcut command untuk developer.

Planned command

```bash
make run

make build

make test

make tidy

make migrate

make clean
```

---

## 12. Environment

Status

* Available

Files

```text
.env.example
```

Purpose

* Development Configuration

---

# Optional Features

Fitur berikut tidak selalu dibuat.

Template akan menghasilkan file hanya jika user memilih fitur tersebut.

## Swagger

```text
docs/
```

---

## JWT

```text
internal/middleware/jwt.go
```

---

## Docker

```text
Dockerfile
docker-compose.yml
```

---

## Database

Jika user memilih database.

```text
internal/database/
migrations/
```

Jika memilih "None", folder tersebut tidak dibuat.

---

# Completed Features

## Logger

```text
internal/logger/
```

Status

* Available

---

## Response Helper

```text
internal/response/
```

Status

* Available

---

## Validator

```text
internal/validator/
```

Status

* Available

---

## Error Package

```text
internal/errors/
```

Status

* Available

---

## Health Check

```text
GET /health

GET /ready

GET /live
```

Status

* Available

---

## Testing

Example

```text
handler_test.go

service_test.go
```

Status

* Available

---

## GitHub Actions

Default CI

```yaml
go test ./...

go vet ./...
```

Status

* Available

---

# Dynamic Template Generation

GoStack akan menghasilkan project sesuai pilihan user.

Contoh:

User memilih

* PostgreSQL
* Docker
* Swagger

Hasil

```text
internal/database/
migrations/
docs/
Dockerfile
docker-compose.yml
```

Jika user memilih

* Database: None
* Docker: No
* Swagger: No

Hasil

```text
cmd/
internal/
go.mod
```

Tidak ada folder yang tidak digunakan.

---

# Design Principle

Setiap folder yang dibuat oleh GoStack harus memiliki alasan yang jelas.

GoStack tidak akan membuat folder kosong yang tidak memiliki fungsi.

Template hanya menyediakan **Foundation Layer**.

Business logic seperti:

* User
* Product
* Order
* Authentication Module
* CRUD

akan dibuat menggunakan perintah:

```bash
gostack generate
```

bukan saat proses `gostack create`.

---

# Goal

Setelah menjalankan:

```bash
gostack create my-api
```

Developer diharapkan dapat langsung mulai menulis business logic tanpa melakukan setup project secara manual.

GoStack bertanggung jawab pada **Foundation**, sedangkan developer bertanggung jawab pada **Application**.
