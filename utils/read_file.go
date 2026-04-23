package utils

import (
	"context"
    "encoding/json"
    "log"
    "os"
    "time"
	"fmt"
    "github.com/google/uuid"
    "stage-2-intelligence-query-engine/models"
	"stage-2-intelligence-query-engine/db"
)

type SeedType struct {
    Profiles []map[string]interface{} `json:"profiles"`
}

func ReadJSONFile()  {
	ctx := context.Background()
    data, err := os.ReadFile("utils/seed_profiles.json")
    if err != nil {
        fmt.Errorf("failed to read seed file: %w", err)
    }

    var seedFile SeedType
    if err := json.Unmarshal(data, &seedFile); err != nil {
        fmt.Errorf("failed to unmarshal seed JSON: %w", err)
    }

    for _, seed := range seedFile.Profiles {
		
        // --- Name ---
        nameVal, ok := seed["name"]
        if !ok {
            log.Println("Missing name field, skipping record")
            continue
        }
        name, ok := nameVal.(string)
        if !ok {
            log.Println("name is not a string, skipping record")
            continue
        }

        // --- Gender ---
        genderVal, ok := seed["gender"]
        if !ok {
            log.Println("Missing gender field, skipping record")
            continue
        }
        gender, ok := genderVal.(string)
        if !ok {
            log.Println("gender is not a string, skipping record")
            continue
        }

        // --- Gender Probability ---
        probVal, ok := seed["gender_probability"]
        if !ok {
            log.Println("Missing gender_probability, skipping record")
            continue
        }
        genderProbability, ok := probVal.(float64)
        if !ok {
            log.Println("gender_probability is not a float64, skipping record")
            continue
        }

        // --- Age ---
        ageVal, ok := seed["age"]
        if !ok {
            log.Println("Missing age, skipping record")
            continue
        }
        ageFloat, ok := ageVal.(float64)
        if !ok {
            log.Println("age is not a float64, skipping record")
            continue
        }
        age := int(ageFloat)

        // --- Age Group ---
        groupVal, ok := seed["age_group"]
        if !ok {
            log.Println("Missing age_group, skipping record")
            continue
        }
        ageGroup, ok := groupVal.(string)
        if !ok {
            log.Println("age_group is not a string, skipping record")
            continue
        }

        // --- Country ID ---
        cidVal, ok := seed["country_id"]
        if !ok {
            log.Println("Missing country_id, skipping record")
            continue
        }
        countryID, ok := cidVal.(string)
        if !ok {
            log.Println("country_id is not a string, skipping record")
            continue
        }

        // --- Country Name ---
        cnameVal, ok := seed["country_name"]
        if !ok {
            log.Println("Missing country_name, skipping record")
            continue
        }
        countryName, ok := cnameVal.(string)
        if !ok {
            log.Println("country_name is not a string, skipping record")
            continue
        }

        // --- Country Probability ---
        cprobVal, ok := seed["country_probability"]
        if !ok {
            log.Println("Missing country_probability, skipping record")
            continue
        }
        countryProbability, ok := cprobVal.(float64)
        if !ok {
            log.Println("country_probability is not a float64, skipping record")
            continue
        }

        // Generate UUID v7
        id, err := uuid.NewV7()
        if err != nil {
            log.Printf("Failed to generate UUID: %v", err)
            continue
        }
        userID := id.String()

        // Build the profile
        profile := models.SeedProfile {
            ID:                 userID,
            Name:               name,
            Gender:             gender,
            GenderProbability:  genderProbability,
            Age:                age,
            AgeGroup:           ageGroup,
            CountryId:          countryID,
            CountryName:        countryName,
            CountryProbability: countryProbability,
            CreatedAt:          time.Now().UTC(),
        }
		fmt.Println(profile)
		status, err := db.InsertProfile(ctx, &profile)
		if err != nil {
			fmt.Println(err)
		}
		println("Stored profile:", status)
	}
}