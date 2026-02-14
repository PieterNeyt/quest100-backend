# Quest100Backend - Architectuur & Systeem Overzicht

Dit document beschrijft hoe de database-structuur, automigration en dependency injection werken in dit Go-backend project.

---

## 📋 Project Overzicht

Dit project is een Go backend API gebouwd met:
- **Framework**: Gin (web framework)
- **Database**: PostgreSQL
- **ORM**: GORM (Go ORM)
- **Architectuur**: Clean Architecture (Domain, Application, Infrastructure)
- **Dependency Injection**: Constructor-based Manual DI

---

## 🗄️ Database Setup & Auto-Migration

### Database Connectie Flow

```
main.go
  └─> Server.NewServer()
      └─> Server.NewServer() (constructor)
          └─> database.New() (Singleton pattern)
              └─> gorm.Open(postgres.Open(dsn))
                  └─> autoMigration(db)
```

### Database Initialisatie (`internal/infrastructure/database/database.go`)

1. **Connectie String**: Via environment variabelen
   ```go
   dsn := fmt.Sprintf(
       "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
       host, username, password, database, port,
   )
   ```

2. **Environment Variabelen** (uit `.env`):
   - `QUEST_DB_HOST` - PostgreSQL host
   - `QUEST_DB_USERNAME` - DB user
   - `QUEST_DB_PASSWORD` - DB password
   - `QUEST_DB_DATABASE` - Database naam
   - `QUEST_DB_PORT` - PostgreSQL port
   - `QUEST_DB_SCHEMA` - Schema naam

3. **Singleton Pattern**: 
   ```go
   var dbInstance *service
   
   func New() Service {
       if dbInstance != nil {
           return dbInstance  // Hergebruik bestaande connectie
       }
       // ... verbind en initialiseer
   }
   ```

### Auto-Migration Proces

**File**: `internal/profile/infrastructure/database/database.go`

```go
func autoMigration(db *gorm.DB) {
    databaseProfile.AutoMigration(db)
}

func AutoMigration(db *gorm.DB) {
    // 1. Profile tabel migratie
    db.AutoMigrate(&domain.Profile{})
    
    // 2. KudosEntry tabel migratie
    db.AutoMigrate(&domain.KudosEntry{})
    
    // 3. Database seeding
    seedDatabase(db)
}
```

**Wat gebeurt er automatisch:**
- GORM inspecteert de Go structs (`Profile`, `KudosEntry`)
- Maakt/updatet tabellen met kolommen op basis van struct velden
- Voegt relaties toe (foreign keys)
- Creeert indexes (bv. `Email` is `uniqueIndex`)

**Tabel Structuur:**

```
Profile (tabel)
├── id (UUID, primary key)
├── first_name (string)
├── last_name (string)
├── email (string, unique index)
├── kudos (integer)
└── created_at, updated_at (timestamps)

KudosEntry (tabel)
├── id (UUID, primary key)
├── profile_id (UUID, foreign key → Profile.id)
├── amount (integer)
├── reason (string)
└── date (timestamp)
```

**Seeding**: Na migratie wordt een test user aangemaakt:
```go
Profile{
    ID: "00000000-0000-0000-0000-000000000001",
    FirstName: "Hugo",
    LastName: "Dor",
    Email: "dorhugo@student.kdg.be",
    Kudos: 0,
}
```

---

## 🔌 Dependency Injection (DI)

Dit project gebruikt **Manual Constructor-based Dependency Injection** (geen IoC container).

### DI Flow - Profile Module

```
routing.go (Entry Point)
    │
    ├─> NewProfileRepository(db)      // Laag 1: Infrastructure
    │   └─> profileRepository{db}
    │
    ├─> NewProfileService(repo)       // Laag 2: Application
    │   └─> profileService{profileRepo}
    │
    └─> NewProfileHandler(service)    // Laag 3: API
        └─> ProfileHandler{profileService}
```

### Stap voor Stap Uitleg

#### 1️⃣ **Infrastructure Layer** - Repository
**File**: `internal/profile/infrastructure/database/repository.go`

```go
type profileRepository struct {
    db *gorm.DB  // Database dependency
}

func NewProfileRepository(db *gorm.DB) domain.ProfileRepository {
    return &profileRepository{db: db}
}

// Repository methods
func (r *profileRepository) GetProfileById(profileId uuid.UUID) (*domain.Profile, error) {
    var profile domain.Profile
    result := r.db.First(&profile, "id = ?", profileId)
    return &profile, nil
}

func (r *profileRepository) UpdateProfile(profile *domain.Profile) error {
    result := r.db.Save(profile)
    return result.Error
}
```

**Verantwoordelijkheid**: Database queries

---

#### 2️⃣ **Application Layer** - Service
**File**: `internal/profile/application/service.go`

