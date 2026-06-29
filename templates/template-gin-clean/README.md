# template-gin-clean

> GoStack starter template — Gin + Clean Architecture

This template is used by [GoStack CLI](https://github.com/gostack/cli) when you run:

```bash
gostack create my-api --framework gin --arch clean
```

## Structure

```
template/
├── cmd/api/            # entry point
├── internal/
│   ├── config/         # env config
│   ├── database/       # db connection
│   ├── entity/         # domain models + request types
│   ├── handler/        # HTTP handlers (Gin)
│   ├── middleware/      # CORS, JWT
│   ├── repository/     # data access layer
│   ├── routes/         # route registration (with gostack:routes anchor)
│   └── service/        # business logic
├── migrations/         # SQL migration files
├── docs/               # Swagger output
├── .env.example
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## Placeholders

| Placeholder        | Replaced with         |
|--------------------|-----------------------|
| `{{MODULE_NAME}}`  | Go module path        |
| `{{PROJECT_NAME}}` | Project name          |

## manifest.json

```json
{
  "name": "Gin Clean Architecture",
  "framework": "gin",
  "architecture": "clean",
  "version": "1.0.0",
  "go": "1.22"
}
```
