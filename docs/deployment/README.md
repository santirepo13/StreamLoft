# StreamLoft Deployment Guide
## Server Components

### Media Server (172.86.73.79)
- SRS Media Server
- RTMP Port: 1935
- HTTP Callbacks: 172.86.73.79:8080

### Go API (172.86.73.79)
- Backend API Server
- Port: 8080
- Database: 87.239.135.39:5432

### Database (87.239.135.39:5432)
- PostgreSQL
- Accessible only from API server
