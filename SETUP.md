# Setup Guide

This guide explains how to set up and run the e-commerce backend locally.

---

# Prerequisites

Make sure the following are installed on your system:

* **Go** (>= 1.22 recommended)
* **PostgreSQL**
* **Redis** (must be running)
* **golang-migrate CLI**

Install `golang-migrate`:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Verify installations:

```bash
go version
psql --version
redis-cli ping
migrate -version
```

Expected Redis response:

```
PONG
```

---

# 1. Clone the Repository

```bash
git clone https://github.com/<your-username>/<repo-name>.git
cd <repo-name>
```

---

# 2. Configure Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values.

---

# 3. Create the Database

Create the PostgreSQL database manually:

```bash
createdb <your_db_name>
```

Ensure the database name matches the value configured in `.env`.

---

# 4. Run Database Migrations

Apply the database schema:

```bash
make migrate-up
```

Rollback the last migration if needed:

```bash
make migrate-down
```

---

# 5. Run the Application

Start the backend server:

```bash
make run
```

The application entrypoint is:

```
cmd/main.go
```

The server will start on the port defined in `.env`.

---

# Redis Usage

Redis is required for **OTP storage during authentication**.

Ensure Redis is running before starting the application:

```bash
redis-cli ping
```

Expected output:

```
PONG
```

---

# Running Tests

```bash
make test
```

---

# Common Issues

### Database connection errors

Ensure PostgreSQL is running and credentials in `.env` are correct.

### Redis connection errors

Ensure Redis server is running before starting the application.
