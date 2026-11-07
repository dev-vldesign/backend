# Ventor API Documentation

This document describes the API endpoints for managing ventors (vendors/contractors) in the system.

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. Create Ventor
**POST** `/ventors`

Creates a new ventor in the system. Supports two types of ventors: contractors and suppliers. Accepts form data with optional file uploads.

#### Request Format
This endpoint accepts **multipart/form-data** with the following fields:

#### Form Fields
```
ventor_name: "New company" (required)
phone: "01111111111" (required)
email: "email@team.com" (required)
website: "http://localhost:5173/" (optional)
city: "Cairo" (required)
state: "Dar El-Salam" (required)
country: "Australia" (required)
role: "contractor" (required) - must be "contractor" or "supplier"
years_of_experience: "5" (required for contractors)
company_size: "51–200 employees" (required for contractors)
expected_budget: "7000 AUD" (required for contractors)
license_number: "N12342" (required)
license_expiration: "30/06/1998" (required)
ventor_type: "General Contractor" (required)
ventor_services: ["Ceramics", "Jewelry Making", "Pottery", "Leathercraft", "Textiles"] (required for contractors)
how_did_you_hear: "Pinterest" (optional)
logo: [file] (optional) - image file (JPEG, PNG, GIF, WEBP)
license: [file] (optional) - PDF file
```

#### File Upload Requirements

**Logo (Optional):**
- File type: Image files only (JPEG, PNG, GIF, WEBP)
- Field name: `logo`
- Stored in: `/uploads/ventors/logos/`

**License (Optional):**
- File type: PDF files only
- Field name: `license`
- Stored in: `/uploads/ventors/licenses/`

#### Example cURL Request
```bash
curl -X POST http://localhost:8080/ventors \
  -F "ventor_name=New company" \
  -F "phone=01111111111" \
  -F "email=email@team.com" \
  -F "website=http://localhost:5173/" \
  -F "city=Cairo" \
  -F "state=Dar El-Salam" \
  -F "country=Australia" \
  -F "role=contractor" \
  -F "years_of_experience=5" \
  -F "company_size=51–200 employees" \
  -F "expected_budget=7000 AUD" \
  -F "license_number=N12342" \
  -F "license_expiration=30/06/1998" \
  -F "ventor_type=General Contractor" \
  -F "ventor_services=Ceramics" \
  -F "ventor_services=Jewelry Making" \
  -F "ventor_services=Pottery" \
  -F "how_did_you_hear=Pinterest" \
  -F "logo=@/path/to/logo.png" \
  -F "license=@/path/to/license.pdf"
```

#### Field Descriptions

**Required for all ventors:**
- `ventor_name` (string): Name of the ventor/company
- `phone` (string): Contact phone number
- `email` (string): Contact email address (must be unique)
- `city` (string): City location
- `state` (string): State/province location
- `country` (string): Country location
- `role` (string): Must be either "contractor" or "supplier"
- `license_number` (string): Business license number
- `license_expiration` (string): License expiration date
- `ventor_type` (string): Type of ventor (e.g., "General Contractor", "Supplier")

**Optional for all ventors:**
- `website` (string): Company website URL
- `how_did_you_hear` (string): How the ventor heard about the platform
- `logo` (file): Company logo image file
- `license` (file): License document PDF file

**Required for contractors only:**
- `years_of_experience` (string): Years of experience in the field
- `company_size` (string): Size of the company (e.g., "51–200 employees")
- `expected_budget` (string): Expected budget range
- `ventor_services` (array): Array of services offered

**Not applicable for suppliers:**
- Contractors have access to company information fields
- Suppliers do not need years of experience, company size, or expected budget
- Ventor services are only for contractors

#### Response
**Success (201 Created):**
```json
{
    "message": "Ventor created successfully",
    "ventor": {
        "id": "507f1f77bcf86cd799439011",
        "ventor_name": "New company",
        "phone": "01111111111",
        "email": "email@team.com",
        "website": "http://localhost:5173/",
        "city": "Cairo",
        "state": "Dar El-Salam",
        "country": "Australia",
        "role": "contractor",
        "logo": "/uploads/ventors/logos/logo_507f1f77bcf86cd799439011.png",
        "license": "/uploads/ventors/licenses/license_507f1f77bcf86cd799439011.pdf",
        "years_of_experience": "5",
        "company_size": "51–200 employees",
        "expected_budget": "7000 AUD",
        "license_number": "N12342",
        "license_expiration": "30/06/1998",
        "ventor_type": "General Contractor",
        "ventor_services": [
            "Ceramics",
            "Jewelry Making",
            "Pottery",
            "Leathercraft",
            "Textiles"
        ],
        "how_did_you_hear": "Pinterest",
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
    }
}
```

**Error (400 Bad Request):**
```json
{
    "error": "Years of experience is required for contractors"
}
```

**Error (400 Bad Request - File Type):**
```json
{
    "error": "Logo must be an image file (JPEG, PNG, GIF, WEBP)",
    "received_type": "text/plain"
}
```

**Error (409 Conflict):**
```json
{
    "error": "A ventor with this email already exists"
}
```

### 2. Get All Ventors
**GET** `/ventors`

Retrieves all ventors from the database.

#### Response
**Success (200 OK):**
```json
{
    "ventors": [
        {
            "id": "507f1f77bcf86cd799439011",
            "ventor_name": "New company",
            "phone": "01111111111",
            "email": "email@team.com",
            "role": "contractor",
            "logo": "/uploads/ventors/logos/logo_507f1f77bcf86cd799439011.png",
            "license": "/uploads/ventors/licenses/license_507f1f77bcf86cd799439011.pdf",
            // ... other fields
        }
    ],
    "count": 1
}
```

### 3. Get Ventor by ID
**GET** `/ventors/:id`

Retrieves a specific ventor by their ID.

#### Parameters
- `id` (string): The ventor's unique identifier

#### Response
**Success (200 OK):**
```json
{
    "ventor": {
        "id": "507f1f77bcf86cd799439011",
        "ventor_name": "New company",
        "phone": "01111111111",
        "email": "email@team.com",
        "role": "contractor",
        "logo": "/uploads/ventors/logos/logo_507f1f77bcf86cd799439011.png",
        "license": "/uploads/ventors/licenses/license_507f1f77bcf86cd799439011.pdf",
        // ... other fields
    }
}
```

**Error (404 Not Found):**
```json
{
    "error": "Ventor not found"
}
```

## File Access

Uploaded files can be accessed via the static file server:
- Logo: `http://localhost:8080/uploads/ventors/logos/[filename]`
- License: `http://localhost:8080/uploads/ventors/licenses/[filename]`

## Validation Rules

1. **Email uniqueness**: Each ventor must have a unique email address
2. **Role validation**: Role must be either "contractor" or "supplier"
3. **Contractor requirements**: If role is "contractor", the following fields are required:
   - `years_of_experience`
   - `company_size`
   - `expected_budget`
   - `ventor_services` (must not be empty)
4. **Email format**: Email must be in valid format
5. **Required fields**: All marked fields are required for the respective role
6. **File validation**:
   - Logo must be an image file (JPEG, PNG, GIF, WEBP)
   - License must be a PDF file
   - Both files are optional

## Error Codes

- `400`: Bad Request (validation errors, invalid input, invalid file types)
- `404`: Not Found (ventor not found)
- `409`: Conflict (email already exists)
- `500`: Internal Server Error (database errors, file system errors) 