# Intelligence Query Engine — HNG Stage 2 Backend

A high‑performance queryable intelligence engine built with Go. It provides advanced filtering, sorting, pagination, and a rule‑based natural language search over a seeded demographic profile database.

## Live API Base URL

`https://stage-2-intelligence-query-engine.vercel.app/`


## Table of Contents

- [Features](#features)
- [Tech Stack](#tech-stack)
- [API Endpoints](#api-endpoints)
- [Natural Language Query Rules](#natural-language-query-rules)
- [Running Locally](#running-locally)
- [Database Seeding](#database-seeding)
- [Environment Variables](#environment-variables)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Project Structure](#project-structure)
- [Author](#author)

## Features

- **Advanced Filtering** – Filter by `gender`, `age_group`, `country_id`, `min_age`, `max_age`, `min_gender_probability`, `min_country_probability`.
- **Sorting** – Sort by `age`, `created_at`, or `gender_probability` in ascending or descending order.
- **Pagination** – Page through results with configurable `page` and `limit` (max 50).
- **Natural Language Search** – Interpret plain English queries like `"young males from nigeria"` using rule‑based parsing.
- **Idempotent Seeding** – Pre‑populate the database with 2,026 profiles without duplicates.
- **Structured Error Responses** – Consistent JSON error format with appropriate HTTP status codes.
- **CORS Support** – Allows cross‑origin requests from any domain.

## Tech Stack

| Component           | Technology                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| Language            | Go 1.22+                                                                    |
| Database            | PostgreSQL (Supabase / local)                                               |
| Database Driver     | [pgx/v5](https://github.com/jackc/pgx)                                      |
| UUID                | UUID v7 ([google/uuid](https://github.com/google/uuid))                     |
| Environment Config  | [godotenv](https://github.com/joho/godotenv)                                |
| Web Server          | Net/http (standard library)                                                 |

## API Endpoints

### `GET /api/profiles`

Retrieve profiles with optional filtering, sorting, and pagination.

**Query Parameters**

| Parameter                 | Type    | Description                                                       | Example        |
|---------------------------|---------|-------------------------------------------------------------------|----------------|
| `gender`                  | string  | Filter by gender (`male` / `female`)                              | `male`         |
| `age_group`               | string  | Filter by age group (`child`, `teenager`, `adult`, `senior`)      | `adult`        |
| `country_id`              | string  | Filter by 2‑letter ISO country code                               | `NG`           |
| `min_age`                 | integer | Minimum age (inclusive)                                           | `25`           |
| `max_age`                 | integer | Maximum age (inclusive)                                           | `50`           |
| `min_gender_probability`  | float   | Minimum gender confidence score                                    | `0.7`          |
| `min_country_probability` | float   | Minimum country confidence score                                   | `0.5`          |
| `sort_by`                 | string  | Field to sort by (`age`, `created_at`, `gender_probability`)      | `age`          |
| `order`                   | string  | Sort order (`asc` / `desc`)                                       | `desc`         |
| `page`                    | integer | Page number (default: `1`)                                        | `2`            |
| `limit`                   | integer | Results per page (default: `10`, max: `50`)                       | `20`           |

**Example Request**

```bash
curl "http://localhost:9000/api/profiles?gender=female&min_age=25&country_id=NG&sort_by=age&order=desc&page=1&limit=10"
```

**Example Response**

```json
{
  "status": "success",
  "page": 1,
  "limit": 10,
  "total": 142,
  "data": [
    {
      "id": "019d9b1d-39e4-7b10-b24b-ae213fccdacf",
      "name": "Aisha Bello",
      "gender": "female",
      "gender_probability": 0.88,
      "age": 34,
      "age_group": "adult",
      "country_id": "NG",
      "country_name": "Nigeria",
      "country_probability": 0.72,
      "created_at": "2026-04-15T12:00:00Z"
    }
  ]
}
```

### `GET /api/profiles/search`

Natural language query endpoint. Accepts a `q` parameter with a plain English phrase.

**Query Parameter**

| Parameter | Type   | Description                       | Example                           |
|-----------|--------|-----------------------------------|-----------------------------------|
| `q`       | string | Natural language search phrase    | `young males from nigeria`        |
| `page`    | integer| Page number (default: `1`)        | `1`                               |
| `limit`   | integer| Results per page (max: `50`)      | `10`                              |

**Example Request**

```bash
curl "http://localhost:9000/api/profiles/search?q=young+males+from+nigeria&page=1&limit=10"
```

**Example Response**

Same paginated format as `/api/profiles`.

## Natural Language Query Rules

The parser is **rule‑based** (no AI/LLM). It maps common phrases to filters.

| Phrase                                | Generated Filters                                                 |
|---------------------------------------|-------------------------------------------------------------------|
| `young males`                         | `gender=male`, `min_age=16`, `max_age=24`                         |
| `females above 30`                    | `gender=female`, `min_age=30`                                     |
| `people from angola`                  | `country_id=AO`                                                   |
| `adult males from kenya`              | `gender=male`, `age_group=adult`, `country_id=KE`                 |
| `male and female teenagers above 17`  | `age_group=teenager`, `min_age=17`                                |

> **Note:** `"young"` maps to ages **16–24** for parsing purposes only; it is not a stored age group.

If the phrase cannot be interpreted, the endpoint returns:

```json
{ "status": "error", "message": "Unable to interpret query" }
```

## Running Locally

### Prerequisites

- Go 1.22+
- PostgreSQL 14+ (or a Supabase project)


### Steps

1. **Clone the repository**

   ```bash
   git clone https://github.com/LunarKhord/stage-2-intelligence-query-engine.git
   cd stage-2-intelligence-query-engine
   ```

2. **Install dependencies**

   ```bash
   go mod tidy
   ```

3. **Set up environment variables**

   Create a `.env` file in the project root:

   ```
   DATABASE_URL=postgresql://user:password@localhost:5432/profiles_db?sslmode=disable
   PORT=9000
   ```

4. **Run database migrations (if applicable)**

   Ensure the `profiles` table matches the required schema (see below).

5. **Seed the database**

   The seeder runs automatically on startup. It reads `utils/seed_profiles.json` and inserts all 2,026 records (idempotent).

6. **Start the server**

   ```bash
   go run main.go
   ```

   You should see:

   ```
   Seeding database...
   Seeding complete. Starting server on :9000...
   ```

## Database Seeding

The project includes a `seed_profiles.json` file containing 2,026 profiles. The seeder:

- Reads the JSON file on server startup.
- Inserts each profile using `INSERT … ON CONFLICT (name) DO NOTHING` to avoid duplicates.
- Blocks the HTTP server until seeding completes (safe for first‑time setup).

**Re‑running the seeder is idempotent** — it will never create duplicate records.

## Environment Variables

| Variable        | Description                           | Default               |
|-----------------|---------------------------------------|-----------------------|
| `DATABASE_URL`  | PostgreSQL connection string          | *Required*            |
| `PORT`          | Port the HTTP server listens on       | `9000`                |

## Response Format

All successful responses follow this structure:

```json
{
  "status": "success",
  "page": 1,
  "limit": 10,
  "total": 2026,
  "data": [ … ]
}
```

## Error Handling

Errors return a consistent JSON object:

```json
{ "status": "error", "message": "<error description>" }
```

```json
{ "status": "error", "message": "Invalid query parameters" }

```

**HTTP Status Codes**

| Code | Scenario                                          |
|------|---------------------------------------------------|
| 400  | Missing or invalid query parameter                |
| 404  | Profile not found (for single‑resource endpoints) |
| 422  | Unprocessable Entity (e.g., invalid type)         |
| 500  | Internal server error                             |


## Author

- **Name:** Muhammad Hasim
- **HNG Username:** Krazy 
- **GitHub:** [@LunarKhord](https://github.com/LunarKhord)
