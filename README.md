# Pillar-Bank

At Pillar Bank N.A. you've been tasked to build a wire dashboard and the subsequent APIs that format and save wire messages.

## Setup

### Prerequisites

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
