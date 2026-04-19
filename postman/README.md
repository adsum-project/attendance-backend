# Postman API Tests

Functional tests for the Adsum attendance backend API.

## Prerequisites

1. **Backend running** - `go run ./cmd/attendanceapi`
2. **Authenticated session** - provide a valid backend `session` cookie
   - Log in through the frontend first
   - In Postman, add the `session` cookie for `localhost`
   - In Newman, provide the same cookie via a cookie jar file
3. **Internal FR service running for verification checks** - `Get Embedding Status` depends on the internal service configured by `EMBEDDINGS_API_URL`

## Running in Postman

1. Import the collection and environment
2. Add your backend `session` cookie
3. Select the `Adsum - Local` environment
4. Run the collection in order

Important:
- Do not define `courseId`, `moduleId`, `classId`, `authUserId`, or `authRoles` in the environment.
- Those values are generated during the run and stored as collection variables.
- Hardcoded environment values for those keys will override generated values and break the sequence.

## Running with Newman

Newman runs headless; you must provide a valid backend `session` cookie in a cookie jar file:

```bash
newman run postman/Adsum-Attendance-API.postman_collection.json \
  -e postman/Adsum-Local.postman_environment.json \
  --cookie-jar /path/to/cookies.json \
  --reporters cli
```

## Collection behaviour

- The run verifies auth first and stops early if the session is invalid.
- Course, module, and class IDs are created dynamically and reused across later requests.
- Timetable assertions are role-aware:
  - `default` users are expected to access `/v1/timetable/me`
  - `admin` and `staff` users are expected to receive `403` for that route
- Verification record checks are role-aware:
  - `admin` and `staff` requests include the authenticated `userId`
- Room values are asserted using the backend's normalised uppercase form.

## What the Tests Do

| Folder | Flow |
|--------|------|
| **0 - Prerequisites** | Health check, verify auth (`/v1/auth/me`) |
| **1 - Course CRUD** | Create course -> Get -> Update -> Get |
| **2 - Module CRUD** | Create module -> Get -> Update -> Get |
| **3 - Course-Module** | Assign module to course -> Get course modules -> Unassign -> Get empty result |
| **4 - Class CRUD** | Create class -> Get -> Get list -> Update -> Get |
| **5 - Timetable** | Role-aware timetable check, paginated courses/modules, module courses |
| **6 - Verification** | Embedding status, attendance records |
| **7 - Users** | Get users |
| **8 - Cleanup** | Delete class, module, course -> verify deleted course returns `404` |

Random course and module codes are generated per run to reduce duplicate-code conflicts.
