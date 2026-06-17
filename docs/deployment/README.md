# StreamLoft Deployment Guide
## Server Components

### Media Server (74.81.35.20)
- SRS Media Server
- RTMP Port: 10004
- HTTP Callbacks: 74.81.35.20:10002

### Go API (74.81.35.20)
- Backend API Server
- Port: 10002
- Database: 87.239.135.39:5432

### Database (87.239.135.39:5432)
- PostgreSQL
- Accessible only from API server
