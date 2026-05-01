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

Think of it as a digital scrapbook, but with **spatial awareness** and **clean architecture**. 🧠✨

---

## 🏗️ Architecture (The Cool Part)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        🌐 HTTP Layer (Gin)                         │
│  ┌───────────┐  ┌──────────┐  ┌───────────┐  ┌──────────────────┐  │
│  │   Auth    │  │ Fragments│  │ Middleware│  │     Router       │  │
│  │  Handler  │  │  Handler │  │   (JWT)   │  │   (REST API)     │  │
│  └─────┬─────┘  └─────┬─────┘  └───────────┘  └────────┬─────────┘  │
└────────┼──────────────┼────────────────────────────────┼────────────┘
         │              │                                │
┌────────▼──────────────▼────────────────────────────────▼────────────┐
│                    🎯 Application Layer                            │
│  ┌───────────────┐  ┌──────────────────┐  ┌──────────────────────┐  │
│  │  AuthService  │  │ FragmentService  │  │         DTOs         │  │
│  └───────┬───────┘  └────────┬─────────┘  └──────────────────────┘  │
└──────────┼───────────────────┼──────────────────────────────────────┘
           │                   │
┌──────────▼───────────────────▼──────────────────────────────────────┐
│                  🏛️ Domain Layer (Pure Go, No Dependencies)        │
│  ┌──────────────────┐  ┌──────────────────────┐  ┌──────────────┐  │
│  │     Entities     │  │  Repository Interfaces│  │   Business   │  │
│  │ (User, Fragment) │  │                      │  │    Rules     │  │
│  └──────────────────┘  └──────────────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
           │                   │
┌──────────▼───────────────────▼──────────────────────────────────────┐
│                  🔧 Infrastructure Layer                           │
│  ┌───────────────────────┐  ┌────────────────────────────────────┐  │
│  │  PostgreSQL + PostGIS │  │           MinIO (S3)               │  │
│  │  (Spatial Queries 🗺️) │  │     (Photos & Sounds 📸🎵)        │  │
│  └───────────────────────┘  └────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- 🐳 Docker & Docker Compose
- 🔑 Google OAuth credentials (optional, for Google sign-in)
- ☕ A sense of adventure (optional)

### 1️⃣ Clone It

```bash
git clone https://github.com/dmitokk/FragmentsBE.git
cd FragmentsBE
```

### 2️⃣ Configure It

```bash
cp .env.example .env
# Edit .env with your Google OAuth credentials (or leave them empty for testing)
```

### 3️⃣ Launch It

```bash
make docker-up
```

**That's it.** Three services spin up:
- 🗄️ **PostgreSQL + PostGIS** → `localhost:5434`
- 📦 **MinIO** → `localhost:9000` (console: `localhost:9001`)
- 🚀 **App** → `localhost:8080`

---

## 📚 API Endpoints

### 🔐 Authentication

| Method | Endpoint               | Description                | Auth Required |
|--------|------------------------|----------------------------|:-------------:|
| POST   | `/api/auth/register`   | Register with email/password | ❌          |
| POST   | `/api/auth/login`      | Login & get JWT token      | ❌          |
| GET    | `/api/auth/google/url` | Get Google OAuth URL       | ❌          |
| POST   | `/api/auth/google`     | Google OAuth callback      | ❌          |

### 🧩 Fragments

| Method | Endpoint               | Description                | Auth Required |
|--------|------------------------|----------------------------|:-------------:|
| POST   | `/api/fragments`       | Create fragment + upload files | ✅       |
| GET    | `/api/fragments/:id`   | Get fragment by ID         | ✅          |
| GET    | `/api/fragments`       | List fragments (with spatial query) | ✅ |
| PUT    | `/api/fragments/:id`   | Update fragment            | ✅          |
| DELETE | `/api/fragments/:id`   | Delete fragment            | ✅          |

---

## 🧪 Try It Out

### Register & Login

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"supersecret"}'
```

### Create a Fragment with Photos

```bash
TOKEN="your-jwt-token-here"

curl -X POST http://localhost:8080/api/fragments \
  -H "Authorization: Bearer $TOKEN" \
  -F "text=Found this amazing spot!" \
  -F "lat=55.7558" \
  -F "lng=37.6173" \
  -F "photos=@cool_photo.jpg" \
  -F "sound=@ambient_noise.mp3"
```

### Find Nearby Fragments

```bash
curl -X GET "http://localhost:8080/api/fragments?lat=55.7558&lng=37.6173&radius=5000" \
  -H "Authorization: Bearer $TOKEN"
```

> 💡 **Pro Tip:** The spatial query uses PostGIS `ST_DWithin` for accurate radius-based searches. Your GPS coordinates aren't just numbers here—they're *magic*. 🪄

---

## 🛠️ Makefile Commands

| Command             | Description                          |
|---------------------|--------------------------------------|
| `make help`         | Show all available commands          |
| `make build`        | Build the application                |
| `make run`          | Run locally (requires local DB/MinIO)|
| `make docker-up`    | Start all containers                 |
| `make docker-down`  | Stop all containers                  |
| `make docker-logs`  | View logs                            |
| `make docker-build` | Rebuild & restart                    |
| `make test`         | Run tests                            |
| `make clean`        | Remove build artifacts               |

---

## 📁 Project Structure

```
cmd/fragments/main.go                    # 🚪 Entry point
internal/
├── app/                                 # ⚙️ App initialization & config
├── domain/                              # 🧠 Pure domain logic (entities, interfaces)
├── application/                         # 🎯 Use cases & services
├── infrastructure/                      # 🔧 External services (DB, MinIO)
└── http/                                # 🌐 HTTP handlers, middleware, router
```

---

## 🎭 Fun Facts

- 🗺️ The database stores your locations as **PostGIS geometry points**, so your memories are geographically indexed. Even if you forget where you were, the database remembers. *Creepy? Maybe. Useful? Absolutely.*
- 📸 Photos and sounds are stored in **MinIO** (S3-compatible). Because your memories deserve a *cloud*, not just a hard drive.
- 🔐 JWT tokens expire in **24 hours**. Like Cinderella's carriage, they turn back into pumpkins at midnight. 🎃
- 🏗️ Clean Architecture means if you ever decide to swap Gin for something else, the core logic won't even notice. It's like moving houses but keeping your favorite couch. 🛋️

---

## 🧑‍💻 Development

```bash
# Run locally (ensure DB & MinIO are running)
make run

# Watch logs in real-time
make docker-logs
```

---

## 📦 Postman Collection

Import `postman_collection.json` into Postman to test all endpoints. The collection automatically saves your JWT token after login/register.

---

## 🤝 Contributing

1. Fork it 🍴
2. Create your feature branch (`git checkout -b feature/amazing-thing`)
3. Commit your changes (`git commit -m 'Add amazing thing'`)
4. Push to the branch (`git push origin feature/amazing-thing`)
5. Open a Pull Request 🎉

*Note: Contributions must include at least one emoji. This is non-negotiable.* 😎

---

## 📄 License

MIT. Do whatever you want with it. Just don't blame me if it accidentally maps your fridge to the North Pole. 🧊🗺️

---

> *"The map is not the territory, but with PostGIS, it's pretty damn close."*

Made with ❤️ and ☕ by **dmitokk**
