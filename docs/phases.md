# Full Phase Steps for Coding StreamLoft

Based on the provided SRS and LLM rules. The project must stay separated into `backend-api/`, `windows-app/`, `media-server/`, `database/`, and `docs/`; Go code only belongs in `backend-api/`, C# code only in `windows-app/`, SRS config only in `media-server/`, and admin SQL only in `database/admin-sql/`. 

---

## Phase 0 — Coding Boundaries

1. Read `SRS.md`.
2. Read `docs/LLM_RULES.md`.
3. Do not create extra database tables.
4. Do not create extra API endpoints.
5. Do not create extra UI screens.
6. Do not hardcode secrets.
7. Do not commit encryption keys.
8. Keep each component inside its own folder.
9. Use only SRS-defined behavior, validation, security, data handling, interfaces, tests, and pass/fail rules. 

---

## Phase 1 — Project Folder Setup

Create or verify this structure:

```text
StreamLoft/
├── backend-api/
├── windows-app/
├── media-server/
├── database/
│   └── admin-sql/
└── docs/
```

Steps:

1. Place Go backend API code only in `backend-api/`.
2. Place C# Windows app code only in `windows-app/`.
3. Place SRS config only in `media-server/`.
4. Place admin SQL scripts only in `database/admin-sql/`.
5. Place documentation and rules in `docs/`.

---

## Phase 2 — Database Schema Coding

Use only the SRS-defined data entities:

```text
User
UserDestination
BroadcastSession
UserMachine
UserSession
```

The SRS defines these entities for users, destinations, broadcast history, machine login tracking, and per-machine sessions. 

Steps:

1. Create `database/schema.sql`.

2. Create `User` table.

3. Add:

   * `id`
   * `numeric_id`
   * `name`
   * `stream_key`
   * `created_at`
   * `updated_at`

4. Enforce unique `numeric_id`.

5. Enforce unique `stream_key`.

6. Create `UserDestination` table.

7. Add:

   * `id`
   * `user_id`
   * `name`
   * `rtmp_url`
   * encrypted `stream_key`
   * `enabled`
   * `created_at`
   * `updated_at`

8. Add foreign key from `UserDestination.user_id` to `User.id`.

9. Create `BroadcastSession` table.

10. Add:

* `id`
* `user_id`
* `user_destination_id`
* `date`
* `duration_minutes`
* `started_at`
* `ended_at`

11. Add foreign keys to `User` and `UserDestination`.

12. Store timestamps using UTC-5 behavior required by the SRS.

13. Create `UserMachine` table.

14. Add:

* `id`
* `machine_id`
* `user_id`
* `last_used_at`

15. Create `UserSession` table.
16. Add:

* `id`
* `user_id`
* `machine_id`
* encrypted `access_token`
* encrypted `refresh_token`
* `token_expires_at`
* `created_at`
* `updated_at`

17. Add indexes for:

* `users.numeric_id`
* `users.stream_key`
* `user_destinations.user_id`
* `broadcast_sessions.user_id`
* `broadcast_sessions.user_destination_id`
* `user_machines.machine_id`
* `user_sessions.access_token`
* `user_sessions.refresh_token`

18. Add cleanup SQL for broadcast sessions older than 30 days, because broadcast sessions are retained for 30 days. 

---

## Phase 3 — Admin SQL Coding

The SRS says admin management is done through SQL files, not an admin API. 

Create files inside:

```text
database/admin-sql/
```

Required admin SQL files:

```text
create_user.sql
update_user.sql
assign_destination.sql
update_destination.sql
edit_streamloft_stream_key.sql
edit_destination_stream_key.sql
disable_destination.sql
delete_or_deactivate_user.sql
list_users.sql
list_user_destinations.sql
list_broadcast_sessions.sql
```

Steps:

