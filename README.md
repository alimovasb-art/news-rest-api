# News REST API

> **Note**: This is an educational project created for learning and mastering Go (Golang), PostgreSQL, JWT authentication, and RESTful API architecture.

A high-performance, robust RESTful API built with Go (Golang) and PostgreSQL.

This application provides complete authentication, user management, and a news publishing system with features like dynamic searching, pagination, partial updates (PATCH), soft deletes, and JWT-based authorization.

---

## Features

- **Authentication & Security**:
  - User Registration and Login.
  - Secure Password Hashing using `bcrypt`.
  - Stateless Authentication via JWT (JSON Web Tokens).
  - Role/Ownership Authorization (Users can only modify or delete their own data and news).

- **News Management**:
  - Create, Read, Update (PUT/PATCH), and Soft Delete news articles.
  - **Dynamic Search & Filtering**: Search by title (case-insensitive `ILIKE`) and filter by `author_id`.
  - **Pagination**: Efficient SQL `LIMIT` and `OFFSET` query pagination.
  - **Relational Data**: News items automatically join and embed full author information (`models.Author`).
  - **View Counter**: Automatic view incrementing on news retrieval.

- **Database & Architecture**:
  - PostgreSQL database powered by `jackc/pgx/v5` connection pool.
  - Efficient Partial Updates using PostgreSQL `COALESCE(NULLIF(...))` queries.
  - Soft Deletes (`deleted_at IS NULL`) for data preservation.
  - Request data validation using `go-playground/validator`.

---

## Tech Stack

- **Language**: Go 1.22+
- **Database**: PostgreSQL
- **Driver/Pool**: `github.com/jackc/pgx/v5`
- **Authentication**: `github.com/golang-jwt/jwt`
- **Password Security**: `golang.org/x/crypto/bcrypt`
- **Validation**: `github.com/go-playground/validator/v10`

---

## Project Structure

```text
news-restapi/
├── handlers/         # HTTP handler functions (Users, News, Middleware)
├── models/           # Go struct definitions for DB and Request/Response DTOs
├── routes/           # Route definitions using Go 1.22 net/http ServeMux
├── storage/          # Database connection pool and schema initialization
├── utils/            # Helper utilities (JSON responses, pagination parser)
├── main.go           # Application entrypoint
└── README.md
```

---

## API Reference

### Auth & Users

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/auth/signup` | Register a new user | No |
| `POST` | `/auth/login` | Login and receive a JWT token | No |
| `GET` | `/users` | Get paginated list of active users | No |
| `GET` | `/users/{id}` | Get user details by ID | No |
| `PATCH` | `/users/{id}` | Update current user's profile | Yes |
| `DELETE` | `/users/{id}` | Soft delete current user's account | Yes |

### News

| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `GET` | `/news` | Get paginated news (Query params: `page`, `limit`, `title`, `author_id`) | No |
| `GET` | `/news/{id}` | Get news by ID (Increments view count) | No |
| `POST` | `/news` | Create a new news article | Yes |
| `PUT` | `/news/{id}` | Full update of a news article | Yes |
| `PATCH` | `/news/{id}` | Partial update of a news article | Yes |
| `DELETE` | `/news/{id}` | Soft delete a news article | Yes |

---

## Getting Started

### Prerequisites

- Go (version 1.22 or higher)
- PostgreSQL

### Setup Instructions

1. **Clone the Repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/news-restapi.git
   cd news-restapi
   ```

2. **Configure Database Connection**
   Ensure PostgreSQL is running locally and create a database named `news_db`:
   ```sql
   CREATE DATABASE news_db;
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   ```

4. **Run the Application**
   ```bash
   go run main.go
   ```
   The server will start on `http://localhost:8080`. Database tables will be created automatically on startup.

---

## Authentication Usage

To access protected endpoints, pass the JWT token in the `Authorization` header:

```http
Authorization: Bearer <your_jwt_token_here>
```

---

## License

This project is open-source and available under the [MIT License](LICENSE).
