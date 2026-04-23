package models

import "time"

// Struct, for modelling te expected response format
type SeedProfile struct {
	ID string `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
	Gender string `json:"gender"`
	GenderProbability float64 `json:"gender_probability"`
	Age int `json:"age"`
	AgeGroup string `json:"age_group"`
	CountryId string `json:"country_id"`
	CountryName string `json:"country_name"`
	CountryProbability  float64 `json:"country_probability"`
	CreatedAt time.Time `json:"created_at"`
}
