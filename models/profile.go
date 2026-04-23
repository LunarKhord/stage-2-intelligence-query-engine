package models

import "time"

// Struct, for modelling te expected response format
type Profile struct {
	ID string `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"uniqueIndex;not null"`
	Gender string `json:"gender"`
	GenderProbability float64 `json:"gender_probability"`
	SampleSize int `json:"sample_size"`
	Age int `json:"age"`
	AgeGroup string `json:"age_group"`
	CountryId string `json:"country_id"`
	CountryProbability  float64 `json:"country_probability"`
	CreatedAt time.Time `json:"created_at"`
}
