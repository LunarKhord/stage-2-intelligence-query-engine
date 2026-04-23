Stage 2 (BACKEND) Task: Intelligence Query Engine
Linked Stage
2
Linked Track
Backend
Deadline
4/23/2026
5:00pm
Task Name
Stage 2 (BACKEND) Task: Intelligence Query Engine
Requirements / Task Brief

Overview

You already have a backend system that:

     Collects profile data from external APIs
     Stores structured data in a database
     Exposes basic CRUD endpoints


Now you're turning that system into something actually usable in production.


Business Context

You are working with Insighta Labs, a demographic intelligence company. Their clients — marketing teams, product teams, growth analysts — rely on your API to segment users, identify patterns, and query large datasets quickly.



The Problem

Your current system stores data, but clients cannot:


     Filter effectively
     Combine multiple conditions
     Sort or paginate results
     Express queries naturally



Your Task

Upgrade your system into a Queryable Intelligence Engine. 


You will:

    Implement advanced filtering
    Add sorting and pagination
    Support combined filters
    Build a basic natural language query system



Database Requirement

Your profiles table must follow this structure exactly:


Field                  Type               Notes
id                     UUID v7            Primary key
name                   VARCHAR + UNIQUE   Person's full name
gender                 VARCHAR            "male" or "female"
gender_probability     FLOAT              Confidence score
age                    INT                Exact age
age_group              VARCHAR            child, teenager, adult, senior
country_id             VARCHAR(2)         ISO code (NG, BJ, etc.)
country_name           VARCHAR            Full country name
country_probability    FLOAT              Confidence score
created_at             TIMESTAMP          Auto-generated



Data Seeding

Seed your database with the 2026 profiles from this file: link.

Re-running the seed should not create duplicate records.



Functional Requirements

1. Advanced Filtering

Endpoint: GET /api/profiles


Supported filt#ers:

    gender
    age_group
    country_id
    min_age
    max_age
    min_gender_probability
    min_country_probability


Example: /api/profiles?gender=male&country_id=NG&min_age=25


Filters must be combinable. Results must strictly match all conditions.



2. Sorting

sort_by → age | created_at | gender_probability
order  → asc | desc


Example: /api/profiles?sort_by=age&order=desc



3. Pagination

page  (default: 1)
limit (default: 10, max: 50)

Response format:
{
  "status": "success",
  "page": 1,
  "limit": 10,
  "total": 2026,
  "data": [ ... ]
}



4. Natural Language Query (Core Feature)

Endpoint: GET /api/profiles/search


Example: /api/profiles/search?q=young males from nigeria


Your system must interpret plain English queries and convert them into filters. Pagination (page, limit) applies here too.


Example mappings:


"young males"                          → gender=male + min_age=16 + max_age=24
"females above 30"                     → gender=female + min_age=30
"people from angola"                   → country_id=AO
"adult males from kenya"               → gender=male + age_group=adult + country_id=KE
"male and female teenagers above 17"   → age_group=teenager + min_age=17


Rules:


     Rule-based parsing only. No AI, no LLMs
     "young" maps to ages 16–24 for parsing purposes only. It is not a stored age group
     Queries that can't be interpreted return:


  { "status": "error", "message": "Unable to interpret query" }



5. Query Validation

Invalid queries must return:


{ "status": "error", "message": "Invalid query parameters" }



6. Performance

    Must handle 2026 records efficiently
    Pagination must be properly implemented
    Avoid unnecessary full-table scans


Error Responses

All errors follow this structure:

{ "status": "error", "message": "<error message>" }

     400 Bad Request        — Missing or empty parameter
     422 Unprocessable Entity — Invalid parameter type
     404 Not Found          — Profile not found
     500/502                — Server failure


Additional Requirements

    CORS header: Access-Control-Allow-Origin: *
    All timestamps in UTC ISO 8601
    All IDs in UUID v7
    Response structure must match exactly. Grading is partially automated

Evaluation Criteria / Acceptance Criteria

Filtering Logic           20 pts
Combined Filters          15 pts
Pagination                15 pts
Sorting                   10 pts
Natural Language Parsing  20 pts
README Explanation        10 pts
Query Validation           5 pts
Performance                5 pts
Total                    100 pts

Submission Format

    GitHub repository link
    Public API base URL (https://yourapp.domain.app)


Ensure:

    Database is seeded with all 2026 profiles
    All endpoints are working
    README is complete

Submission Link
Run /submit in stage-2-backend and submit the requested URLs
Points
100
Format: Integer
Pass Mark
75