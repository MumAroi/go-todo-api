# Go Todo API - สรุปความรู้และเทคนิค

## โครงสร้าง Project (Domain-Based Architecture)

```
go-todo-api/
├── cmd/api/main.go          # Entry point
├── internal/
│   ├── auth/                # Domain: Authentication
│   │   ├── handler.go       # Register, Login handlers
│   │   └── routes.go        # Auth routes
│   ├── users/               # Domain: Users
│   │   ├── handler.go       # CRUD handlers
│   │   ├── model.go         # User struct
│   │   ├── repository.go    # Database operations
│   │   └── routes.go        # User routes
│   ├── todos/               # Domain: Todos
│   │   ├── handler.go       # CRUD handlers
│   │   ├── model.go         # Todo struct
│   │   ├── repository.go    # Database operations
│   │   └── routes.go        # Todo routes
│   ├── shared/              # Shared modules
│   │   ├── middleware/      # Auth middleware
│   │   ├── utils/           # Validation error parser
│   │   └── validators/      # Custom validators
│   ├── config/              # Environment config
│   └── database/            # Database connection
├── migrations/              # SQL migrations
├── docker-compose.yml       # PostgreSQL container
└── .env                     # Environment variables
```

**ข้อดีของ Domain-Based Structure:**
- โค้ดที่เกี่ยวข้องกันอยู่ด้วยกัน (cohesion)
- หา code ง่ายขึ้นเมื่อ project โต
- แยก domain ชัดเจน ลด coupling
- ขยาย domain ใหม่ได้ง่าย (เช่น `internal/comments/`)

---

## เทคนิคและเครื่องมือที่ใช้

### 1. Gin Framework

**สิ่งที่ใช้:**
- `gin.Default()` - สร้าง router พร้อม logger + recovery middleware
- `gin.H{}` - shortcut สำหรับ `map[string]interface{}`
- `ctx.ShouldBindJSON()` - bind JSON request body
- `ctx.JSON()` - ส่ง JSON response
- `ctx.Get()` - ดึงค่าจาก context (เช่น user_id จาก middleware)
- `ctx.Set()` - เก็บค่าใน context
- `ctx.Abort()` - หยุด middleware chain
- `router.Group()` - จัดกลุ่ม routes

**BindJSON vs ShouldBindJSON:**
```go
// BindJSON - auto-handle error (abort + 400)
ctx.BindJSON(&req)

// ShouldBindJSON - ต้อง handle error เอง (customize ได้)
if err := ctx.ShouldBindJSON(&req); err != nil {
    ctx.JSON(400, gin.H{"error": err.Error()})
    return
}
```

---

### 2. GORM (ORM)

**Struct Tags:**
```go
type User struct {
    ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Password  string    `gorm:"not null" json:"password"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
```

**GORM Tags ที่ใช้บ่อย:**
- `primaryKey` - กำหนด primary key
- `type:uuid` - กำหนด column type
- `default:gen_random_uuid()` - default value (database function)
- `unique` - unique constraint
- `not null` - not null constraint
- `autoCreateTime` - auto-set timestamp ตอน create
- `autoUpdateTime` - auto-update timestamp ตอน update
- `column:name` - เปลี่ยน column name

**GORM Methods:**
```go
db.Create(&user)                    // INSERT
db.Find(&users)                     // SELECT *
db.First(&user, id)                 // SELECT * WHERE id = ? LIMIT 1
db.Where("id = ?", id).First(&todo) // SELECT * WHERE id = ? LIMIT 1
db.Model(&Todo{}).Where("id = ?", id).Updates(updates) // UPDATE
db.Delete(&Todo{}, id)              // DELETE
```

**UUID + Database Default:**
```go
// Model
ID uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`

// GORM จะไม่ส่ง id ใน INSERT ถ้าเป็น zero value
// PostgreSQL จะใช้ DEFAULT gen_random_uuid()
// GORM จะดึง id กลับมาอัตโนมัติ
```

---

### 3. UUID (Universally Unique Identifier)

**ทำไมใช้ UUID แทน Integer:**
- Global uniqueness - ไม่ต้องกลัวชนกันแม้มีหลาย database/services
- Security - ไม่สามารถเดา ID ถัดไปได้ (ป้องกัน enumeration attacks)
- Distributed systems - สร้าง ID ได้จาก client โดยไม่ต้องถาม server
- Merge-friendly - รวม data จากหลาย sources ได้ง่าย