```go
type ProfileService interface {
    HandleAttendance(classId uuid.UUID, profileId uuid.UUID) (*domain.Profile, error)
}

type profileService struct {
    profileRepo domain.ProfileRepository  // Repository dependency (interface!)
}

func NewProfileService(profileRepo domain.ProfileRepository) ProfileService {
    return &profileService{
        profileRepo: profileRepo,
    }
}

func (s *profileService) HandleAttendance(classId, profileId uuid.UUID) (*domain.Profile, error) {
    // Business logic
    profile, err := s.profileRepo.GetProfileById(profileId)
    if err != nil {
        return nil, err
    }
    
    // Add kudos
    profile.AddKudos(100, "Attendance")
    
    // Save back
    err = s.profileRepo.UpdateProfile(profile)
    return profile, nil
}
```

**Verantwoordelijkheid**: Business logic, orchestratie

**Key**: Gebruikt `interface` (`ProfileRepository`) i.p.v. concrete type → Loosely coupled!

---

#### 3️⃣ **API Layer** - Handler/Controller
**File**: `internal/profile/api/controller.go`

```go
type ProfileHandler struct {
    profileService application.ProfileService  // Service dependency (interface!)
}

func NewProfileHandler(profileService application.ProfileService) *ProfileHandler {
    return &ProfileHandler{
        profileService: profileService,
    }
}

func (h *ProfileHandler) HandleAttendance(c *gin.Context) {
    // Parse request
    classId, _ := uuid.Parse(c.Param("classId"))
    profileId, _ := uuid.Parse(os.Getenv("HARDCODED_PROFILE_ID"))
    
    // Call service
    profile, err := h.profileService.HandleAttendance(classId, profileId)
    
    // Return response
    c.JSON(http.StatusOK, gin.H{"profile": profile})
}
```

**Verantwoordelijkheid**: HTTP request/response handling

---

#### 4️⃣ **Wiring in Routes** - Assembly Point
**File**: `internal/profile/infrastructure/server/routing.go`

```go
func SetupProfileRoutes(r *gin.Engine, db *gorm.DB) {
    // 1. Maak repository met db dependency
    profileRepo := database.NewProfileRepository(db)
    
    // 2. Maak service met repository dependency
    profileService := application.NewProfileService(profileRepo)
    
    // 3. Maak handler met service dependency
    profileHandler := api.NewProfileHandler(profileService)
    
    // 4. Registreer routes
    profileGroup := r.Group("/profiles")
    profileGroup.POST("/attendance/:classId", profileHandler.HandleAttendance)
}
```

**Dit is waar alles samen komt!**

---

### DI Visueel Diagram

```
┌──────────────────────────────────────────────────┐
│  Gin Handler (HTTP Request)                      │
│  /profiles/attendance/:classId                   │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
        ┌─────────────────────┐
        │ ProfileHandler      │ ◄─── API Layer
        │ - profileService    │ (Dependency)
        └────────┬────────────┘
                 │
                 ▼ calls
        ┌─────────────────────┐
        │ ProfileService      │ ◄─── Application Layer
        │ - profileRepo       │ (Dependency, via interface)
        └────────┬────────────┘
                 │
                 ▼ calls
        ┌─────────────────────┐
        │ ProfileRepository   │ ◄─── Infrastructure Layer
        │ - db                │ (Dependency)
        └────────┬────────────┘
                 │
                 ▼ queries
            PostgreSQL
```

---

## 🏗️ Clean Architecture Lagen

```
internal/
├── profile/
│   ├── api/
│   │   └── controller.go          ◄─ Presentatie laag (HTTP)
│   ├── application/
│   │   └── service.go             ◄─ Applicatie laag (Business logic)
│   ├── domain/
│   │   └── profile.go             ◄─ Domain laag (Entities, Interfaces)
│   └── infrastructure/
│       ├── database/
│       │   ├── database.go        ◄─ Migration & Seeding
│       │   ├── repository.go      ◄─ Database implementatie
│       │   └── routing.go         ◄─ Dependency wiring
│       └── server/
│           └── routing.go         ◄─ Route setup
│
└── infrastructure/
    ├── database/
    │   └── database.go            ◄─ Global DB setup
    └── server/
        ├── server.go              ◄─ HTTP Server init
        └── routing.go             ◄─ Global routes
```

---

## 📝 Domain Layer

**File**: `internal/profile/domain/profile.go`

