# SMEfin Backend API

A Go backend API for the SMEfin application that allows SMEs to register with their details and request financing.

## Features

- **Authentication**: OTP-based authentication with JWT tokens
- **User Registration**: Multi-step registration process
  - Personal Details
  - Business Details
  - Trade License Upload
- **Account Status**: Track new/old account status based on completion
- **Financing Requests**: Submit and manage financing requests (requires completed registration)
- **Database**: PostgreSQL (Supabase) integration
- **Error Handling**: Comprehensive error responses with status codes
- **Form Data Support**: All endpoints accept `multipart/form-data` (with JSON fallback for backward compatibility)
- **File Upload**: Support for direct file uploads in trade license endpoints

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

# Supabase Storage Configuration (for file uploads)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
# Alternative: SUPABASE_ANON_KEY=your-anon-key (if bucket is public)
SUPABASE_BUCKET_NAME=vercel_bucket
```

## Database Setup

### Using Supabase

1. Create a new Supabase project
2. Run the SQL migration files in order:
   - `supabase/migrations/001_initial_schema.sql` (users, registration tables)
   - `supabase/migrations/002_financing_requests.sql` (financing requests table)
3. Update your `.env` file with the Supabase connection details

**Note:** The database connection supports multiple environment variable formats:
- `DATABASE_URL` (preferred for Supabase/Vercel)
- Individual variables: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Alternative naming: `POSTGRES_URL`, `POSTGRES_HOST`, etc.

The connection code automatically detects and uses the available format.

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
Content-Type: multipart/form-data

Form Data:
- email: user@example.com
```

**Note:** All endpoints support both `multipart/form-data` and `application/json` formats. JSON format is also accepted for backward compatibility.

#### Verify OTP
```
POST /api/auth/verify-otp
Content-Type: multipart/form-data

Form Data:
- email: user@example.com
- otp: 123456

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

Response:
{
    "success": true,
    "message": "Account status retrieved successfully",
    "status_code": 200,
    "data": {
        "user_id": "uuid",
        "email": "user@example.com",
        "status": "new",
        "has_personal_details": true,
        "has_business_details": true,
        "has_trade_license": false,
        "is_complete": false
    }
}
```

#### Get User Data
```
GET /api/user/data
Authorization: Bearer <token>

Response:
{
    "success": true,
    "message": "User data retrieved successfully",
    "status_code": 200,
    "data": {
        "user_id": "uuid",
        "email": "user@example.com",
        "status": "old",
        "personal": {
            "id": "uuid",
            "user_id": "uuid",
            "full_name": "John Smith",
            "email": "john@example.com",
            "phone_number": "(100) 123456789",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "business": {
            "id": "uuid",
            "user_id": "uuid",
            "business_name": "ABC Company",
            "trade_license_number": "TL123456789",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "trade_license": {
            "id": "uuid",
            "user_id": "uuid",
            "filename": "license.pdf",
            "file_url": "https://supabase.co/storage/...",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        }
    }
}
```

**Note:** Returns `null` for `personal`, `business`, or `trade_license` if not yet saved.

#### Save Full Registration
```
POST /api/user/full-registration
Authorization: Bearer <token>
Content-Type: multipart/form-data

Form Data (nested format):
- personal[full_name]: Muntasir Efaz
- personal[email]: efaz@example.com
- personal[phone_number]: (+880) 123456789
- business[business_name]: ABC Company
- business[trade_license_number]: TL123456789
- trade[filename]: license.pdf
- trade[file_url]: https://example.com/storage/license.pdf
- trade[file]: [file upload] (optional - alternative to trade[file_url])

Alternative flat format (also supported):
- full_name: Muntasir Efaz
- email: efaz@example.com
- phone_number: (+880) 123456789
- business_name: ABC Company
- trade_license_number: TL123456789
- filename: license.pdf
- file_url: https://example.com/storage/license.pdf
```

**Note:** This endpoint saves personal details, business details, and trade license in a single API call. If a file is uploaded via `trade[file]`, it will be automatically uploaded to Supabase storage.

#### Request Financing
```
POST /api/financing/request
Authorization: Bearer <token>
Content-Type: multipart/form-data

Form Data:
- amount: 50000 (required, positive number)
- purpose: Business expansion and inventory purchase (required)
- repayment_period: 12 (required, number of months)

Response:
{
    "success": true,
    "message": "Financing request submitted successfully",
    "status_code": 201,
    "data": {
        "id": "uuid",
        "user_id": "uuid",
        "amount": 50000,
        "purpose": "Business expansion and inventory purchase",
        "repayment_period": 12,
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

**Note:** 
- User must have completed registration (status: "old") before requesting financing.
- Users can submit multiple financing requests. Each request is stored separately.

#### Get All Financing Requests
```
GET /api/financing/requests
Authorization: Bearer <token>