**UUID vs ULID vs UUIDv7:**
- UUID v4 - random, ไม่ sequential
- ULID - sequential + random, sort ได้ตามเวลา
- UUIDv7 - sequential, sort ได้ตามเวลา (นิยมใหม่)

**Package ที่ใช้:** `github.com/google/uuid`
```go
id := uuid.New()           // สร้าง UUID v4
fmt.Println(id.String())   // "550e8400-e29b-41d4-a716-446655440000"
```

---

### 4. JWT (JSON Web Token) Authentication

**Flow:**
```
1. User login (email + password)
2. Server ตรวจสอบ → สร้าง JWT token
3. Client เก็บ token → ส่งใน Authorization header ทุก request
4. Server ตรวจสอบ token → อนุญาต/ปฏิเสธ
```

**สร้าง Token:**
```go
claims := jwt.MapClaims{
    "user_id": user.ID,
    "email":   user.Email,
    "exp":     time.Now().Add(cfg.JWTExpiration).Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
```

**ตรวจสอบ Token (Middleware):**
```go
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
    // ตรวจสอบ algorithm
    if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return []byte(cfg.JWTSecret), nil
})

if err != nil || !token.Valid {
    c.JSON(401, gin.H{"error": "Invalid or expired token"})
    c.Abort()
    return
}

// เก็บ user_id ใน context
claims := token.Claims.(jwt.MapClaims)
c.Set("user_id", claims["user_id"].(string))
c.Next()
```

**Security Best Practices:**
- ตรวจสอบ algorithm ทุกครั้ง (ป้องกัน algorithm confusion attack)
- ใช้ secret ที่ยาวพอ (32+ characters)
- ตั้ง expiration time ให้เหมาะสม
- อย่า commit secret ขึ้น git

---

### 5. Password Hashing (bcrypt)

**ทำไมต้องใช้ bcrypt:**
- Slow by design - ป้องกัน brute force attacks
- Salt อัตโนมัติ - ไม่ต้องจัดการ salt เอง
- Adaptive - เพิ่ม cost ได้ตาม hardware ที่แรงขึ้น

**Hash Password:**
```go
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(req.Password),
    bcrypt.DefaultCost,  // cost = 10 (2^10 iterations)
)
```

**Verify Password:**
```go
err := bcrypt.CompareHashAndPassword(
    []byte(hashedPassword),
    []byte(req.Password),
)
// nil = ตรง, bcrypt.ErrMismatchedHashAndPassword = ไม่ตรง
```

---

### 6. Validation (go-playground/validator)

**Basic Validation Tags:**
```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}
```

**Tags ที่ใช้บ่อย:**
- `required` - ต้องมีค่า
- `email` - ต้องเป็น email format
- `min=6` - ความยาวอย่างน้อย 6
- `max=100` - ความยาวไม่เกิน 100
- `omitempty` - ข้าม validation ถ้าเป็น zero value

**Pointer Fields + Validation:**
```go
type UpdateUserRequest struct {
    Email    *string `json:"email" binding:"omitempty,email"`
    Password *string `json:"password" binding:"omitempty,min=6"`
}
```
- `nil` → ข้าม validation (ไม่ update)
- `""` → validate (fail ถ้าไม่ผ่านเงื่อนไข)
- `"value"` → validate ปกติ

**Custom Validator:**
```go
func strongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    
    if password == "" {
        return false
    }
    
    if len(password) < 6 || len(password) > 10 {
        return false
    }
    
    hasLower, hasUpper, hasDigit, hasSpecial := false, false, false, false
    
    for _, char := range password {
        switch {
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsDigit(char):
            hasDigit = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }
    
    return hasLower && hasUpper && hasDigit && hasSpecial
}

// Register
validate.RegisterValidation("strong_password", strongPassword)
```

**Register Tag Name Function:**
```go
v.RegisterTagNameFunc(func(fld reflect.StructField) string {
    name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
    if name == "-" {
        return ""
    }
    return name
})
```
- ใช้ json tag name ใน error message แทน struct field name
- `RegisterRequest.Email` → `email`

**Parse Validation Errors:**
```go
func ParseValidationError(err error) BindErrorResponse {
    errorMap := make(map[string]string)
    
    if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
        for _, fe := range ve {
            errorMap[fe.Field()] = getCustomMessage(fe)
        }
    }
    
    return BindErrorResponse{
        Success: false,
        Message: "ข้อมูลนำเข้าไม่ถูกต้อง",
        Errors:  errorMap,
    }
}
```

---

### 7. Dependency Injection (DI)

**Pattern ที่ใช้: Manual DI (ส่ง dependencies ตรงๆ)**

