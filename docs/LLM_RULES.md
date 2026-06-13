# StreamLoft LLM Rules

These rules define how an LLM is allowed to edit and work with the StreamLoft project.

---

## 1. Folder Structure Rules

| Rule | Description |
| --- | --- |
| FOLDER-001 | The project shall have separate folders for each component: backend-api/, windows-app/, media-server/, database/, docs/ |
| FOLDER-002 | Go API code shall only be in backend-api/ folder |
| FOLDER-003 | C# Windows app code shall only be in windows-app/ folder |
| FOLDER-004 | SRS configuration shall only be in media-server/ folder |
| FOLDER-005 | Admin SQL files shall only be in database/admin-sql/ folder |
| FOLDER-006 | When editing one component, do not modify files in other component folders |

---

## 2. Coding Standards

### 2.1 Go API Rules

| Rule | Description |
| --- | --- |
| GO-001 | Follow Go conventions and idioms |
| GO-002 | Use go modules for dependency management |
| GO-003 | Implement proper error handling |
| GO-004 | Use environment variables for configuration, never hardcode secrets |

### 2.2 C# Windows App Rules

| Rule | Description |
| --- | --- |
| CS-001 | Follow .NET patterns and best practices |
| CS-002 | Use proper namespace organization |
| CS-003 | Implement MVVM pattern where appropriate |
| CS-004 | Never expose stream keys in UI - hide them after save |

### 2.3 SRS Configuration Rules

| Rule | Description |
| --- | --- |
| SRS-001 | Follow SRS documentation standards |
| SRS-002 | Configure HTTP callbacks for stream events |
| SRS-003 | Do not use global forwarding as runtime control mechanism |

---

## 3. Database Rules

| Rule | Description |
| --- | --- |
| DB-001 | Use only the tables defined in SRS.md Section 14.2 |
| DB-002 | Do not create additional tables without user approval |
| DB-003 | Store encrypted stream keys for destinations |
| DB-004 | Never commit encryption keys to Git |
| DB-005 | Use database/admin-sql/ for admin SQL files |

---

## 4. API Rules

| Rule | Description |
| --- | --- |
| API-001 | Use only the endpoints defined in SRS.md Section 13.2 |
| API-002 | Do not create additional endpoints without user approval |
| API-003 | Require authentication token for user endpoints |
| API-004 | Never return full destination stream keys in API responses to Windows app |

---

## 5. UI Rules

| Rule | Description |
| --- | --- |
| UI-001 | Use only the screens defined in SRS.md Section 12.1 |
| UI-002 | Implement login, welcome, dashboard, destination, and events screens |
| UI-003 | Show live/offline status indicator |
| UI-004 | Allow logout from current machine only |

---

## 6. SRS Integration Rules

| Rule | Description |
| --- | --- |
| INT-001 | SRS shall notify Go API via HTTP callbacks |
| INT-002 | Go API shall manage independent forwarding workers per destination |
| INT-003 | Do not restart SRS for destination toggle changes |
| INT-004 | Do not rewrite global SRS config for runtime control |
| INT-005 | Each user's destination toggle shall not affect other users |

---

## 7. Development Workflow Rules

| Rule | Description |
| --- | --- |
| WORK-001 | Always verify changes against SRS.md before finalizing |
| WORK-002 | When unsure about a requirement, ask user before implementing |
| WORK-003 | Use clear comments explaining the "why" not the "what" |
| WORK-004 | Keep components in separate folders per defined structure |
| WORK-005 | Test changes before committing |

---

## 8. Hallucination Prevention Rules

| Rule | Description |
| --- | --- |
| HAL-001 | Do not fabricate database schema not defined in SRS.md |
| HAL-002 | Do not create API endpoints not defined in SRS.md |
| HAL-003 | Do not create UI screens not defined in SRS.md |
| HAL-004 | State when required information is unavailable |
| HAL-005 | Always refer to glossary terms from SRS.md Section 3 |

---

## 9. Security Rules

| Rule | Description |
| --- | --- |
| SEC-001 | Never commit secrets, keys, or credentials to Git |
| SEC-002 | Store encryption key on API VPS in protected local file |
| SEC-003 | Only Go API process shall read encryption key |
| SEC-004 | Encrypt destination stream keys at rest |
| SEC-005 | Validate all user input before database operations |
| SEC-006 | PostgreSQL shall only accept connections from API VPS IP |

---

## 10. Implementation Verification Checklist

Before any implementation, verify:

- [ ] Changes align with SRS.md requirements
- [ ] No new files created in wrong folders
- [ ] No new tables without user approval
- [ ] No new API endpoints without user approval
- [ ] No new UI screens without user approval
- [ ] Secrets not committed to Git
- [ ] Code follows language-specific conventions (Go for API, C# for Windows app)