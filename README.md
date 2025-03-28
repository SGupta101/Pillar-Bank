# Pillar-Bank

A simple wire dashboard and API where users can:

- Login to view wire messages
- View paginated wire messages
- Submit new wire messages

## 🎥 Demo
[Watch the Demo](https://drive.google.com/file/d/15byFnmmt_zxys7GHvLFWKPkZw5r81Anw/view?usp=sharing)

## Setup

### Using Docker

```bash
# Clone the repository
git clone https://github.com/yourusername/pillar-bank.git
cd pillar-bank

# Start all services
docker compose up --build
```

Visit http://localhost:3000 and login with:

- Username: `user1`
- Password: `password1`

### Manual Setup

1. **Prerequisites**

   - Node.js 18+
   - Go 1.21+
   - PostgreSQL 15+

2. **Database Setup**

```bash
# Create databases
createdb pillar_bank
createdb pillar_bank_test  # for tests
```

3. **Run Backend**

```bash
cd backend
go mod download
go run main.go
```

4. **Run Frontend**

```bash
cd frontend
npm install
npm start
```

## Testing

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

## Project Structure

```
.
├── backend/
│   ├── auth/       # JWT authentication
│   ├── models/     # Data models
│   ├── testdata/   # Tests
│   └── main.go     # API endpoints
└── frontend/
    ├── src/
    │   ├── components/  # React components
    │   └── App.tsx     # Main app component
    └── package.json
```

## API Endpoints

- `POST /login` - User authentication
- `GET /wire-messages` - List wire messages (paginated)
- `POST /wire-messages` - Create new wire message
- `GET /wire-message/:seq` - Get specific wire message

## Technologies Used

- Cursor
- Frontend: React, TypeScript
- Backend: Go, Gin
- Database: PostgreSQL
- Authentication: JWT
- Containerization: Docker

## Future Improvements

- Add user registration
- Add logout functionality
- Add frontend tests
