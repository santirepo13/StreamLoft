# StreamLoft SRS

---

# 1. Purpose

## 1.1 System Purpose

```text
The system exists to:
StreamLoft is a streaming relay system installed on a USA VPS with high bandwidth. Users stream once to StreamLoft via RTMP, and StreamLoft forwards that stream to multiple external platforms (YouTube, Twitch, Facebook, etc.). It solves the problem of poor international upload routing by receiving the stream on a USA VPS with better bandwidth and routing.
```

## 1.2 Review Purpose

```text
This SRS context is intended to help an LLM review:
- Bugs
- Missing requirements
- Compliance failures
- Test coverage gaps
- Implementation mismatches
- Security violations
- Data handling errors
- Guide LLM coding across all components (Go API, C# Windows app, SRS config)
```

---

# 2. Scope

## 2.1 In Scope

| ID | In-Scope Item |
| --- | --- |
| SC-001 | SRS media server configuration and stream forwarding |
| SC-002 | Go backend API for user authentication and destination management |
| SC-003 | C# Windows desktop application for stream control |
| SC-004 | PostgreSQL database for user, destination, and session data |
| SC-005 | User login by numeric ID with auto-login on same machine |
| SC-006 | Pre-allowed destinations per user (admin-assigned, unlimited as configured) |
| SC-007 | User configuration of destination stream keys |
| SC-008 | Enable/disable destinations from Windows app at runtime (does not affect other users) |
| SC-009 | Incoming stream detection with API notification |
| SC-010 | Live/Offline status display in Windows app dashboard |
| SC-011 | Stream start/stop event logging with division per site and session times |
| SC-012 | Bitrate detection and replication with warning when user uploads below 30% of configured speed |

## 2.2 Out of Scope

| ID | Out-of-Scope Item | Reason |
| --- | --- | --- |
| OOS-001 | Mobile apps | Not part of v1 scope |
| OOS-002 | Web dashboard | Windows app covers user needs |
| OOS-003 | Video transcoding/encoding | SRS handles basic forwarding |
| OOS-004 | Recording streams | Not requested |
| OOS-005 | Admin API endpoints | Admin uses SQL query files instead |

---

# 3. Glossary, Definitions, Acronyms, and Domain Terms

This section is mandatory for LLM review.

| Term | Type | Definition | System Meaning |
| --- | --- | --- | --- |
| RTMP | Protocol | Real-Time Messaging Protocol | Protocol used for streaming audio/video over TCP |
| SRS | Software | Simple Realtime Server | Media server that receives and forwards RTMP streams |
| Stream Key | Domain term | Unique identifier for stream authentication | User's key for external platforms (YouTube, Twitch, Facebook, etc.) - user inputs this |
| StreamLoft Stream Key | Domain term | Unique identifier assigned to user to stream TO StreamLoft | Assigned on first login, only changeable via admin SQL |
| Destination | Domain term | External platform where stream is forwarded | YouTube, Twitch, Facebook, etc. - also called "site" |
| Site | Domain term | Same as destination | External streaming platform |
| User ID | Domain term | Numeric identifier for user login | Users login with their assigned numeric ID |
| UserDestination | Domain term | Destination assigned to a specific user | Admin assigns which destinations each user can use; each user has their own instance |
| BroadcastSession | Domain term | Record of a streaming event per site | Logs date, duration of each stream per destination |
| Bitrate | Technical term | Data transmission rate (kbps) | User sets OBS upload bitrate, system detects actual and warns if below threshold |
| Machine ID | Technical term | Identifier for user's computer | Used for auto-login on the same machine |
| Access Token | Technical term | Short-lived token used for API authentication | Valid for 3 days |
| Refresh Token | Technical term | Long-lived credential used to request a new access token without requiring user re-authentication | Stored per session per machine |
| Forwarding Worker | Technical term | Independent process to forward to one destination | Managed by Go API, one per enabled destination |
| UTC-5 | Technical term | Timezone for all timestamps | Used for broadcast session times |
| SSE | Protocol | Server-Sent Events | HTTP-based mechanism for server-to-client push; client subscribes, server pushes events without polling |

---

# 4. Requirement Language Rules

Include this because the LLM needs to understand requirement force.

| Word | Meaning |
| --- | --- |
| Shall | Mandatory |
| Must | Mandatory |
| Must not | Forbidden |
| Should | Recommended but not mandatory |
| May | Optional / allowed |
| Can | Capability, not necessarily required |

---

# 5. System Context

## 5.1 Product Overview

| Area | Description |
| --- | --- |
| Product name | StreamLoft |
| Product type | Desktop app + API + media server service |
| Primary users | Streamers who need to stream to multiple platforms from one source, especially those with poor international upload routing |
| Main goal | Forward one incoming RTMP stream to multiple external RTMP destinations |
| Deployment environment | Media Server + Go API VPS: 172.86.73.79; Database VPS: 87.239.135.39:5432 (PostgreSQL, accessible only from API VPS) |
| Main external dependencies | SRS, PostgreSQL, OBS/streaming software on user side, external streaming platforms (YouTube, Twitch, Facebook) |