1. Create SQL to add users.
2. Create SQL to assign destinations to users.
3. Create SQL to edit destinations.
4. Create SQL to edit StreamLoft stream keys.
5. Create SQL to edit destination stream keys.
6. Create SQL to list users.
7. Create SQL to list destinations per user.
8. Create SQL to list broadcast sessions by site and date.
9. Do not create an admin API.

---

## Phase 4 — Go API Project Setup

The Go API must follow Go conventions, use Go modules, handle errors properly, and use environment variables for configuration. 

Inside:

```text
backend-api/
```

Create:

```text
backend-api/
├── cmd/
│   └── streamloft-api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── auth/
│   ├── users/
│   ├── destinations/
│   ├── streams/
│   ├── broadcasts/
│   ├── workers/
│   ├── crypto/
│   ├── middleware/
│   └── validation/
├── .env.example
├── go.mod
└── go.sum
```

Steps:

1. Initialize Go module.
2. Add config loader.
3. Add database connection.
4. Add routing.
5. Add middleware.
6. Add validation.
7. Add structured error responses.
8. Add logging.
9. Add encryption service.
10. Add forwarding worker manager.

---

## Phase 5 — Go API Configuration Coding

The SRS defines runtime config locations:

```text
backend-api/.env.example
/etc/streamloft/api.env
/etc/streamloft/streamloft_master.key
```

The Go runtime config must stay outside the database and must not be hardcoded. 

Steps:

1. Create `backend-api/.env.example`.
2. Include only variable names.
3. Do not include real secrets.
4. Load runtime variables from `/etc/streamloft/api.env`.
5. Load encryption key from `/etc/streamloft/streamloft_master.key`.
6. Ensure only the Go API process reads the encryption key.
7. Fail startup if required runtime config is missing.
8. Fail startup if encryption key file is missing.
9. Do not store the encryption key in:

   * `.env`
   * `appsettings.json`
   * database
   * Git
   * SRS config
   * Windows app

---

## Phase 6 — Authentication API Coding

Use only the API endpoints defined in SRS Section 13.2. 

Code these routes:

```text
POST /auth/login
POST /auth/refresh
POST /auth/logout
```

Steps for `POST /auth/login`:

1. Accept numeric user ID.
2. Accept machine ID.
3. Validate user ID is numeric and within `1-999999`. 
4. Look up user by numeric ID.
5. If user does not exist, return invalid ID error.
6. If user has no StreamLoft stream key, generate one.
7. Store generated StreamLoft stream key.
8. Create or update `UserMachine`.
9. Create `UserSession`.
10. Generate access token.
11. Generate refresh token.
12. Set access token expiry to 3 days.
13. Return:

    * user name
    * numeric ID
    * StreamLoft RTMP URL
    * StreamLoft stream key
    * access token
    * refresh token
    * allowed destinations

Steps for `POST /auth/refresh`:

1. Accept refresh token.
2. Validate refresh token exists.
3. Validate session exists.
4. Generate new access token.
5. Replace old access token.
6. Update token expiry.
7. Return new access token.

Steps for `POST /auth/logout`:

1. Accept access token.
2. Resolve current user session.
3. Delete only the session for the current machine.
4. Do not delete other sessions.
5. Return logout success.

Logout must only affect the current machine session. 

---

## Phase 7 — User API Coding

Code:

```text
GET /user
```

Steps:

1. Require access token.
2. Validate token.
3. Resolve current user.
4. Return:

   * user name
   * numeric ID
   * StreamLoft RTMP URL
   * StreamLoft stream key
   * allowed destinations
5. Do not return full saved destination stream keys after they are saved.
6. Return only allowed destinations for the current user.

---

## Phase 8 — Destination API Coding

Code:

```text
GET /destinations
PUT /destinations/:id
PUT /destinations/:id/toggle
```

Steps for `GET /destinations`:

1. Require access token.
2. Resolve current user.
3. Query only destinations assigned to that user.
4. Return destination names, RTMP URLs, enabled state, and configured/not-configured status.
5. Do not return full destination stream keys.

