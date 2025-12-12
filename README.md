# SMEfin Backend API

A Go backend API for the SMEfin application that allows SMEs to register with their details and request financing.

## Features

- **Authentication**: OTP-based authentication with JWT tokens
- **User Registration**: Multi-step registration process
  - Personal Details
  - Business Details
  - Trade License Upload
- **Account Status**: Track new/old account status based on completion
- **Database**: PostgreSQL (Supabase) integration
- **Error Handling**: Comprehensive error responses with status codes

## Prerequisites

- Go 1.21 or higher
- PostgreSQL database (Supabase)

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Database Configuration
DB_HOST=your-supabase-host.supabase.co
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=postgres
DB_SSLMODE=require

# JWT Configuration
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRY_HOURS=24

# Server Configuration
PORT=8080

# Default OTP (for development/testing)
DEFAULT_OTP=123456
```

## Database Setup

### Using Supabase

1. Create a new Supabase project
2. Run the SQL migration file located at `supabase/migrations/001_initial_schema.sql` in your Supabase SQL editor
3. Update your `.env` file with the Supabase connection details

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd sme_fin_backend
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables (copy `.env.example` to `.env` and update values)

4. Run database migrations (if using Supabase, run the SQL file in Supabase SQL editor)

## Running the Application

### Local Development

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Public Endpoints

#### Health Check
```
GET /health
```

#### Send OTP
```
POST /api/auth/send-otp
Content-Type: application/json

{
    "email": "user@example.com"
}
```

#### Verify OTP
```
POST /api/auth/verify-otp
Content-Type: application/json

{
    "email": "user@example.com",
    "otp": "123456"
}

Response:
{
    "success": true,
    "message": "OTP verified successfully",
    "status_code": 200,
    "data": {
        "token": "jwt_token_here",
        "user_id": "uuid",
        "email": "user@example.com",
        "account_status": "new"
    }
}
```

### Protected Endpoints (Require JWT Token)

All protected endpoints require the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

#### Get Account Status
```
GET /api/user/status
Authorization: Bearer <token>
```

#### Save Personal Details
```
POST /api/user/personal-details
Authorization: Bearer <token>
Content-Type: application/json

{
    "full_name": "Muntasir Efaz",
    "email": "efaz@example.com",
    "phone_number": "(+880) 123456789"
}
```

#### Save Business Details
```
POST /api/user/business-details
Authorization: Bearer <token>
Content-Type: application/json

{
    "business_name": "ABC Trading Company",
    "trade_license_number": "TL123456789"
}
```

#### Upload Trade License
```
POST /api/user/trade-license
Authorization: Bearer <token>
Content-Type: application/json

{
    "filename": "license.pdf",
    "file_url": "https://example.com/storage/license.pdf"
}
```

#### Save Full Registration (single call)
```
POST /api/user/full-registration
Authorization: Bearer <token>
Content-Type: application/json

{
    "personal": {
        "full_name": "Muntasir Efaz",
        "email": "efaz@example.com",
        "phone_number": "(+880) 123456789"
    },
    "business": {
        "business_name": "ABC Company",
        "trade_license_number": "TL123456789"
    },
    "trade": {
        "filename": "license.pdf",
        "file_url": "https://example.com/storage/license.pdf"
    }
}
```

#### Submit Registration
```
POST /api/user/submit
Authorization: Bearer <token>
```

## Response Format

All API responses follow this format:

### Success Response
```json
{
    "success": true,
    "message": "Operation successful",
    "status_code": 200,
    "data": { ... }
}
```

### Error Response
```json
{
    "success": false,
    "message": "Error message",
    "status_code": 400
}
```

## Account Status

The account status is determined as follows:
- **"new"**: Account exists but registration is incomplete
- **"old"**: All three steps are completed (personal details, business details, and trade license)

## Postman Collection

Import the `postman_collection.json` file into Postman to test all endpoints. Make sure to set the `base_url` variable to your server URL.

## Testing

### Default OTP
For development and testing, the default OTP is `123456`. This can be configured via the `DEFAULT_OTP` environment variable.

## Deployment to Vercel

Since Vercel primarily supports serverless functions, you'll need to:

1. Use Vercel's Go runtime
2. Create a `vercel.json` configuration file
3. Set environment variables in Vercel dashboard
4. Ensure your Supabase database is accessible from Vercel

### Vercel Configuration

Create a `vercel.json` file:

```json
{
  "version": 2,
  "builds": [
    {
      "src": "main.go",
      "use": "@vercel/go"
    }
  ],
  "routes": [
    {
      "src": "/(.*)",
      "dest": "/main.go"
    }
  ]
}
```

## Project Structure

```
sme_fin_backend/
├── database/
│   └── db.go              # Database connection
├── handlers/
│   ├── auth.go            # Authentication handlers
│   └── user.go            # User handlers
├── middleware/
│   └── auth.go            # JWT authentication middleware
├── models/
│   └── user.go            # Database models and methods
├── utils/
│   ├── jwt.go             # JWT utilities
│   ├── response.go        # Response helpers
│   └── validator.go       # Validation utilities
├── supabase/
│   └── migrations/
│       └── 001_initial_schema.sql
├── main.go                # Application entry point
├── go.mod                 # Go dependencies
├── postman_collection.json # Postman API collection
└── README.md              # This file
```

## Error Codes

- `200`: Success
- `400`: Bad Request (validation errors, missing fields)
- `401`: Unauthorized (invalid/missing token, invalid OTP)
- `404`: Not Found (user/resource not found)
- `500`: Internal Server Error (database errors, server errors)

## License

This project is part of the SMEfin application.