## 5.2 Actors and External Systems

| Actor / System | Type | Interaction With System |
| --- | --- | --- |
| Streamer/User | Human user | Uses C# Windows app to login, configure destinations, view live status |
| Admin | Human user | Uses SQL query files to add/edit users, assign destinations, edit stream keys, manage all data |
| PostgreSQL | Database | Stores users, destinations, broadcast sessions |
| SRS | External service | Receives incoming RTMP stream, forwards to enabled destinations |
| OBS/Streaming Software | External system | Sends stream to StreamLoft RTMP URL |
| External Platforms | External API | YouTube, Twitch, Facebook receive forwarded streams |

---

# 6. User Classes and Permissions

| User Class | Description | Allowed Actions | Forbidden Actions |
| --- | --- | --- | --- |
| Guest | Not applicable | None - login required | Access any data |
| User | Authenticated streamer | Login with numeric ID, view allowed destinations, input/configure stream keys for allowed destinations, enable/disable destinations, view live status and stream events, logout from current machine | Add new destinations, change StreamLoft stream key, access other users' data |
| Admin | System operator | Execute SQL query files to add/edit users, assign destinations, edit stream keys, manage all data | Access via Windows app (uses SQL directly) |

---

# 7. Functional Requirements

## 7.1 Functional Requirement Format

| Field | Value |
| --- | --- |
| Requirement ID | FR-001 |
| Requirement | The system shall... |
| Actor | |
| Trigger | |
| Input | |
| Processing Rule | |
| Output | |
| Error Behavior | |
| Acceptance Criteria | |
| Related Business Rules | |
| Related Test Cases | |

## 7.2 Functional Requirements

| ID | Requirement | Actor | Trigger | Input | Processing Rule | Output | Error Behavior | Acceptance Criteria | Related Business Rules | Related Test Cases |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| FR-001 | The system shall authenticate users by numeric ID | User | User enters ID in Windows app | User ID | Lookup user in database by ID | User authenticated, return StreamLoft RTMP URL + stream key + name | Invalid ID → show error | User can login with valid ID | BR-001 | TC-001 |
| FR-002 | The system shall auto-login last used user when same machine is detected | User | App launch | Machine identifier | Compare stored machine_id with current machine, get last used user for this machine | Auto-login successful, show welcome screen with ID and name | No stored machine → show login | Same machine shows welcome screen with last user's ID and name | BR-001 | TC-002 |
| FR-003 | The system shall display welcome screen after auto-login | User | Auto-login successful | User data | Display "Welcome, [Name] (ID: [numeric_id])" | Welcome screen with user info, proceed to dashboard button, logout button | Error → show login | User sees their name and ID on welcome screen | BR-001 | TC-003 |
| FR-004 | The system shall assign a StreamLoft stream key on first login | User | First login attempt | User ID | Generate unique stream key if not exists, store in database | Stream key returned and stored | Database error → show error | New users receive unique stream key | BR-002 | TC-004 |
| FR-005 | The system shall return pre-allowed destinations for the logged-in user | User | Login successful | User ID | Query destinations assigned to user | List of allowed destinations with names, no stream keys populated | No destinations → show empty list | User sees only destinations admin allowed | BR-003 | TC-005 |
| FR-006 | The system shall allow user to input stream keys for allowed destinations | User | User enters stream key for a destination | Destination ID, stream key | Validate format, store in database | Stream key saved | Invalid format → show error | User can save stream key for each allowed destination | BR-004 | TC-006 |
| FR-007 | The system shall forward to a destination only if a user_destination row exists for that user and destination | User | User opens destination config | Destination ID | Go API checks user_destination row exists with stream_key set; starts forwarding worker | Destination forwarded | No row or no stream_key → log error, don't forward | Toggle affects only current user's destination, no impact on other users | BR-005 | TC-007 |
| FR-008 | The system shall forward incoming RTMP stream to all enabled destinations | Go API | Stream pushed to StreamLoft RTMP, Go API receives callback | RTMP stream | Go API starts isolated forwarding workers that read user's incoming SRS stream and push to each enabled destination | Stream appears on external platforms | Destination unreachable → log error, continue to others | All enabled destinations receive stream | BR-005 | TC-008 |
| FR-009 | The system shall detect incoming stream and notify API | SRS | Stream starts | Stream metadata via HTTP callback | SRS calls API on_publish callback with stream info | API receives notification, logs session | API unreachable → log locally | API aware of active stream | BR-006 | TC-009 |
| FR-010 | The system shall push live/offline status updates to Windows app via Server-Sent Events (SSE) | User | Stream starts/stops from SRS callback | Stream status event | Go API pushes SSE event to subscribed client, dashboard updates instantly | SSE unavailable → fall back to polling, show unknown | User sees stream status with zero latency | BR-006 | TC-010 |
| FR-011 | The system shall log stream start/stop events with division per site and session times | Go API | Stream start/stop from SRS callbacks | Timestamps, user ID, user_destination_id | Create/update broadcast session record per destination | Session logged with date, duration in minutes | Database error → log locally | Events queryable per site with duration | BR-007 | TC-011 |
| FR-012 | The system shall detect upload bitrate and warn when below 30% of configured speed | System | Incoming stream | Actual bitrate, user configured speed | Compare actual to (configured × 0.7), warn if below | Show warning in Windows app, log warning | Detection fails → log and continue | Warning shown when upload < 30% below configured | BR-008 | TC-012 |
| FR-013 | The system shall allow logout from current machine only | User | User clicks logout button | User ID, machine ID | End session for current machine only | Logout successful, show login screen | Error → show error | User logged out, other sessions unaffected | BR-009 | TC-013 |
| FR-014 | The admin shall manage users and destinations via SQL files | Admin | Admin executes SQL | SQL commands | Execute against database | User/destination data modified | Syntax error → show error | Admin can perform all management via SQL | BR-010 | TC-014 |
| FR-015 | The system shall issue access tokens valid for 3 days | User | Login successful | User ID | Generate token with 3-day expiry | Access token returned to user | Token generation error → show error | Token valid for 72 hours | BR-011 | TC-015 |
| FR-016 | The system shall refresh access token by updating the stored token | User | User requests token refresh | User ID | Update token in database with new token and expiry | New token stored, old token replaced | Database error → show error | Token can be refreshed by updating database | BR-011 | TC-016 |
| FR-017 | The system shall return access token on successful login | User | Login successful | User ID | Generate and return token | Access token in response | Error → show error | Token included in login response | BR-011 | TC-017 |