Steps for `PUT /destinations/:id`:

1. Require access token.
2. Resolve current user.
3. Validate destination belongs to current user.
4. Validate stream key is non-empty and max 256 chars. 
5. Encrypt destination stream key.
6. Save encrypted stream key.
7. Return saved status.
8. Do not return the full stream key.

Steps for `PUT /destinations/:id/toggle`:

1. Require access token.
2. Resolve current user.
3. Validate destination belongs to current user.
4. Validate destination has a configured stream key before enabling.
5. Update enabled state.
6. If user is live and destination is enabled, start only that destination worker.
7. If user is live and destination is disabled, stop only that destination worker.
8. Do not affect other destinations.
9. Do not affect other users.

Destination toggling must affect only the current user and not other users. 

---

## Phase 9 — Stream Callback API Coding

Code:

```text
POST /stream/start
POST /stream/stop
GET /stream/status
```

Steps for `POST /stream/start`:

1. Accept SRS callback data.
2. Identify user from StreamLoft stream key.
3. Mark user stream as live.
4. Detect bitrate if available.
5. Compare actual bitrate to configured bitrate.
6. Trigger warning if bitrate is below threshold.
7. Query enabled destinations for user.
8. Start one isolated forwarding worker per enabled destination.
9. Create one broadcast session record per destination.
10. Return callback success.

Steps for `POST /stream/stop`:

1. Accept SRS callback data.
2. Identify user from StreamLoft stream key.
3. Mark user stream as offline.
4. Stop active forwarding workers for that user.
5. Update active broadcast sessions.
6. Set `ended_at`.
7. Calculate `duration_minutes`.
8. Return callback success.

Steps for `GET /stream/status`:

1. Require access token.
2. Resolve current user.
3. Return:

   * live
   * offline
   * unknown if connection/status cannot be confirmed
   * bitrate warning status if active

SRS must notify the Go API through callbacks, and the Go API must manage forwarding independently per user destination. 

---

## Phase 10 — Forwarding Worker Coding

The SRS requires isolated forwarding workers per enabled destination. 

Create worker logic in:

```text
backend-api/internal/workers/
```

Steps:

1. Create worker manager.
2. Track workers by:

   * user ID
   * user destination ID
3. Start worker when:

   * SRS sends stream start callback
   * destination is enabled while user is already live
4. Stop worker when:

   * SRS sends stream stop callback
   * destination is disabled while user is live
5. Decrypt destination stream key only inside Go API.
6. Build external RTMP target from destination RTMP URL plus decrypted key.
7. Start forwarding process.
8. Log destination unreachable errors.
9. Continue other destination workers if one fails.
10. Never restart SRS for destination toggle changes.
11. Never rewrite global SRS config as runtime control.

The stream forwarding flow must start workers on publish and stop them on unpublish. 

---

## Phase 11 — Broadcast History API Coding

Code:

```text
GET /broadcasts
```

Steps:

1. Require access token.
2. Resolve current user.
3. Query broadcast sessions for that user only.
4. Group or order by:

   * site/destination
   * date
   * started time
5. Return:

   * destination name
   * date
   * started time
   * ended time
   * duration in minutes
6. Do not return other users’ broadcast sessions.

---

## Phase 12 — SRS Configuration Coding

Inside:

```text
media-server/srs.conf
```

Steps:

1. Configure RTMP ingest.
2. Configure HTTP callbacks.
3. Set callback for stream start to:

```text
POST /stream/start
```

4. Set callback for stream stop to:

```text
POST /stream/stop
```

5. Do not configure global forwarding as runtime control.
6. Do not require SRS restart when destination toggles change.
7. Leave runtime forwarding control to Go API workers.

SRS configuration must include HTTP callbacks and must not use global forwarding as the runtime control mechanism. 

---

## Phase 13 — Windows App Project Setup

The Windows app must follow .NET patterns, proper namespaces, MVVM where appropriate, and must never expose stream keys in UI after save. 

