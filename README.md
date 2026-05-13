# 🧩 FragmentsBE

> *"Because every fragment of memory deserves a map, a sound, and a snapshot."*

[![Go](https://img.shields.io/badge/Go-1.26.2+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![Gin](https://img.shields.io/badge/Gin-00ADD8?style=for-the-badge&logo=gin&logoColor=white)](https://github.com/gin-gonic/gin)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![PostGIS](https://img.shields.io/badge/PostGIS-Enabled-4DB33D?style=for-the-badge&logo=postgis&logoColor=white)](https://postgis.net/)
[![MinIO](https://img.shields.io/badge/MinIO-C72E49?style=for-the-badge&logo=minio&logoColor=white)](https://min.io/)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)

---

## 🗺️ What Is This Thing?

**FragmentsBE** is a backend service for capturing *fragments* of your life:
- 📍 **Where you were** (geolocation via PostGIS)
- 📝 **What you thought** (text notes)
- 🎵 **What you heard** (audio recordings)
- 📸 **What you saw** (photos)

Plus achievements, profiles with avatars, Google OAuth (web + Android), and file streaming through the backend.

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        🌐 HTTP Layer (Gin)                         │
│  ┌───────┐  ┌──────────┐  ┌───────────┐  ┌───────┐  ┌──────────┐ │
│  │ Auth  │  │ Fragment │  │   User    │  │Achiev.│  │  Router  │ │
│  │Handler│  │  Handler │  │  Handler  │  │Handler│  │(REST API)│ │
│  └───┬───┘  └────┬─────┘  └─────┬─────┘  └───┬───┘  └────┬─────┘ │
└──────┼───────────┼──────────────┼────────────┼───────────┼───────┘
       │           │              │            │           │
┌──────▼───────────▼──────────────▼────────────▼───────────▼───────┐
│                    🎯 Application Layer                          │
│  ┌───────┐  ┌──────────┐  ┌────────┐  ┌────────┐  ┌──────────┐ │
│  │ Auth  │  │ Fragment │  │  User  │  │Achiev. │  │   DTOs   │ │
│  │Service│  │  Service │  │ Service│  │ Service│  │          │ │
│  └───┬───┘  └────┬─────┘  └───┬────┘  └───┬────┘  └──────────┘ │
└──────┼───────────┼────────────┼────────────┼────────────────────┘
       │           │            │            │
┌──────▼───────────▼────────────▼────────────▼────────────────────┐
│               🏛️ Domain Layer (Pure Go)                        │
│  ┌────────────┐  ┌────────────────────┐  ┌───────────────────┐ │
│  │  Entities  │  │  Repository       │  │    Business       │ │
│  │(User,Frag.)│  │  Interfaces       │  │     Rules         │ │
│  └────────────┘  └────────────────────┘  └───────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
       │           │              │            │
┌──────▼───────────▼──────────────▼────────────▼────────────────────┐
│                  🔧 Infrastructure Layer                         │
│  ┌───────────────────────┐  ┌────────────────────────────────┐   │
│  │  PostgreSQL + PostGIS │  │        MinIO (S3)             │   │
│  │  (Spatial Queries 🗺️) │  │  (Photos, Sounds, Avatars)   │   │
│  └───────────────────────┘  └────────────────────────────────┘   │
└───────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites
- 🐳 Docker & Docker Compose
- 🔑 Google OAuth credentials (optional)

### 1️⃣ Clone & Configure

```bash
git clone https://github.com/dmitokk/FragmentsBE.git
cd FragmentsBE
cp .env.example .env
# Edit GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, etc.
```

### 2️⃣ Launch

```bash
make docker-up
```

Services spin up on these ports:
- 🗄️ **PostgreSQL + PostGIS** → `localhost:5434`
- 📦 **MinIO API/Console** → `localhost:10006` / `localhost:9001`
- 🚀 **App** → `localhost:10005`

---

## 📚 API Endpoints

### 🔐 Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|:----:|
| POST | `/api/auth/register` | Register with email/password | ❌ |
| POST | `/api/auth/login` | Login & get JWT | ❌ |
| GET | `/api/auth/google/url` | Get Google OAuth URL | ❌ |
| POST | `/api/auth/google` | Google OAuth callback (web) | ❌ |
| POST | `/api/auth/google/android` | Google OAuth (Android idToken) | ❌ |

### 👤 Users & Profiles

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|:----:|
| GET | `/api/users/profile` | Get own profile | ✅ |
| PUT | `/api/users/profile` | Update name + avatar (multipart) | ✅ |
| GET | `/api/users/:id` | Get public profile by user ID | ✅ |

### 🧩 Fragments

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|:----:|
| POST | `/api/fragments` | Create fragment + upload files (multipart) | ✅ |
| GET | `/api/fragments` | List fragments (spatial query: `lat`, `lng`, `radius`) | ✅ |
| GET | `/api/fragments/:id` | Get fragment by ID | ✅ |
| POST | `/api/fragments/:id/found` | Mark fragment as found | ✅ |
| GET | `/api/fragments/found` | List IDs of found fragments | ✅ |

### 🏆 Achievements

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|:----:|
| GET | `/api/achievements` | All achievements with completion status | ✅ |
| GET | `/api/achievements/mine` | Only unlocked achievements | ✅ |

### 📁 Files

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|:----:|
| GET | `/api/files/*filepath` | Serve files (photos, sounds, avatars) | ✅ |

---

## 🧪 Try It Out

### Register & Login

```bash
curl -X POST http://localhost:10005/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"supersecret"}'
```

### Create a Fragment with Photos

```bash
TOKEN="your-jwt-token"

curl -X POST http://localhost:10005/api/fragments \
  -H "Authorization: Bearer $TOKEN" \
  -F "text=Found this amazing spot!" \
  -F "lat=55.7558" \
  -F "lng=37.6173" \
  -F "photos=@cool_photo.jpg" \
  -F "sound=@ambient_noise.mp3"
```

### Update Profile with Avatar

```bash
curl -X PUT http://localhost:10005/api/users/profile \
  -H "Authorization: Bearer $TOKEN" \
  -F "name=New Name" \
  -F "avatar=@avatar.jpg"
```

### Find Nearby Fragments

```bash
curl -X GET "http://localhost:10005/api/fragments?lat=55.7558&lng=37.6173&radius=5000" \
  -H "Authorization: Bearer $TOKEN"
```

### Mark Fragment as Found

```bash
curl -X POST "http://localhost:10005/api/fragments/<fragment-id>/found" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🛠️ Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all commands |
| `make build` | Build application |
| `make run` | Run locally |
| `make docker-up` | Start all containers |
| `make docker-down` | Stop all containers |
| `make docker-logs` | View logs |
| `make docker-build` | Rebuild Docker image & restart |
| `make test` | Run tests |
| `make clean` | Remove build artifacts |

---

## 📁 Project Structure

```
cmd/fragments/main.go
internal/
├── app/                         # Init, config, migrations, seed
│   ├── app.go
│   └── config.go
├── domain/
│   ├── entity/                  # User, Fragment, UserFragment, Achievement
│   └── repository/              # Interfaces
├── application/
│   ├── dto/                     # Request/response structs
│   └── service/                 # Auth, Fragment, User, Achievement
├── infrastructure/
│   ├── persistence/postgres/    # Repository implementations
│   └── storage/minio/           # File storage client
└── http/
    ├── handler/                 # Auth, Fragment, User, Achievement, File
    ├── middleware/              # JWT auth middleware
    └── router.go                # Route setup
```

---

## 🔑 Configuration

| Variable | Description |
|----------|-------------|
| `HTTP_PORT` | App port (default: `8080`) |
| `DB_URL` | PostgreSQL connection string |
| `MINIO_ENDPOINT` | MinIO server address |
| `MINIO_ACCESS_KEY` | MinIO access key |
| `MINIO_SECRET_KEY` | MinIO secret key |
| `MINIO_BUCKET` | MinIO bucket name |
| `MINIO_USE_SSL` | MinIO SSL flag |
| `JWT_SECRET` | JWT signing secret |
| `GOOGLE_CLIENT_ID` | Web OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Web OAuth client secret |
| `GOOGLE_REDIRECT_URL` | OAuth redirect URL |
| `GOOGLE_ANDROID_CLIENT_ID` | Android OAuth client ID (optional) |
| `APP_ID` | Android/iOS app bundle ID |

---

## 🎯 Achievements (Seeded)

| Code | Name | Condition |
|------|------|-----------|
| `first_found` | Первый шаг | Find 1 fragment |
| `five_found` | Коллекционер | Find 5 fragments |
| `ten_found` | Искатель | Find 10 fragments |
| `twenty_five_found` | Охотник за воспоминаниями | Find 25 fragments |
| `fifty_found` | Легенда | Find 50 fragments |
| `with_photo` | Фотограф | Find a fragment with a photo |
| `with_sound` | Аудиофил | Find a fragment with sound |

---

## 🎭 Fun Facts

- 🗺️ Locations stored as **PostGIS geometry points** — spatially indexed memories
- 📸 Files served through backend at `/api/files/*filepath` with JWT auth
- 🔐 JWT tokens expire in **24 hours**
- 🏗️ Clean Architecture — swap Gin for anything without touching domain logic
- 🏆 Achievements auto-check when a fragment is found

---

## 📦 Postman Collection

Import `postman_collection.json` for all endpoints with auto-saved JWT tokens.

---

## 📄 License

MIT