---

# 8. Use Cases / User Flows

## UC-001: User Login

| Field | Description |
| --- | --- |
| Use Case ID | UC-001 |
| Actor | User |
| Goal | Authenticate and access stream configuration |
| Preconditions | User has numeric ID assigned by admin |
| Trigger | User opens Windows app and enters ID |
| Postconditions | User sees welcome screen then dashboard with allowed destinations |
| Related Requirements | FR-001, FR-002, FR-003, FR-004, FR-005 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | User enters numeric ID | |
| 2 | | Validate ID exists in database |
| 3 | | Check machine_id for auto-login |
| 4 | | Return StreamLoft RTMP URL, stream key, name |
| 5 | | Load allowed destinations for user |
| 6 | | Display welcome screen with ID and name |
| 7 | User clicks proceed | |
| 8 | | Display dashboard |

| Alternative Flows |

| Flow ID | Condition | Expected Behavior |
| --- | --- | --- |
| AF-001 | First login | Generate stream key, store machine_id |
| AF-002 | Same machine detected | Auto-login, show welcome with last user's info |

| Exception Flows |

| Exception ID | Error Condition | Expected System Response |
| --- | --- | --- |
| EX-001 | Invalid ID | Show "Invalid ID" error |
| EX-002 | Database unavailable | Show connection error |

---

## UC-002: Configure Destination Stream Key

| Field | Description |
| --- | --- |
| Use Case ID | UC-002 |
| Actor | User |
| Goal | Input stream key for an allowed destination |
| Preconditions | User is logged in, destination is allowed |
| Trigger | User clicks on destination to add stream key |
| Postconditions | Stream key saved to database |
| Related Requirements | FR-006 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | User clicks destination | |
| 2 | | Show input field for stream key |
| 3 | User enters stream key from platform | |
| 4 | | Validate format |
| 5 | | Encrypt and save to database |
| 6 | | Confirm save |

| Exception Flows |

| Exception ID | Error Condition | Expected System Response |
| --- | --- | --- |
| EX-001 | Invalid stream key format | Show format error |

---

## UC-003: Configure Destination Stream Key

| Field | Description |
| --- | --- |
| Use Case ID | UC-003 |
| --- | --- |
| Actor | User |
| Goal | Input stream key for an allowed destination |
| Preconditions | User is logged in, destination is assigned to user |
| Trigger | User clicks on destination to add stream key |
| Postconditions | Stream key saved to database, Go API starts forwarding worker |
| Related Requirements | FR-006, FR-007 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | User clicks destination | |
| 2 | | Show input field for stream key |
| 3 | User enters stream key from platform | |
| 4 | | Validate format |
| 5 | | Encrypt and save to database |
| 6 | | Start forwarding worker for that destination |
| 7 | | Confirm save |

| Exception Flows |

| Exception ID | Error Condition | Expected System Response |
| --- | --- | --- |
| EX-001 | Invalid stream key format | Show format error |
| EX-002 | Worker start fails | Show warning, log error |

---

## UC-004: Stream Forwarding

