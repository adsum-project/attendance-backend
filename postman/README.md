# Postman API Tests

Functional tests for the Adsum attendance backend API.

## Prerequisites

1. **Backend running** — `go run ./cmd/attendanceapi`
2. **Authenticated session** — Copy your session cookie from the browser into Postman:
   - Log in via the frontend
   - Use a cookie extension (e.g. "EditThisCookie") to copy the session cookie
   - In Postman: **Cookies** (below Send) → Add cookie for your API domain → paste name and value
   - Or add `Cookie: session=your-session-id` to requests

## Running in Postman

1. Import the collection and environment
2. Add your session cookie (see above)
3. Select the "Adsum - Local" environment
4. **Run collection** — tests run in order; each saves IDs for the next

## Running with Newman

Newman runs headless; you must provide a valid session cookie:

```bash
newman run postman/Adsum-Attendance-API.postman_collection.json \
  -e postman/Adsum-Local.postman_environment.json \
  --cookie-jar cookies.json \
  --export-cookies cookies.json
```

Or set a cookie via `--env-var` if you have a valid session ID.

## What the Tests Do

Tests exercise real backend logic in sequence:

| Folder | Flow |
|--------|------|
| **0 - Prerequisites** | Health check, verify auth (Me) |
| **1 - Course CRUD** | Create course → Get (assert data) → Update → Get (assert update) |
| **2 - Module CRUD** | Create module → Get → Update → Get |
| **3 - Course-Module** | Assign module to course → Get course modules (assert) → Unassign → Get (assert empty) |
| **4 - Class CRUD** | Create class in module → Get → Get list → Update → Get |
| **5 - Timetable** | Own timetable, paginated courses/modules, module courses |
| **6 - Verification** | Embedding status, own records |
| **7 - Users** | Get users (admin/staff; 403 acceptable for students) |
| **8 - Cleanup** | Delete class, module, course → Get course (assert 404) |

IDs are stored in collection variables and reused. Random course/module codes avoid 409 on re-runs.