Inside:

```text
windows-app/
```

Create:

```text
windows-app/
├── StreamLoft.App/
│   ├── Views/
│   ├── ViewModels/
│   ├── Models/
│   ├── Services/
│   ├── Configuration/
│   └── appsettings.json
```

Steps:

1. Create Windows desktop app project.
2. Add `appsettings.json`.
3. Store only API base URL in `appsettings.json`.
4. Ensure `appsettings.json` copies beside `.exe` during publish.
5. Add API client service.
6. Add token storage service.
7. Add machine ID service.
8. Add navigation service.
9. Add validation helpers.

---

## Phase 14 — Windows App Screens Coding

Use only the SRS-defined screens:

```text
Login
Welcome
Dashboard
Destination
Events
```

The SRS defines these screens and their allowed actions. 

Steps for Login screen:

1. Show numeric ID input.
2. Validate numeric-only input.
3. Submit to `POST /auth/login`.
4. Show invalid ID error when needed.
5. Store returned tokens.
6. Navigate to Welcome screen.

Steps for auto-login:

1. Read stored machine ID.
2. Use stored session/token if valid.
3. If same machine session is valid, show Welcome screen.
4. If unavailable or expired, show Login screen.

Steps for Welcome screen:

1. Show user name.
2. Show numeric ID.
3. Add proceed button.
4. Add logout button.
5. Proceed opens Dashboard.
6. Logout calls `POST /auth/logout`.

Steps for Dashboard screen:

1. Show StreamLoft RTMP URL.
2. Show StreamLoft stream key.
3. Add copy button for RTMP URL.
4. Add copy button for StreamLoft stream key.
5. Show live/offline indicator.
6. Show allowed destinations.
7. Show destination enabled/disabled state.
8. Show bitrate warning when applicable.
9. Add logout button.
10. Poll or call `GET /stream/status`.

Steps for Destination screen:

1. Show destination name.
2. Show stream key input.
3. Hide saved stream key after save.
4. Add save button.
5. Add enable/disable toggle.
6. Save calls `PUT /destinations/:id`.
7. Toggle calls `PUT /destinations/:id/toggle`.

Steps for Events screen:

1. Call `GET /broadcasts`.
2. Show site name.
3. Show date.
4. Show duration in minutes.
5. View only.

---

## Phase 15 — Validation and Error Handling Coding

Use SRS validation rules and error behavior. 

Backend validation steps:

1. Validate user ID numeric-only.
2. Validate user ID range.
3. Validate user name non-empty and max 100 chars.
4. Validate stream key non-empty and max 256 chars.
5. Validate RTMP URL format.
6. Validate destination name non-empty and max 100 chars.
7. Validate user configured bitrate as positive integer.
8. Reject invalid database operations.
9. Return consistent error responses.

Windows app error steps:

1. Show invalid ID error.
2. Show database/API connection error.
3. Show destination unreachable warning.
4. Show stream key save error.
5. Show bitrate warning.
6. Show unknown stream status when status cannot be confirmed.
7. Keep user input when save fails.

---

## Phase 16 — Security Coding

Security requirements include token validation, encrypted destination stream keys, protected encryption key file, and never sending full saved destination stream keys back to the Windows app. 

Steps:

1. Add token middleware.
2. Require token for user endpoints.
3. Do not require user token for SRS callback endpoints unless SRS callback authentication is explicitly defined later.
4. Encrypt destination stream keys before database storage.
5. Encrypt access tokens and refresh tokens if stored as encrypted fields.
6. Store master encryption key only at:

```text
/etc/streamloft/streamloft_master.key
```

7. Ensure Go API alone reads the encryption key.
8. Do not expose destination stream keys after save.
9. Do not commit secrets.
10. Validate all user input before database operations.
11. Ensure PostgreSQL accepts connections only from API VPS IP.

---

## Phase 17 — Bitrate Warning Coding

The SRS requires bitrate detection and warning when upload is below the configured speed threshold. 

