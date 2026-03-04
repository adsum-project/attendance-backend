# Adsum – Attendance Backend

Go backend for the Adsum attendance system. Handles authentication (Microsoft Entra ID), verification (face embeddings via internal FR service, QR codes), timetable management, and user/course/student administration.

## Requirements

- **Go** 1.24+
- **SQL Server** (e.g. Azure SQL)
- **Microsoft Entra ID** app registration
- **Embeddings API** (attendance-backend-internal-fr) running

## Setup

1. Clone the repository.

2. Copy `.env.example` to `.env` and configure:

   ```
   # Retrieve from Azure
   ENTRA_CLIENT_ID=
   ENTRA_CLIENT_SECRET=
   ENTRA_TENANT_ID=

   # Callback endpoint, also needs to be registered in the enterprise app in Azure
   ENTRA_REDIRECT_URI=

   # Cookie domain
   COOKIE_DOMAIN=

   # Typically both the frontend URL
   FRONTEND_URL=
   CORS_ORIGINS=

   # Database connection string (retrieve from Azure)
   DATABASE_DSN=

   # Embeddings API URL
   EMBEDDINGS_API_URL=
   ```

   Optional:

   - `ENVIRONMENT=production` — enables production cookie settings; defaults to development when unset.

## Running

```bash
go run ./cmd/attendanceapi
```

The API listens on `:8080`. Health check: `GET /health`.

## Testing

Run all tests:

```bash
go test ./...
```

Some tests require `EMBEDDINGS_API_URL` to be set (e.g. `http://localhost:9999` for mocked endpoints). Ensure the embeddings service is reachable if running integration-style tests.

## Environment Variables

| Variable             | Required | Description                                  |
| -------------------- | -------- | -------------------------------------------- |
| `ENTRA_CLIENT_ID`    | Yes      | Azure AD / Entra ID application client ID    |
| `ENTRA_CLIENT_SECRET`| Yes      | Azure AD / Entra ID application secret       |
| `ENTRA_TENANT_ID`    | Yes      | Azure AD / Entra ID tenant ID                |
| `ENTRA_REDIRECT_URI` | Yes      | OAuth redirect URI (must be registered)       |
| `COOKIE_DOMAIN`      | Yes      | Domain for auth cookies                      |
| `FRONTEND_URL`       | Yes      | Frontend base URL (for redirects, QR flow)   |
| `CORS_ORIGINS`       | Yes      | Comma-separated allowed CORS origins         |
| `DATABASE_DSN`       | Yes      | SQL Server connection string                 |
| `EMBEDDINGS_API_URL` | Yes      | Base URL of the embeddings (internal FR) API |
| `ENVIRONMENT`        | No       | `production` for prod cookies; else dev      |