| Field | Description |
| --- | --- |
| Use Case ID | UC-004 |
| Actor | SRS (system) |
| Goal | Forward incoming stream to enabled destinations |
| Preconditions | User is streaming to StreamLoft RTMP |
| Trigger | RTMP stream received at StreamLoft |
| Postconditions | Stream forwarded to all enabled destinations |
| Related Requirements | FR-008, FR-009, FR-010, FR-011, FR-012 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | User streams from OBS to StreamLoft | |
| 2 | | SRS receives RTMP stream, sends on_publish callback to Go API |
| 3 | | Go API detects bitrate, compares to user's configured speed |
| 4 | | If below threshold, trigger warning |
| 5 | | Go API starts forwarding workers for all of the user's destinations that have a stream_key set |
| 6 | | Go API logs session start for each destination with stream_key |
| 7 | | Forward to all destinations with stream_key set via Go API workers |
| 8 | | SSE pushes "live" status to Windows app, dashboard updates instantly |
| 9 | User stops streaming | |
| 10 | | SRS sends on_unpublish callback to Go API |
| 11 | | Go API stops forwarding workers, logs session end, calculates duration |
| 12 | | SSE pushes "offline" status to Windows app, dashboard updates instantly |

Note: Windows app maintains a persistent SSE connection to receive push events on stream start/stop. Go API pushes status to all subscribed clients when stream lifecycle changes occur. If SSE connection is lost, app falls back to polling. SSE eliminates continuous polling requests.

---

## UC-005: Logout

| Field | Description |
| --- | --- |
| Use Case ID | UC-005 |
| Actor | User |
| Goal | End session on current machine |
| Preconditions | User is logged in |
| Trigger | User clicks logout button |
| Postconditions | Session ended for current machine, other sessions unaffected |
| Related Requirements | FR-013 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | User clicks logout | |
| 2 | | End session for current machine_id |
| 3 | | Show login screen |

---

## UC-006: Admin User Management

| Field | Description |
| --- | --- |
| Use Case ID | UC-006 |
| Actor | Admin |
| Goal | Manage users and destinations via SQL |
| Preconditions | Admin has database access |
| Trigger | Admin executes SQL query file |
| Postconditions | User/destination data modified |
| Related Requirements | FR-014 |

| Step | Actor Action | System Response |
| --- | --- | --- |
| 1 | Admin writes/loads SQL query | |
| 2 | | Execute against PostgreSQL |
| 3 | | Return result |

---

# 9. Business Rules

| Rule ID | Business Rule | Related Requirements |
| --- | --- | --- |
| BR-001 | Users authenticate using numeric ID; auto-login shows last used user with ID and name | FR-001, FR-002, FR-003 |
| BR-002 | StreamLoft stream key is assigned on first login and only changeable via admin SQL | FR-004 |
| BR-003 | Each user has pre-allowed destinations assigned by admin; unlimited as long as admin configures them | FR-005 |
| BR-004 | User inputs stream keys for allowed destinations; stream keys stored encrypted in database | FR-006 |
| BR-005 | A user streams to a destination only if a user_destination row exists for that user and destination. Assignment IS the permission. | FR-007, FR-008 |
| BR-006 | Windows app displays live/offline status based on stream detection | FR-009, FR-010 |
| BR-007 | Stream start/stop events logged per site with duration in minutes | FR-011 |
| BR-008 | System warns when upload bitrate is below 30% of user's configured speed | FR-012 |
| BR-009 | Logout ends session on current machine only, other sessions unaffected | FR-013 |
| BR-010 | Admin manages all user/destination data via SQL query files (no admin API) | FR-014 |
| BR-011 | Access tokens are valid for 3 days; refresh updates the token in the database | FR-015, FR-016, FR-017 |

---

# 10. Validation Rules

| Rule ID | Field / Input / Action | Validation Rule | Expected Error |
| --- | --- | --- | --- |
| VAL-001 | User ID input | Numeric only, 1-999999 | "ID must be a number" |
| VAL-002 | User name input | Non-empty, max 100 chars | "Name is required" |
| VAL-003 | Stream key input (user) | Non-empty string, max 256 chars | "Stream key is required" |
| VAL-004 | StreamLoft stream key | Auto-generated, unique, 32 chars | N/A (system generated) |
| VAL-005 | RTMP URL (destination) | Valid RTMP URL format | "Invalid RTMP URL" |
| VAL-006 | Destination name | Non-empty, max 100 chars | "Name is required" |
| VAL-007 | User configured bitrate | Positive integer, kbps | "Bitrate must be positive" |

---

# 11. Error Handling Requirements

| Error ID | Condition | Expected System Behavior | Related Requirement |
| --- | --- | --- | --- |
| ERR-001 | Invalid user ID | Show "Invalid ID" error in Windows app | FR-001 |
| ERR-002 | Database connection failure | Show connection error, log locally | All DB operations |
| ERR-003 | Destination RTMP unreachable | Log error, continue to other destinations, show warning | FR-008 |
| ERR-004 | SRS to API notification fails | Log locally in SRS, retry | FR-009 |
| ERR-005 | Stream key save fails | Show error message, keep input | FR-006 |
| ERR-006 | Machine ID auto-login fails | Show login screen | FR-002 |
| ERR-007 | Bitrate below threshold | Show warning in app, log warning | FR-012 |
| ERR-008 | Logout fails | Show error, session remains active | FR-013 |

---

# 12. UI Requirements

Use only if the system has a user interface.