```go
// Handler รับ repository ตรงๆ
func GetUserHandler(repo *UserRepository) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        users, _ := repo.GetUsers()  // ใช้ repo โดยตรง
        ctx.JSON(200, users)
    }
}

// main.go - ส่ง dependencies ตรงๆ
userRepo := users.NewUserRepository(db)
router.GET("/users", users.GetUserHandler(userRepo))
```

**ข้อดี:**
- Type-safe - compile-time check
- Explicit - เห็นชัดว่าใช้อะไร
- Testable - mock dependencies ง่าย
- No type assertion - ไม่ต้อง assert

**เปรียบเทียบกับ Container Pattern:**

| | Manual DI | Container + Type Assert |
|---|---|---|
| Type Safety | ✓ Compile-time | ✗ Runtime error |
| Code ใน handler | ✓ ใช้ตรงๆ | ⚠️ ต้อง type assert |
| main.go | ⚠️ ยาวกว่า | ✓ สั้น |
| Testing | ✓ Mock ง่าย | ⚠️ Mock ยาก |

**เมื่อ project โต (> 15 repos):**
- ใช้ interfaces ใน shared package
- ใช้ DI framework (Wire/Fx)

---

### 8. Middleware

**Auth Middleware Pattern:**
```go
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. ตรวจสอบ header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // 2. Extract token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 3. Verify token
        token, err := jwt.Parse(tokenString, ...)
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }
        
        // 4. เก็บข้อมูลใน context
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", claims["user_id"].(string))
        
        // 5. ไป handler ถัดไป
        c.Next()
    }
}
```

**Middleware Best Practices:**
- ใช้ `c.Abort()` เมื่อ fail
- ตรวจสอบ algorithm ทุกครั้ง
- เก็บข้อมูลใน context ด้วย `c.Set()`
- Return "Invalid credentials" แทน "User not found" หรือ "Wrong password" (ป้องกัน enumeration)

---

### 9. Routes Organization

**แยก Routes ตาม Domain:**
```go
// internal/auth/routes.go
func RegisterRoutes(router *gin.Engine, userRepo *UserRepository, cfg *config.Config) {
    router.POST("/auth/register", RegisterHandler(userRepo))
    router.POST("/auth/login", LoginHandler(userRepo, cfg))
}

// internal/users/routes.go
func RegisterRoutes(router *gin.RouterGroup, repo *UserRepository) {
    router.GET("", GetUserHandler(repo))
    router.GET("/:id", GetUserByIDHandler(repo))
    router.PUT("/:id", UpdateUserHandler(repo))
    router.DELETE("/:id", DeleteUserHandler(repo))
}
```

**main.go:**
```go
// Public routes
auth.RegisterRoutes(router, userRepo, cfg)

// Protected routes
protected := router.Group("")
protected.Use(middleware.AuthMiddleware(cfg))

users.RegisterRoutes(protected.Group("/users"), userRepo)
todos.RegisterRoutes(protected.Group("/todos"), todoRepo)
```

**ข้อดี:**
- Routes อยู่กับ domain
- main.go สั้น
- เพิ่ม domain ใหม่ → สร้าง routes.go ใน domain นั้น

---

### 10. Database Index

**Index คืออะไร:**
- ทำงานเหมือนสารบัญหนังสือ
- ทำให้ query เร็วขึ้น (O(log n) แทน O(n))

**สร้าง Index:**
```sql
CREATE INDEX idx_users_email ON users(email);
```

**ข้อควรระวัง:**
- `UNIQUE` constraint สร้าง index อัตโนมัติ → ไม่ต้องสร้างซ้ำ
- ใช้ index กับ column ที่ query บ่อย (WHERE, JOIN, ORDER BY)
- อย่าสร้างทุก column เพราะจะเปลือง resource

---

### 11. TIMESTAMP WITH TIME ZONE

**TIMESTAMPTZ vs TIMESTAMP:**
```sql
-- WITH TIME ZONE (แนะนำ)
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP

-- WITHOUT TIME ZONE
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

**TIMESTAMPTZ:**
- เก็บ timezone ไว้ด้วย
- PostgreSQL แปลงเป็น UTC อัตโนมัติก่อนเก็บ
- Query แล้ว convert ตาม session timezone
- เหมาะกับ app ที่มี user หลายประเทศ

**หมายเหตุ:** `TIMESTAMPTZ` เป็น alias ของ `TIMESTAMP WITH TIME ZONE`

---

### 12. Docker + PostgreSQL

**docker-compose.yml:**
```yaml
services:
  postgres:
    image: postgres:16
    container_name: todo-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: todo_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**คำสั่งที่ใช้:**
