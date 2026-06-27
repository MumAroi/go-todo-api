# Go Todo API

## Tools ที่ต้องติดตั้ง

```bash
# Live reload - auto rebuild/restart เมื่อแก้ไข code
go install github.com/air-verse/air@latest

# Database migration CLI สำหรับ PostgreSQL
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

- **air** — live reload tool, แก้ code แล้ว save จะ rebuild และ restart อัตโนมัติ
- **migrate** — จัดการ database migrations (create, up, down, rollback)

## Database Migration

### สร้าง migration ใหม่

```bash
migrate create -ext sql -dir migrations -seq <migration_name>
```

ตัวอย่าง:
```bash
migrate create -ext sql -dir migrations -seq create_users_table
```

### รัน migration ด้วย Nushell

```nushell
# โหลด migration commands
source scripts/migrate.nu

# Apply ทุก migration (สร้างตาราง)
migrate up

# Rollback ทุก migration (ลบตาราง)
migrate down

# ไปยัง version ที่ต้องการ
migrate goto 1

# ดู version ปัจจุบัน
migrate version
```

## การรัน Project ด้วย Air

### ติดตั้ง Air

Air เป็น live reload tool ที่จะ rebuild และ restart server อัตโนมัติเมื่อแก้ไข code

```bash
# ติดตั้ง air
go install github.com/air-verse/air@latest
```

**ตรวจสอบว่าติดตั้งสำเร็จ:**
```bash
air -v
```

### รัน Project

```bash
# รันด้วย air (live reload)
air 
```