| ID | Requirement | Screen / Component | Acceptance Criteria |
| --- | --- | --- | --- |
| UI-001 | The system shall display login screen with numeric ID input | Login Screen | User can enter ID and submit |
| UI-002 | The system shall display welcome screen showing user name and ID after auto-login | Welcome Screen | Shows "Welcome, [Name] (ID: [numeric_id])", proceed and logout buttons |
| UI-003 | The system shall display dashboard showing StreamLoft RTMP URL and stream key | Dashboard | URL and key visible, copy buttons work |
| UI-004 | The system shall display list of allowed destinations | Dashboard | Each destination shows name, toggle, stream key field |
| UI-005 | The system shall display live/offline status indicator | Dashboard | Green = Live, Gray = Offline |
| UI-006 | The system shall display bitrate warning when below threshold | Dashboard | Warning shown when upload < 30% below configured |
| UI-007 | The system shall display destinations with stream key input and status | Destination List | Each destination shows name, stream key field, and active/inactive status |
| UI-008 | The system shall allow input of stream keys for allowed destinations | Destination Detail | Input field accepts key, save button works |
| UI-009 | The system shall display stream events per site with duration | Dashboard / Events | Events listed with site name, date, duration in minutes |
| UI-010 | The system shall display logout button on dashboard | Dashboard | User can logout from current machine only |

## 12.1 Screen Definitions

| Screen ID | Screen Name | Purpose | Visible Data | Allowed Actions |
| --- | --- | --- | --- | --- |
| SCR-001 | Login | Authenticate user | ID input field | Enter ID, submit |
| SCR-002 | Welcome | Show logged in user info | User name, ID | Proceed to dashboard, logout |
| SCR-003 | Dashboard | Main control center | StreamLoft URL, stream key, live status, destination list, logout button | View, toggle destinations, input keys, logout |
| SCR-004 | Destination | Configure single destination | Destination name, stream key input, enable toggle | Input key, enable/disable, save |
| SCR-005 | Events | View stream history | List of broadcast sessions with site, date, duration | View only |

## 12.2 Color Palette

The Windows app uses the following color palette for consistent styling:

| Color Name | Hex Code | Usage |
|------------|----------|-------|
| Primary Dark | #272757 | Main background, headers |
| Primary Light | #8686AC | Accents, secondary buttons |
| Primary Medium | #505081 | Cards, containers |
| Primary Darkest | #0F0E47 | Text, dark elements |

Additional UI colors:

| Element | Color | Hex |
|---------|-------|-----|
| Live status | Green | #00FF00 |
| Offline status | Gray | #808080 |
| Bitrate warning | Orange | #FFA500 |
| Error | Red | #FF0000 |
| Success | Green | #00FF00 |

---

# 13. API Requirements

Use only if the system exposes or consumes APIs.

## 13.1 Endpoint Format

| Field | Value |
| --- | --- |
| Endpoint ID | API-001 through API-013 |
| Authentication Required | Yes (token-based for user endpoints), No for stream events from SRS |
| Related Requirements | FR-001 through FR-014 |

## 13.2 API Endpoints

| ID | Method | Route | Expected Behavior | Related Requirement |
| --- | --- | --- | --- | --- |
| API-001 | POST | /auth/login | Authenticate user by ID, create UserSession for machine, return RTMP URL + stream key + name + access token + refresh token | FR-001, FR-002, FR-017 |
| API-002 | POST | /auth/refresh | Accept refresh token, validate, create new access token in UserSession | FR-016 |
| API-003 | POST | /auth/logout | Delete UserSession for current machine, keep other sessions active | FR-013 |
| API-004 | GET | /user | Return user info, RTMP URL, stream key, name, allowed destinations | FR-003, FR-004, FR-005 |
| API-005 | GET | /destinations | Return list of allowed destinations for user | FR-005 |
| API-006 | PUT | /destinations/:id | Update stream key for destination | FR-006 |
| API-007 | PUT | /destinations/:id | Set or update stream key for destination; Go API starts forwarding if stream_key is present | FR-006, FR-007 |
| API-008 | POST | /stream/start | Receive stream start notification from SRS, log session | FR-009, FR-011 |
| API-009 | POST | /stream/stop | Receive stream stop notification from SRS, calculate duration, log session | FR-011 |
| API-010 | GET | /stream/status | Return current stream status (live/offline) plus bitrate warning if below threshold (fallback if SSE unavailable) | FR-010, FR-012 |
| API-011 | GET | /broadcasts | Return broadcast sessions for user, grouped by site and date | FR-011 |
| API-012 | PUT | /user/bitrate | Update user's configured upload bitrate for bitrate warning comparison | FR-012 |
| API-013 | GET | /user/stream/events | SSE endpoint — client subscribes, Go API pushes stream status events (type: "start"|"stop", status: "live"|"offline", bitrate_warning: bool) on stream lifecycle | FR-010, FR-012 |

---

# 14. Data Requirements

## 14.1 Data Entities