Steps:

1. Read actual incoming bitrate from stream metadata or stream inspection.
2. Read user configured bitrate.
3. Compare actual bitrate to configured bitrate.
4. If actual bitrate is below allowed threshold, create warning.
5. Expose warning in `GET /stream/status`.
6. Show warning in Windows Dashboard.
7. Log warning event.

Blocking SRS issue:

```text
SRS.md requires user configured bitrate storage, but the Section 14.2 data dictionary does not define a column for it.
```

Do not add a new column or table without approval, because the LLM rules forbid fabricating database schema or creating new tables without user approval.  

---

## Phase 18 — Integration Coding

Steps:

1. Run PostgreSQL schema.
2. Insert test user through admin SQL.
3. Assign test destinations through admin SQL.
4. Start Go API.
5. Start SRS.
6. Start Windows app.
7. Login with numeric ID.
8. Confirm StreamLoft stream key is generated.
9. Confirm allowed destinations are returned.
10. Save destination stream key.
11. Confirm destination key is encrypted in database.
12. Toggle destination on.
13. Start OBS stream to StreamLoft.
14. Confirm SRS sends start callback.
15. Confirm API marks stream live.
16. Confirm forwarding workers start.
17. Confirm external destination receives stream.
18. Stop OBS stream.
19. Confirm SRS sends stop callback.
20. Confirm API marks stream offline.
21. Confirm forwarding workers stop.
22. Confirm broadcast sessions are recorded with duration.

---

## Phase 19 — Test Coding

Required test coverage is defined for functional, UI, integration, and security checks. 

Create tests for:

```text
TC-001 User login
TC-002 Auto-login
TC-003 Welcome screen
TC-004 StreamLoft stream key generation
TC-005 Allowed destinations
TC-006 Destination stream key save
TC-007 Destination toggle
TC-008 Stream forwarding
TC-009 SRS stream start callback
TC-010 Live/offline status
TC-011 Broadcast session logging
TC-012 Bitrate warning
TC-013 Logout current machine only
TC-014 Admin SQL management
TC-015 Access token validity
TC-016 Token refresh
TC-017 Access token returned on login
```

Security tests:

1. Verify PostgreSQL is not publicly accessible.
2. Verify database only accepts API VPS IP.
3. Verify user endpoints require token.
4. Verify destination stream keys are encrypted.
5. Verify encryption key is not in Git.
6. Verify Windows app does not receive saved full destination stream keys.

Acceptance requires mandatory functional requirements to have tests, security requirements to have verification evidence, validation rules to have positive and negative coverage, API requirements to define success and error behavior, and components to remain separated. 

---

## Phase 20 — Final Compliance Pass

Steps:

1. Compare database schema against SRS Section 14.2.
2. Confirm no extra tables exist.
3. Compare API routes against SRS Section 13.2.
4. Confirm no extra endpoints exist.
5. Compare Windows screens against SRS Section 12.1.
6. Confirm no extra screens exist.
7. Confirm Go API code is only in `backend-api/`.
8. Confirm C# code is only in `windows-app/`.
9. Confirm SRS config is only in `media-server/`.
10. Confirm admin SQL files are only in `database/admin-sql/`.
11. Confirm no secrets are committed.
12. Confirm destination stream keys are encrypted.
13. Confirm destination toggle does not affect other users.
14. Confirm logout only affects current machine.
15. Confirm SRS is not restarted for destination toggles.
16. Confirm global SRS forwarding config is not rewritten for runtime control.
17. Confirm each enabled destination uses an isolated forwarding worker.
18. Confirm one worker failure does not stop other destination workers.
19. Confirm broadcast sessions are logged per destination.
20. Confirm broadcast session cleanup respects 30-day retention.

The final implementation checklist must confirm alignment with SRS, correct folder placement, no unauthorized tables, no unauthorized endpoints, no unauthorized screens, no committed secrets, and language-specific coding conventions. 
