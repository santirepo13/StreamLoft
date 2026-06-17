# StreamLoft API Documentation
## Base URL: http://172.86.73.79:10002

### Authentication Endpoints
- POST /auth/login
- POST /auth/refresh
- POST /auth/logout

### User Management
- GET /user
- GET /destinations
- PUT /destinations/:id
- PUT /destinations/:id/toggle

### Stream Management
- POST /stream/start (SRS callback)
- POST /stream/stop (SRS callback)
- GET /stream/status
- GET /broadcasts