| Entity | Purpose |
| --- | --- |
| User | Stores user credentials, StreamLoft stream key, name |
| Destination | Global platform templates (YouTube, Twitch, Facebook) with name and RTMP URL |
| UserDestination | Links user to destination with their stream key. Assignment IS the permission — no separate enabled/disabled flag |
| BroadcastSession | Logs stream start/stop events per site with date and duration |
| UserMachine | Tracks which user was last logged in on each machine |
| UserSession | Stores per-machine authentication tokens - allows multiple devices per user |

## 14.2 Data Dictionary

**User Table**

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| numeric_id | VARCHAR(20) | Yes | User's login ID | Unique, numeric |
| name | VARCHAR(100) | Yes | User's display name | Non-empty |
| stream_key | VARCHAR(32) | Yes | StreamLoft stream key | Unique, generated |
| bitrate | INTEGER | No | User's configured upload bitrate (kbps) | Positive integer per VAL-007 |
| created_at | TIMESTAMP | Yes | Creation timestamp | Auto |
| updated_at | TIMESTAMP | Yes | Last update timestamp | Auto |

**Destination Table** (platform templates created by admin, one per external streaming service)

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| name | VARCHAR(100) | Yes | Platform name | Unique, non-empty |
| rtmp_url | VARCHAR(500) | Yes | Platform RTMP ingest URL | Valid RTMP URL format |
| created_at | TIMESTAMP | Yes | Creation timestamp | Auto |

**UserDestination Table** (links user to destination with their stream key — assignment IS the permission)

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| user_id | INTEGER | Yes | Foreign key to User | Required |
| destination_id | INTEGER | Yes | Foreign key to Destination | Required |
| stream_key | TEXT | No | User's stream key for this destination | Encrypted |
| created_at | TIMESTAMP | Yes | Creation timestamp | Auto |
| updated_at | TIMESTAMP | Yes | Last update timestamp | Auto |

**BroadcastSession Table** (all timestamps in UTC-5 timezone)

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| user_id | INTEGER | Yes | Foreign key to User | Required |
| user_destination_id | INTEGER | Yes | Foreign to UserDestination | Required |
| date | DATE | Yes | Broadcast date (UTC-5) | Required |
| duration_minutes | INTEGER | Yes | Session duration in minutes | Min 0 |
| started_at | TIMESTAMP | Yes | Stream start time (UTC-5) | Required |
| ended_at | TIMESTAMP | No | Stream end time (UTC-5) | Null if active |

**UserMachine Table**

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| machine_id | VARCHAR(255) | Yes | Machine identifier | Required |
| user_id | INTEGER | Yes | Foreign key to User | Required |
| last_used_at | TIMESTAMP | Yes | Last login timestamp | Auto |

**UserSession Table** (per-machine sessions - allows multiple devices per user)

| Field | Type | Required | Meaning | Validation Rule |
| --- | --- | --- | --- | --- |
| id | SERIAL | Yes | Primary key | Auto-increment |
| user_id | INTEGER | Yes | Foreign key to User | Required |
| machine_id | VARCHAR(255) | Yes | Machine identifier | Required |
| access_token | TEXT | Yes | Per-session access token (JWT) | Plain - HMAC verified |
| refresh_token | TEXT | Yes | Per-session refresh token | Plain - HMAC verified |
| token_expires_at | TIMESTAMP | Yes | Token expiration time | After 3 days |
| created_at | TIMESTAMP | Yes | Session start time | Auto |
| updated_at | TIMESTAMP | Yes | Last activity timestamp | Auto |

## 14.3 Data Storage Rules

| ID | Requirement |
| --- | --- |
| DR-001 | Destination stream keys shall be encrypted at rest |
| DR-002 | UserMachine table shall store last used user per machine |
| DR-003 | Admin assigns destinations to users by inserting into user_destinations |
| DR-004 | User configured bitrate shall be stored for comparison |
| DR-005 | Admin SQL files shall be stored in database/admin-sql/ folder |
| DR-006 | Admin SQL files shall be version-controlled in Git |
| DR-007 | Session tokens (access_token, refresh_token) stored as plain text (HMAC-verified, not encrypted) |
| DR-007 | Database name shall be "StreamLoft" |
| DR-008 | Database shall be hosted at 87.239.135.39:5432 |
| DR-009 | Go API runtime configuration shall be stored outside the database |
| DR-010 | Go API runtime configuration shall not be hardcoded in source code |
| DR-011 | The destination stream key encryption key shall be stored in /opt/StreamLoft/streamloft_master.key on the API VPS |
| DR-012 | The destination stream key encryption key shall not be stored in backend-api/.env, appsettings.json, the database, Git, SRS config, or the Windows app |

## 14.4 Data Integrity Rules

| ID | Rule |
| --- | --- |
| DI-001 | User numeric_id must be unique |
| DI-002 | StreamLoft stream_key must be unique |
| DI-003 | Foreign key constraints must enforce user-destination relationship |
| DI-004 | BroadcastSession allows multiple sessions per user per destination per day |
| DI-005 | user_destinations must have one row per user per destination |