Response:
{
    "success": true,
    "message": "Financing requests retrieved successfully",
    "status_code": 200,
    "data": [
        {
            "id": "uuid",
            "user_id": "uuid",
            "amount": 50000,
            "purpose": "Business expansion",
            "repayment_period": 12,
            "status": "pending",
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        {
            "id": "uuid-2",
            "user_id": "uuid",
            "amount": 30000,
            "purpose": "Equipment purchase",
            "repayment_period": 6,
            "status": "approved",
            "created_at": "2024-01-15T00:00:00Z",
            "updated_at": "2024-01-16T00:00:00Z"
        }
    ]
}
```

**Note:** Returns all financing requests for the authenticated user, ordered by creation date (newest first). Returns an empty array `[]` if no requests exist.

#### Get Financing Request Detail
```
GET /api/financing/request-detail?id=<request_id>
Authorization: Bearer <token>

Response:
{
    "success": true,
    "message": "Financing request retrieved successfully",
    "status_code": 200,
    "data": {
        "id": "uuid",
        "user_id": "uuid",
        "amount": 50000,
        "purpose": "Business expansion",
        "repayment_period": 12,
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}
```

#### Get Latest Financing Request
```
GET /api/financing/latest
Authorization: Bearer <token>

Response (if request exists):
{
    "success": true,
    "message": "Latest financing request retrieved successfully",
    "status_code": 200,
    "data": {
        "id": "uuid",
        "user_id": "uuid",
        "amount": 50000,
        "purpose": "Business expansion",
        "repayment_period": 12,
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
    }
}

Response (if no request exists):
{
    "success": true,
    "message": "No financing request found",
    "status_code": 200,
    "data": null
}
```

**Note:** Returns the most recent financing request for the user, or `null` if no requests exist.

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
- **"new"**: Account exists but registration is incomplete (missing personal details, business details, or trade license)
- **"old"**: All three steps are completed (personal details, business details, and trade license)

## API Summary

The API provides the following endpoints:

### User Management:
1. **GET /api/user/status** - Get account completion status
2. **GET /api/user/data** - Get all user registration data (personal, business, trade license)
3. **POST /api/user/full-registration** - Save all registration data in one call

### Financing:
1. **POST /api/financing/request** - Submit a financing request (requires completed registration)
2. **GET /api/financing/requests** - Get all financing requests for the user
3. **GET /api/financing/request-detail?id=<id>** - Get details of a specific financing request
4. **GET /api/financing/latest** - Get the latest financing request (returns null if none exists)

All user data is saved through the single `full-registration` endpoint, which handles personal details, business details, and trade license upload in one request.

### Financing Request Status
- **"pending"**: Request submitted, awaiting review
- **"approved"**: Request approved by admin
- **"rejected"**: Request rejected
- **"disbursed"**: Funds have been disbursed

## Postman Collection

Import the `postman_collection.json` file into Postman to test all endpoints. The collection is configured to use form-data format. Make sure to set the `base_url` variable to your server URL (default: `https://sm-efin-backend.vercel.app`).

**Note:** The Postman collection uses `multipart/form-data` format. All requests are pre-configured with the correct form fields.

## Testing

### Default OTP
For development and testing, the default OTP is `123456`. This can be configured via the `DEFAULT_OTP` environment variable.

### Request Format
All POST endpoints accept data in `multipart/form-data` format. JSON format (`application/json`) is also supported for backward compatibility. The API automatically detects the content type and parses accordingly.

### Form Field Naming
The API supports multiple naming conventions:
- **Nested format**: `personal[full_name]`, `business[business_name]`
- **Flat format**: `personal_full_name`, `business_business_name`
- **Simple format**: `full_name`, `email` (for full-registration endpoint)

## Deployment to Vercel

Since Vercel primarily supports serverless functions, you'll need to:

1. Use Vercel's Go runtime
2. Create a `vercel.json` configuration file
3. Set environment variables in Vercel dashboard
4. Ensure your Supabase database is accessible from Vercel

### Vercel Configuration

The `vercel.json` file is already configured:

```json
{
  "version": 2,
  "builds": [
    {
      "src": "api/index.go",
      "use": "@vercel/go"
    }
  ],
  "routes": [
    {
      "src": "/(.*)",
      "dest": "/api/index.go"
    }
  ]
}
```

**Important:** Make sure to set all required environment variables in your Vercel project settings:
- Database connection variables (`DATABASE_URL` or individual `DB_*` variables)
- `JWT_SECRET`
- `JWT_EXPIRY_HOURS` (optional, defaults to 24)
- `DEFAULT_OTP` (optional, defaults to 123456)

## Project Structure

```
sme_fin_backend/
├── api/
│   └── index.go           # Vercel serverless function entry point
├── database/
│   └── db.go              # Database connection
├── handlers/
│   ├── auth.go            # Authentication handlers
│   ├── user.go            # User handlers
│   └── financing.go       # Financing request handlers
├── middleware/
│   └── auth.go            # JWT authentication middleware
├── models/
│   └── user.go            # Database models and methods
├── utils/
│   ├── jwt.go             # JWT utilities
│   ├── response.go        # Response helpers
│   ├── validator.go       # Validation utilities
│   └── formdata.go        # Form data parsing utilities
├── supabase/
│   └── migrations/
│       ├── 001_initial_schema.sql
│       └── 002_financing_requests.sql
├── main.go                # Local development entry point
├── go.mod                 # Go dependencies
├── vercel.json            # Vercel deployment configuration
├── postman_collection.json # Postman API collection
└── README.md              # This file
```

## Error Codes

- `200`: Success
- `201`: Created (resource created successfully)
- `400`: Bad Request (validation errors, missing fields)
- `401`: Unauthorized (invalid/missing token, invalid OTP)
- `403`: Forbidden (unauthorized to access resource)
- `404`: Not Found (user/resource not found)
- `500`: Internal Server Error (database errors, server errors)

## License

This project is part of the SMEfin application.

