# Hospital Middleware

Go backend ที่ให้เจ้าหน้าที่โรงพยาบาล (staff) ค้นหาข้อมูลผู้ป่วยจาก HIS โดย staff ค้นหาได้เฉพาะผู้ป่วยในโรงพยาบาลของตัวเองเท่านั้น

**Stack**: Go + Gin + GORM + PostgreSQL + Nginx + Docker

---

## วิธีรัน

```bash
cp .env.example .env
# แก้ JWT_SECRET และ ADMIN_PASSWORD ใน .env ก่อน
make up
```

รอจนบริการพร้อม (ประมาณ 10–20 วินาที) แล้วเช็ค:

```bash
curl http://localhost/health
# → {"data":"ok"}
```

หยุดบริการ:

```bash
make down
```

---

## วิธีรัน test

```bash
# Unit tests (ไม่ต้องมี Docker)
make test

# Integration tests (ต้องมี Docker daemon ทำงานอยู่)
make test-integration

# Coverage report (เปิด browser)
make coverage

# สร้าง mocks ใหม่ (หลังแก้ interface)
make mocks

# vet
make lint
```

---


## API Reference

| Method | Path | Auth | คำอธิบาย |
|---|---|---|---|
| GET | `/health` | — | Health check |
| POST | `/admin/login` | — | Admin รับ JWT |
| POST | `/staff/create` | JWT (admin) | สร้างบัญชี staff |
| POST | `/staff/login` | — | Staff รับ JWT ผูกกับโรงพยาบาล |
| GET | `/patient/search` | JWT (staff) | ค้นหาผู้ป่วย |


## โครงสร้างโปรเจค

```
.
├── cmd/                    # entry point (--with-migrate flag)
├── di/                     # dependency injection wiring
│   ├── config/             # environment config (envconfig)
│   ├── database/           # GORM + PostgreSQL init
│   └── server/             # Gin engine + graceful shutdown
├── entity/                 # GORM models + response DTOs
│   └── migrator/           # AutoMigrate + seed hospitals + admin
├── client/his/             # HIS client (http + mock)
├── repository/             # database access layer
├── service/                # business logic layer
├── handler/                # HTTP layer (Gin)
│   └── middleware/         # RequireAuth, RequireRole, Logger
├── util/                   # JWT, bcrypt, AppError, response
├── testsuite/              # testcontainers helper
├── mocks/                  # mockery-generated mocks
├── nginx/nginx.conf
├── docs/er-diagram.md
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```