---

# 15. Security Requirements

| ID | Requirement | Verification Method |
| --- | --- | --- |
| SEC-001 | PostgreSQL at 87.239.135.39:5432 shall only accept connections from API VPS IP | pg_hba.conf configuration |
| SEC-002 | Database shall not be publicly accessible (only API VPS allowed) | Firewall rules |
| SEC-003 | API shall require authentication token for user endpoints | Token validation in middleware |
| SEC-004 | Destination stream keys shall be encrypted in database | AES-256 encryption |
| SEC-005 | API shall validate all input before database operations | Input validation middleware |
| SEC-006 | Machine ID shall be used only for auto-login, not for authorization | Logic review |
| SEC-007 | Destination stream keys shall be stored encrypted in PostgreSQL | Encryption verification |
| SEC-008 | Encryption key shall not be stored in the database | Code review |
| SEC-009 | Encryption key shall not be committed to Git | Git configuration |
| SEC-010 | Encryption key shall be stored on API VPS in a protected local secret file | File system review |
| SEC-011 | Only the Go API process shall read the encryption key | Permission review |
| SEC-012 | Only the Go API shall decrypt destination stream keys | Code review |
| SEC-013 | The Windows app shall never receive full destination stream keys after they are saved | API response review |
| SEC-014 | Users shall be allowed to edit their destination stream keys via Windows app at any time | UI test |

---

# 16. Compliance Requirements

Use only when checking legal, institutional, rubric, security, privacy, or technical compliance.

| ID | Compliance Rule | Required System Behavior | Evidence Needed |
| --- | --- | --- | --- |
| COMP-001 | Streaming platform terms | Users must comply with YouTube/Twitch/FB terms | Documentation |

---

# 17. Non-Functional Requirements

## 17.1 Performance

| ID | Requirement | Measurement |
| --- | --- | --- |
| PERF-001 | The system shall support at least 10 simultaneous incoming streams | Load test |
| PERF-002 | The system shall support unlimited destinations per incoming stream | Configuration |
| PERF-003 | API response time shall be under 200ms for non-stream operations | Benchmark |
| PERF-004 | Stream latency shall be minimized based on bandwidth | Real-world testing |

## 17.2 Reliability

| ID | Requirement | Measurement |
| --- | --- | --- |
| REL-001 | The system shall log all stream start/stop events | Log verification |
| REL-002 | Database connection errors shall be retried with exponential backoff | Code review |
| REL-003 | Destination toggle shall not affect other users' streams | Integration test |

## 17.3 Maintainability

| ID | Requirement | Verification |
| --- | --- | --- |
| MNT-001 | The system shall detect incoming bitrate and compare to user configured | Testing with OBS |
| MNT-002 | The system shall warn when bitrate below 30% of configured | UI testing |
| MNT-003 | Admin SQL files shall be version-controlled | Git repository |
| MNT-004 | Logout shall only end current machine session | Integration test |

## 17.3.1 Configuration Management

| ID | Component | Config File/Location | Contents |
| --- | --- | --- | --- |
| CONFIG-001 | Go API | backend-api/.env.example | Example variable names only, no real secrets |
| CONFIG-002 | Go API runtime | /opt/StreamLoft/api.env | Database connection, server port, SRS callback settings |
| CONFIG-003 | Go API encryption | /opt/StreamLoft/streamloft_master.key | Destination stream key encryption key |
| CONFIG-004 | C# Windows App | windows-app/appsettings.json | API base URL only |
| CONFIG-005 | C# Windows App publish | Output folder | appsettings.json shall be copied beside .exe during publish |
| CONFIG-006 | SRS | media-server/srs.conf | RTMP port, HTTP callbacks to Go API |

## 17.4 Availability

| ID | Requirement | Measurement |
| --- | --- | --- |
| AVL-001 | The system shall show live/offline status in Windows app | UI verification |
| AVL-002 | SRS shall continue forwarding to other destinations if one fails | Error handling test |
| AVL-003 | Multiple users can login on same machine independently | Multi-user test |

---

# 18. SRS-API Integration Requirements

| ID | Requirement |
| --- | --- |
| SRS-API-001 | SRS shall be used as the RTMP ingest server for StreamLoft |
| SRS-API-002 | SRS shall notify the Go API of stream lifecycle events using HTTP callbacks (stream start/stop) |
| SRS-API-003 | The Go API shall not restart SRS when a user enables or disables a destination |
| SRS-API-004 | The Go API shall not rewrite global SRS forwarding configuration as the normal runtime control mechanism |
| SRS-API-005 | The Go API shall manage forwarding independently per user destination |
| SRS-API-006 | Each enabled user destination shall have an isolated forwarding worker |
| SRS-API-007 | Stopping, failing, enabling, or disabling one destination shall not interrupt other destinations for the same user |
| SRS-API-008 | Stopping, failing, enabling, or disabling one user's destination shall not interrupt any stream or destination belonging to another user |
| SRS-API-009 | When SRS sends an on_publish callback, the Go API shall mark the stream as live and start forwarding workers for the user's enabled destinations |
| SRS-API-010 | When SRS sends an on_unpublish callback, the Go API shall mark the stream as offline and stop forwarding workers for that user's active destinations |
| SRS-API-011 | When a user enables a destination while already live, the Go API shall start only that destination's forwarding worker |
| SRS-API-012 | When a user disables a destination while already live, the Go API shall stop only that destination's forwarding worker |
| SRS-API-013 | Go API shall push stream status events to subscribed Windows app clients via SSE when stream starts or stops |