```bash
docker-compose up -d      # Start container
docker-compose down       # Stop container
docker-compose logs -f    # ดู logs
```

---

### 13. Database Migration

**Migration Files:**
```
migrations/
├── 000001_create_todos_table.up.sql
├── 000001_create_todos_table.down.sql
├── 000002_create_users_table.up.sql
└── 000002_create_users_table.down.sql
```

**Up Migration (สร้าง table):**
```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Down Migration (ลบ table):**
```sql
DROP TABLE IF EXISTS users;
```

---

### 14. Environment Variables

**Config Struct:**
```go
type Config struct {
    DatabaseURL    string        `env:"DATABASE_URL"`
    AppPort        string        `env:"APP_PORT"`
    JWTSecret      string        `env:"JWT_SECRET"`
    JWTExpiration  time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`
}
```

**Load Config:**
```go
func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

**.env file:**
```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/todo_db?sslmode=disable
APP_PORT=3005
JWT_SECRET=your-secret-key-here
JWT_EXPIRATION=24h
```

**Best Practices:**
- โหลด config ครั้งเดียวใน main.go
- ส่งผ่าน dependencies ให้ handlers
- อย่า commit .env ขึ้น git
- ใช้ secret ที่ยาวพอ (32+ characters)

---

### 15. Error Handling

**HTTP Status Codes ที่ใช้:**
```go
http.StatusOK          // 200 - สำเร็จ
http.StatusCreated     // 201 - สร้างสำเร็จ
http.StatusNoContent   // 204 - ลบสำเร็จ (ไม่มี response body)
http.StatusBadRequest  // 400 - request ผิด format
http.StatusUnauthorized // 401 - ไม่ได้ login หรือ token หมดอายุ
http.StatusNotFound    // 404 - ไม่เจอ resource
http.StatusConflict    // 409 - duplicate (เช่น email ซ้ำ)
http.StatusInternalServerError // 500 - server error
```

**Error Response Format:**
```go
// Single error
gin.H{"error": "error message"}

// Multiple validation errors
gin.H{
    "success": false,
    "message": "ข้อมูลนำเข้าไม่ถูกต้อง",
    "errors": map[string]string{
        "email": "รูปแบบอีเมลไม่ถูกต้อง",
        "password": "รหัสผ่านต้องมีความแข็งแกร่ง",
    }
}
```

---

## สรุป Pattern ที่ใช้

| Pattern | ใช้ที่ไหน | ข้อดี |
|---------|----------|-------|
| **Domain-Based Structure** | โครงสร้าง project | Organized, scalable |
| **Manual DI** | Handler dependencies | Type-safe, explicit |
| **Repository Pattern** | Database operations | Separation of concerns |
| **Middleware** | Authentication | Reusable, clean |
| **Custom Validators** | Password validation | Reusable, maintainable |
| **Routes per Domain** | Route organization | Organized, easy to extend |

---

## คำสั่งที่ใช้บ่อย

```bash
# Run server
go run cmd/api/main.go

# Build
go build ./...

# Install dependencies
go mod tidy

# Hot reload (air)
air

# Database migration
migrate -path migrations -database "postgres://..." -verbose up

# Docker
docker-compose up -d
docker-compose down
```

---

## สิ่งที่ควรทำเพิ่ม

- [ ] เพิ่ม unit tests
- [ ] เพิ่ม integration tests
- [ ] เพิ่ม rate limiting
- [ ] เพิ่ม CORS configuration
- [ ] เพิ่ม logging (structured logging)
- [ ] เพิ่ม health check endpoint
- [ ] เพิ่ม pagination สำหรับ list endpoints
- [ ] เพิ่ม refresh token สำหรับ JWT
- [ ] เพิ่ม email verification
- [ ] เพิ่ม password reset functionality

---

## สรุป

Project นี้ครอบคลุมความรู้พื้นฐานที่สำคัญในการสร้าง REST API ด้วย Go:
- **Architecture:** Domain-based structure, DI pattern
- **Database:** GORM, PostgreSQL, UUID, migrations
- **Authentication:** JWT, bcrypt, middleware
- **Validation:** Custom validators, error parsing
- **Organization:** Routes per domain, shared modules

สิ่งที่ได้เรียนรู้ไม่ใช่แค่ syntax แต่เป็น **best practices** และ **design patterns** ที่สามารถนำไปใช้กับ project อื่นได้