```go
// Interface (abstraction)
type ProfileRepository interface {
    GetProfileById(id uuid.UUID) (*Profile, error)
    UpdateProfile(profile *Profile) error
}

// Entity
type Profile struct {
    ID           uuid.UUID
    FirstName    string
    LastName     string
    Email        string
    Kudos        int
    KudosHistory []KudosEntry  // Relation
}

// Business logic (domain method)
func (p *Profile) AddKudos(kudos int, reason string) error {
    if kudos < 0 {
        return &NegativeKudosError{}
    }
    p.Kudos += kudos
    p.KudosHistory = append(p.KudosHistory, KudosEntry{
        ID:        uuid.New(),
        ProfileID: p.ID,
        Amount:    kudos,
        Reason:    reason,
    })
    return nil
}
```

**Key Punt**: Interfaces gedefinieerd in domain → Infrastructure implementeert ze!

---

## 🔄 Request Flow - Voorbeeld

**Scenario**: POST `/profiles/attendance/class-123`

```
1. HTTP Request
   └─> Gin router
       └─> ProfileHandler.HandleAttendance()
           │
           ├─> Parse params (classId, profileId)
           │
           └─> h.profileService.HandleAttendance(classId, profileId)
               │
               └─> profileService.HandleAttendance()
                   │
                   ├─> s.profileRepo.GetProfileById(profileId)
                   │   └─> SELECT * FROM profiles WHERE id = ?
                   │       └─> PostgreSQL
                   │
                   ├─> profile.AddKudos(100, "Attendance")
                   │   └─> Business logic in domain
                   │
                   └─> s.profileRepo.UpdateProfile(profile)
                       └─> UPDATE profiles SET kudos = ?, ...
                           └─> PostgreSQL
                           
2. Response
   └─> JSON {profile: {...}}
```

---

## 🌐 Server Startup Flow

```
main.go
└─> server.NewServer()
    │
    ├─> database.New()
    │   ├─> gorm.Open(postgres.Open(dsn))
    │   ├─> autoMigration(db)  ◄─ Auto creates tables
    │   │   ├─> Profile tabel
    │   │   ├─> KudosEntry tabel
    │   │   └─> Seed data
    │   └─> Return db Service
    │
    └─> NewServer HTTP server
        ├─> RegisterRoutes()
        │   └─> SetupProfileRoutes(r, db)
        │       └─> Wire up DI (repo → service → handler)
        │
        └─> ListenAndServe(:8080)

Server is ready to handle requests!
```

---

## 🐳 Docker Setup

**docker-compose.yml**: PostgreSQL container
```yaml
services:
  psql_bp:
    image: postgres:latest
    environment:
      POSTGRES_DB: ${QUEST_DB_DATABASE}
      POSTGRES_USER: ${QUEST_DB_USERNAME}
      POSTGRES_PASSWORD: ${QUEST_DB_PASSWORD}
    ports:
      - "5443:5432"
    volumes:
      - psql_volume_bp:/var/lib/postgresql
```

**Dockerfile**: Go applicatie
```dockerfile
FROM golang:1.25 AS builder
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /quest100 ./cmd/api/main.go

FROM alpine:3.23
COPY --from=builder /quest100 /quest100
EXPOSE 8080
CMD ["/quest100"]
```

---

## 🔑 Key Concepts Samengevat

| Concept | Beschrijving |
|---------|-------------|
| **Auto-Migration** | GORM inspecteert Go structs en maakt/updatet DB tabellen automatisch |
| **Seeding** | Test data wordt automatisch ingevuld bij startup |
| **Singleton DB** | Database connectie wordt eenmalig gecreëerd en hergebruikt |
| **Dependency Injection** | Dependencies worden via constructors doorgegeven (bottom-up) |
| **Interfaces** | Services/Repositories gebruiken interfaces voor loose coupling |
| **Clean Architecture** | Code georganiseerd in lagen: API, Application, Domain, Infrastructure |
| **GORM Features** | AutoMigrate, Preload (eager loading), Debug mode logging |

---

## 🚀 Startup Checklist

1. ✅ `.env` file met database credentials
2. ✅ PostgreSQL running (via docker-compose)
3. ✅ `go mod download` - dependencies
4. ✅ `go run ./cmd/api/main.go` - server start
5. ✅ Auto-migration voert automatisch uit
6. ✅ Routes beschikbaar op `http://localhost:8080`

---

## 📚 Nuttige GORM Features

```go
// Eager loading (Preload)
db.Preload("KudosHistory").First(&profile)

// Query building
db.Where("id = ?", profileId).First(&profile)

// Upsert
db.Where(Profile{ID: hardcodedID}).FirstOrCreate(&hugo)

// Migrations
db.AutoMigrate(&domain.Profile{})

// Indexes
Email string `gorm:"uniqueIndex"`  // Auto index
ProfileID uuid.UUID `gorm:"index"` // Auto index

// Foreign Keys
KudosHistory []KudosEntry `gorm:"foreignKey:ProfileID;references:ID"`
```

---

**Version**: 1.0  
**Last Updated**: 2026-02-12  
**Framework**: Go 1.25 + Gin + GORM + PostgreSQL