---

# 19. LLM-Specific Requirements

This section has been moved to **docs/LLM_RULES.md**

---

# 20. Test Requirements

## 20.1 Test Case Format

| Field | Value |
| --- | --- |
| Test Case ID | TC-001 |
| Related Requirement | |
| Test Type | Functional / API / UI / Security / Regression / Compliance |
| Preconditions | |
| Test Data | |
| Steps | |
| Expected Result | |
| Pass / Fail Rule | |

## 20.2 Required Test Coverage

| Requirement ID | Required Test Type | Test Case ID |
| --- | --- | --- |
| FR-001 | Functional | TC-001 |
| FR-002 | Functional | TC-002 |
| FR-003 | UI | TC-003 |
| FR-004 | Functional | TC-004 |
| FR-005 | Functional | TC-005 |
| FR-006 | Functional | TC-006 |
| FR-007 | Integration | TC-007 |
| FR-008 | Integration | TC-008 |
| FR-009 | Integration | TC-009 |
| FR-010 | UI | TC-010 (API-013, SSE push) |
| FR-011 | Functional | TC-011 |
| FR-012 | Integration | TC-012 |
| FR-013 | Functional | TC-013 |
| FR-014 | Functional | TC-014 |
| SEC-001 | Security | Config check |
| SEC-002 | Security | Firewall check |
| SEC-003 | Security | Token validation |
| SEC-004 | Security | Encryption check |

---

# 20. Traceability Matrix

| Requirement ID | Business Rule | API / UI / Data Element | Test Case | Status |
| --- | --- | --- | --- | --- |
| FR-001 | BR-001 | API-001, Login Screen | TC-001 | Pending |
| FR-002 | BR-001 | API-001, UserMachine table | TC-002 | Pending |
| FR-003 | BR-001 | Welcome Screen | TC-003 | Pending |
| FR-004 | BR-002 | API-003, User table | TC-004 | Pending |
| FR-005 | BR-003 | API-004, AllowedDestination | TC-005 | Pending |
| FR-006 | BR-004 | API-005, Destination | TC-006 | Pending |
| FR-007 | BR-005 | API-006, SRS config | TC-007 | Pending |
| FR-008 | BR-005 | SRS forwarding | TC-008 | Pending |
| FR-009 | BR-006 | API-007, Stream events | TC-009 | Pending |
| FR-010 | BR-006 | API-010, API-013, Dashboard, SSE events | TC-010 | Pending |
| FR-011 | BR-007 | BroadcastSession table | TC-011 | Pending |
| FR-012 | BR-008 | Bitrate detection | TC-012 | Pending |
| FR-013 | BR-009 | API-002, Logout | TC-013 | Pending |
| FR-014 | BR-010 | Admin SQL files | TC-014 | Pending |

---

# 21. Acceptance Criteria

| ID | Acceptance Criterion |
| --- | --- |
| ACC-001 | Every mandatory functional requirement shall have at least one related test case |
| ACC-002 | Every security requirement shall have verification evidence |
| ACC-003 | Every validation rule shall have positive and negative test coverage |
| ACC-004 | Every API requirement shall define success and error behavior |
| ACC-005 | Components shall remain in separate folders per defined structure |
| ACC-006 | User logout shall only affect current machine session |
| ACC-007 | Destination toggle shall not affect other users' streams |
| ACC-008 | Stream sessions shall be logged per site with duration |
| ACC-009 | Stream status changes shall be pushed to Windows app via SSE with zero polling lag |

---

# 22. Excluded From LLM Review Context

These sections should not be sent to the LLM when the task is bug review, compliance review, test coverage, or implementation checking.

| Excluded Data | Reason |
| --- | --- |
| Document title | Does not define system behavior |
| Document version | Administrative metadata |
| Revision history | Administrative metadata |
| Author | Ownership metadata |
| Owner | Ownership metadata |
| Reviewed by | Approval metadata |
| Approved by | Approval metadata |
| Signature table | Approval metadata |
| Project sponsor | Administrative metadata |
| Approval dates | Administrative metadata |
| Change authors | Administrative metadata |
| Document status | Administrative metadata unless checking governance |
| Full release checklist | Not needed unless checking deployment readiness |
| Hardware interface section | Software interacts with network, not hardware |
| Backup section | Data recovery not part of initial review |
| Migration section | Exclude unless reviewing database migration behavior |

---

```text
Rule:
Only include SRS data that defines behavior, terminology, validation, security, data handling, interface behavior, error handling, constraints, tests, compliance, or pass/fail criteria.
```