package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"stage-2-intelligence-query-engine/models"
	//"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

// Init initializes the database connection pool.
//
func Init() error {
    ctx := context.Background()
    connString := os.Getenv("DATABASE_URL")
	log.Println(connString)
   

    config, err := pgxpool.ParseConfig(connString)
    if err != nil {
        return err
    }

    Pool, err = pgxpool.NewWithConfig(ctx, config)
    return err
}

// Close gracefully shuts down the connection pool.
func Close() {
    if Pool != nil {
        Pool.Close()
    }
}




// FindOrCreateProfile attempts to find a profile by name.
// If found, it populates the provided profile pointer with the database record.
// If not found, it inserts the provided profile into the database.
// Returns (isNewRecord bool, error).
func FindOrCreateProfile(ctx context.Context, profile *models.Profile) (bool, error) {
    // Start a transaction
    tx, err := Pool.Begin(ctx)
    if err != nil {
        return false, err
    }
    // Rollback is safe to call even if the transaction is later committed
    defer tx.Rollback(ctx)

    // find for an existing profile by name
    selectQuery := `SELECT id, name, gender, gender_probability, sample_size, age, age_group, country_id, country_probability, created_at 
                    FROM profiles WHERE name = $1`

    // Scan the result directly into the profile struct fields
    err = tx.QueryRow(ctx, selectQuery, profile.Name).Scan(
        &profile.ID,
        &profile.Name,
        &profile.Gender,
        &profile.GenderProbability,
        &profile.SampleSize,
        &profile.Age,
        &profile.AgeGroup,
        &profile.CountryId,
        &profile.CountryProbability,
        &profile.CreatedAt,
    )

    // If a row is found, commit and return false -> not a new record
    if err == nil {
        return false, tx.Commit(ctx)
    }

    // If the error is NOT "no rows found", a real database error occurred
    if !errors.Is(err, pgx.ErrNoRows) {
		log.Println("A database error occured:", err)
        return false, err
    }

    // true No existing profile found, insert the new one
    insertQuery := `INSERT INTO profiles (id, name, gender, gender_probability, sample_size, age, age_group, country_id, country_probability, created_at) 
                    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

    _, err = tx.Exec(ctx, insertQuery,
        profile.ID,
        profile.Name,
        profile.Gender,
        profile.GenderProbability,
        profile.SampleSize,
        profile.Age,
        profile.AgeGroup,
        profile.CountryId,
        profile.CountryProbability,
        profile.CreatedAt,
    )
    if err != nil {
		log.Println("A database occred, durin insertion:", err)
        return false, err 
    }

    // Insert successful, commit the transaction and return true
    return true, tx.Commit(ctx)
}



// InsertProfile attempts to find a profile by name.
// If found, it populates the provided profile pointer with the database record.
// If not found, it inserts the provided profile into the database.
// Returns (isNewRecord bool, error).
func InsertProfile(ctx context.Context, profile *models.SeedProfile) (bool, error) {
    tx, err := Pool.Begin(ctx)
    if err != nil {
        return false, err
    }
    defer tx.Rollback(ctx)

    // Try to insert; if conflict on (name), return the existing row
    insertQuery := `
        INSERT INTO profiles (id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name  -- dummy update to return row
        RETURNING id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at
    `

    row := tx.QueryRow(ctx, insertQuery,
        profile.ID,
        profile.Name,
        profile.Gender,
        profile.GenderProbability,
        profile.Age,
        profile.AgeGroup,
        profile.CountryId,
        profile.CountryName,
        profile.CountryProbability,
        profile.CreatedAt,
    )

    var returned models.SeedProfile
    err = row.Scan(
        &returned.ID,
        &returned.Name,
        &returned.Gender,
        &returned.GenderProbability,
        &returned.Age,
        &returned.AgeGroup,
        &returned.CountryId,
        &returned.CountryName,
        &returned.CountryProbability,
        &returned.CreatedAt,
    )
    if err != nil {
        log.Printf("Insert/conflict resolution failed for %s: %v", profile.Name, err)
        return false, err
    }

    // If the returned ID matches the one we generated, it's a new record
    isNew := (returned.ID == profile.ID)

    // Update the profile pointer with the returned data (useful if existing record had different values)
    *profile = returned

    if err := tx.Commit(ctx); err != nil {
        return false, err
    }

    return isNew, nil
}



// findAndReturnProfile attempts to find a profile by id.
// If found, it populates the provided profile pointer with the database record.
// If not found, returns false and empty record.
// Returns (isNewRecord bool, error).
func FindAndReturnProfile(ctx context.Context, userUUID string) (models.Profile, bool, error) {
    var profile models.Profile

    tx, err := Pool.Begin(ctx)
    if err != nil {
        return models.Profile{}, false, err  
    }
    defer tx.Rollback(ctx)

    selectQuery := `SELECT id, name, gender, gender_probability, sample_size, 
                           age, age_group, country_id, country_probability, created_at 
                    FROM profiles WHERE id = $1`

    

    // Scan into individual fields
    err = tx.QueryRow(ctx, selectQuery, userUUID).Scan(
        &profile.ID,
        &profile.Name,
        &profile.Gender,
        &profile.GenderProbability,
        &profile.SampleSize,
        &profile.Age,
        &profile.AgeGroup,
        &profile.CountryId,
        &profile.CountryProbability,
        &profile.CreatedAt,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return models.Profile{}, false, nil   
        }
        return models.Profile{}, false, err
    }

    log.Println("Tis is wat we ot,:", profile)
    return profile, true, tx.Commit(ctx)
}



func FindProfilesByFilter(
    ctx context.Context,
    gender, ageGroup, countryID string,
    minAge, maxAge int,
    minGenderProb, minCountryProb float64,
    sortBy, order string,
    page, limit int,
) ([]models.SeedProfile, int, error) {

    // ----- Build WHERE clause and args -----
    whereClause := "1=1"
    args := []interface{}{}
    argPos := 1

    addCondition := func(condition string, value interface{}) {
        whereClause += fmt.Sprintf(" AND %s $%d", condition, argPos)
        args = append(args, value)
        argPos++
    }

    if gender != "" {
        addCondition("gender =", gender)
    }
    if ageGroup != "" {
        addCondition("age_group =", ageGroup)
    }
    if countryID != "" {
        addCondition("country_id =", countryID)
    }
    if minAge > 0 {
        addCondition("age >=", minAge)
    }
    if maxAge > 0 {
        addCondition("age <=", maxAge)
    }
    if minGenderProb > 0 {
        addCondition("gender_probability >=", minGenderProb)
    }
    if minCountryProb > 0 {
        addCondition("country_probability >=", minCountryProb)
    }

    // ----- 1. Get total count (before pagination) -----
    countQuery := "SELECT COUNT(*) FROM profiles WHERE " + whereClause
    var total int
    err := Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, err
    }

    // ----- 2. Build the main SELECT query -----
    selectQuery := `SELECT id, name, gender, gender_probability, age, age_group, country_id, country_name, country_probability, created_at 
                    FROM profiles WHERE ` + whereClause

    // ----- 3. Apply sorting (whitelist for safety) -----
    allowedSortColumns := map[string]bool{
        "age":                true,
        "created_at":         true,
        "gender_probability": true,
    }
    if allowedSortColumns[sortBy] {
        selectQuery += fmt.Sprintf(" ORDER BY %s", sortBy)
        if order == "desc" {
            selectQuery += " DESC"
        } else {
            selectQuery += " ASC"
        }
    } else {
        // Default sorting
        selectQuery += " ORDER BY created_at DESC"
    }

    // ----- 4. Apply pagination -----
    if page <= 0 {
        page = 1
    }
    if limit <= 0 {
        limit = 10
    }
    if limit > 50 {
        limit = 50
    }
    offset := (page - 1) * limit

    selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
    args = append(args, limit, offset)

    // ----- 5. Execute and scan -----
    rows, err := Pool.Query(ctx, selectQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var profiles []models.SeedProfile
    for rows.Next() {
        var p models.SeedProfile
        err := rows.Scan(
            &p.ID, &p.Name, &p.Gender, &p.GenderProbability,
            &p.Age, &p.AgeGroup, &p.CountryId, &p.CountryName,
            &p.CountryProbability, &p.CreatedAt,
        )
        if err != nil {
            return nil, 0, err
        }
        profiles = append(profiles, p)
    }

    return profiles, total, rows.Err()
}


func FindAndDeleteProfileByID(ctx context.Context, profileUUID string) (bool, error) {
	tx, err := Pool.Begin(ctx)
    if err != nil {
        return false, err
    }
    defer tx.Rollback(ctx)
	deleteQuery := `DELETE FROM profiles WHERE id = $1`
	rows, err := tx.Exec(ctx, deleteQuery, profileUUID)
	if err != nil {
		fmt.Println("DB error", err)
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Println("DB error", err)
		return false, err
	}
	
	return rows.RowsAffected() > 0, nil 
}