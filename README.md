# Pillar-Bank

A simple wire dashboard and API for Pillar Bank N.A, where a user can login, see all wire messages, and submit new wire messages.

## Setup

1. Install PostgreSQL if you haven't already:

   ```bash
   # For MacOS using Homebrew
   brew install postgresql

   # For Ubuntu/Debian
   sudo apt-get install postgresql
   ```

2. Create a new database:

   ```bash
   # Start PostgreSQL service
   # MacOS:
   brew services start postgresql
   # Ubuntu:
   sudo service postgresql start

   # Create the database
   createdb pillar_bank
   createdb pillar_bank_test  # for running tests
   ```

3. run the backend server

   ```bash
   go run backend/main.go
   ```

4. run the frontend server

   ```bash
   cd frontend
   npm install
   npm start
   ```

## Testing

To run the tests for the backend, run the following command:

```bash
go test
```

To run the tests for the auth, run the following command:

```bash
cd auth
go test
```

To run the tests for the frontend, run the following command:

```bash
npm test
```

## Notes

- The backend is using a JWT token to authenticate requests.
- The frontend is using a cookie to store the JWT token.
- The backend is using a PostgreSQL database to store the wire messages.
- The frontend is using a React frontend to display the wire messages.

## Tools used to write submission

- Cursor
- Tutorials/Articles
  - https://go.dev/doc/tutorial/web-service-gin
  - https://www.calhoun.io/using-postgresql-with-go/
  - https://go.dev/doc/tutorial/add-a-test
  - https://permify.co/post/jwt-authentication-go/
  - https://www.geeksforgeeks.org/how-to-use-react-with-typescript/

## Future Improvements

- Add testing for the frontend
- Add logout/signup functionality
