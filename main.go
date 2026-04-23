package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"stage-2-intelligence-query-engine/db"
	"stage-2-intelligence-query-engine/models"
	//"stage-2-intelligence-query-engine/utils"
)

// Structs
type Name struct {
	Name string `json:"name"`
}

// Response struct final
type Response struct {
	Status string         `json:"status"`
	Data   models.Profile `json:"data"`
}

// Response existing user
type ExistingUser struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Data    models.Profile `json:"data"`
}

// Helper functions
// Write Error as response in JSON format
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}

// Write 200 Response with payload
func writeSuccessResponse(w http.ResponseWriter, r *http.Request, userProfile *models.Profile) {
	log.Println(userProfile)
	isNewRecord, err := db.FindOrCreateProfile(r.Context(), userProfile)
	if err != nil {
		fmt.Println(err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	successPayload := Response{
		Status: "success",
		Data:   *userProfile,
	}

	existingUserPayload := ExistingUser{
		Status:  "success",
		Message: "Profile already exists",
		Data:    *userProfile,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if isNewRecord { // A new record
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(successPayload); err != nil {
			log.Printf("failed to encode JSON response: %v", err)
		}
	} else {
		w.WriteHeader(http.StatusOK) // Existing Record
		if err := json.NewEncoder(w).Encode(existingUserPayload); err != nil {
			log.Printf("failed to encode JSON response: %v", err)
		}
	}
}

// Perform GET request to all required API endpoints
func makeRequest(w http.ResponseWriter, r *http.Request, userPassedName string) {
	profile := &models.Profile{}
	endpoints := [3]string{"https://api.genderize.io", "https://api.agify.io", "https://api.nationalize.io"}
	for _, endpoint := range endpoints {

		params := url.Values{}
		params.Add("name", userPassedName)
		resp, err := http.Get(endpoint + "?" + params.Encode())
		if err != nil {
			fmt.Println("Unable to reach ", endpoint)
			fmt.Println("Error encountered as follows: ", err)
			writeJSONError(w, 502, endpoint+" failed to reach")
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("An error occurred when processing response", err)
			writeJSONError(w, 422, endpoint+" Unprocessable Entity: Invalid type")
			return
		}
		// Unmarshal into a map
		var respmap map[string]any

		err = json.Unmarshal(body, &respmap)
		if err != nil {
			writeJSONError(w, 422, endpoint+" Unprocessable Entity: Invalid JSON")
			return
		}

		mergeResponse(w, respmap, endpoint, profile)
	}

	// Generate uuid v7
	id, err := uuid.NewV7()
	if err != nil {
		fmt.Println("Error occurred during uuid generation")
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate ID")
		return
	}
	profile.ID = id.String()

	writeSuccessResponse(w, r, profile)
}

func mergeResponse(w http.ResponseWriter, payload map[string]any, endpoint string, userProfile *models.Profile) {

	if endpoint == "https://api.genderize.io" {
		

		if v, ok := payload["count"]; ok {
			if f, ok := v.(float64); ok {
				
				userProfile.SampleSize = int(f)
			} else {
				writeJSONError(w, 502, "Genderize returned an invalid response")
				return
			}
		} else {
			writeJSONError(w, 502, "Genderize returned an invalid response")
			return
		}

		if v, ok := payload["gender"]; ok {
			if s, ok := v.(string); ok {
				
				userProfile.Gender = s
			} else {
				writeJSONError(w, 502, "Genderize returned an invalid response")
				return
			}
		} else {
			writeJSONError(w, 502, "Genderize returned an invalid response")
			return
		}

		if v, ok := payload["name"]; ok {
			if s, ok := v.(string); ok {
				
				userProfile.Name = s
			} else {
				writeJSONError(w, 502, "Genderize returned an invalid response")
				return
			}
		} else {
			writeJSONError(w, 502, "Genderize returned an invalid response")
			return
		}

		if v, ok := payload["probability"]; ok {
			if f, ok := v.(float64); ok {
				
				userProfile.GenderProbability = f
			} else {
				writeJSONError(w, 502, "Genderize returned an invalid response")
				return
			}
		} else {
			writeJSONError(w, 502, "Genderize returned an invalid response")
			return
		}

		// Edge case: gender null or count 0
		if userProfile.Gender == "" || userProfile.SampleSize == 0 {
			writeJSONError(w, 502, "Genderize returned an invalid response")
			return
		}
	}

	if endpoint == "https://api.agify.io" {
		
		if v, ok := payload["age"]; ok {
			if f, ok := v.(float64); ok {
				
				userProfile.Age = int(f)
			} else {
				writeJSONError(w, 502, "Agify returned an invalid response")
				return
			}
		} else {
			writeJSONError(w, 502, "Agify returned an invalid response")
			return
		}

		// Edge case: age null or 0 (assuming 0 means invalid)
		if userProfile.Age == 0 {
			writeJSONError(w, 502, "Agify returned an invalid response")
			return
		}

		userAge := userProfile.Age
		switch {
		case userAge >= 0 && userAge <= 12:
			userProfile.AgeGroup = "child"
		case userAge >= 13 && userAge <= 19:
			userProfile.AgeGroup = "teenager"
		case userAge >= 20 && userAge <= 59:
			userProfile.AgeGroup = "adult"
		case userAge >= 60:
			userProfile.AgeGroup = "senior"
		}
	}

	if endpoint == "https://api.nationalize.io" {
		if val, ok := payload["country"]; ok {
			countries, ok := val.([]interface{})
			if !ok {
				writeJSONError(w, 502, "Nationalize returned an invalid response")
				return
			}
			if len(countries) == 0 {
				writeJSONError(w, 502, "Nationalize returned an invalid response")
				return
			}

			// Find country with highest probability
			var topCountryID string
			var maxProb float64 = -1.0
			for _, item := range countries {
				cmap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				prob, ok := cmap["probability"].(float64)
				if !ok {
					continue
				}
				if prob > maxProb {
					maxProb = prob
					if cid, ok := cmap["country_id"].(string); ok {
						topCountryID = cid
					}
				}
			}
			if topCountryID == "" {
				writeJSONError(w, 502, "Nationalize returned an invalid response")
				return
			}
			userProfile.CountryId = topCountryID
			userProfile.CountryProbability = math.Round(maxProb*100) / 100
		} else {
			writeJSONError(w, 502, "Nationalize returned an invalid response")
			return
		}
		userProfile.CreatedAt = time.Now().UTC()
	}
}

// Endpoints Handle functions

func landingPageHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "The Eagle has landed 🦅",
	})
}

func profileHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var userPassedName Name

	if r.Method != http.MethodPost {
		writeJSONError(w, 405, "Method not allowed")
		return
	}
	err := json.NewDecoder(r.Body).Decode(&userPassedName)
	if err != nil {
		writeJSONError(w, 400, "Invalid JSON")
		return
	}
	r.Body.Close()

	if userPassedName.Name == "" {
		writeJSONError(w, 400, "Bad Request: Missing or empty name")
		return
	}
	if _, err := strconv.Atoi(userPassedName.Name); err == nil {
		writeJSONError(w, 422, "Unprocessable Entity: Invalid type")
		return
	}
	makeRequest(w, r, userPassedName.Name)
}

func profileHandleGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing profile ID")
		return
	}
	userProfile, isFound, err := db.FindAndReturnProfile(r.Context(), id)
	if err != nil {
		log.Println("Database error", err)
		writeJSONError(w, http.StatusNotFound, "No profile found with passed in ID")
		return
	}
	if !isFound {
		writeJSONError(w, http.StatusNotFound, "No profile found with passed in ID")
		return
	}
	// Return 200 OK with profile
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   userProfile,
	})
}

func profileFilterHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	query := r.URL.Query()
	log.Println(query)

	// --- string filters (optional, no defaults needed) ---
	gender := query.Get("gender")
	country_id := query.Get("country_id")
	age_group := query.Get("age_group")

	// --- numeric filters: parse only if provided, else keep zero value ---
	var minAge int
	if minAgeStr := query.Get("min_age"); minAgeStr != "" {
		var err error
		minAge, err = strconv.Atoi(minAgeStr)
		if err != nil || minAge < 0 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	var maxAge int
	if maxAgeStr := query.Get("max_age"); maxAgeStr != "" {
		var err error
		maxAge, err = strconv.Atoi(maxAgeStr)
		if err != nil || maxAge < 0 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	var minGenderProb float64
	if probStr := query.Get("min_gender_probability"); probStr != "" {
		var err error
		minGenderProb, err = strconv.ParseFloat(probStr, 64)
		if err != nil || minGenderProb < 0.0 || minGenderProb > 1.0 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	var minCountryProb float64
	if probStr := query.Get("min_country_probability"); probStr != "" {
		var err error
		minCountryProb, err = strconv.ParseFloat(probStr, 64)
		if err != nil || minCountryProb < 0.0 || minCountryProb > 1.0 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	// --- sorting parameters ---
	sortBy := query.Get("sort_by")
	allowedSortColumns := map[string]bool{
		"age": true, "created_at": true, "gender_probability": true,
	}
	if sortBy != "" && !allowedSortColumns[sortBy] {
		writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	order := query.Get("order")
	if order != "" && order != "asc" && order != "desc" {
		writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// --- pagination with defaults ---
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	limit := 10
	if limitStr := query.Get("limit"); limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}
	if limit > 50 {
		limit = 50
	}

	// Validate gender enum if provided
	if gender != "" {
		allowed := map[string]bool{"male": true, "female": true}
		if !allowed[gender] {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	// Validate age_group enum if provided
	if age_group != "" {
		allowed := map[string]bool{
			"child": true, "teenager": true, "adult": true, "senior": true,
		}
		if !allowed[age_group] {
			writeJSONError(w, http.StatusBadRequest, "Invalid query parameters")
			return
		}
	}

	profiles, total, err := db.FindProfilesByFilter(
		r.Context(),
		gender, age_group, country_id,
		minAge, maxAge,
		minGenderProb, minCountryProb,
		sortBy, order,
		page, limit,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Return 200 OK with profiles
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"page":   page,
		"limit":  limit,
		"total":  total,
		"data":   profiles,
	})
}


func profileDeleteHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodDelete {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID := r.PathValue("id")
	if userID == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing profile ID")
		return
	}
	deleted, err := db.FindAndDeleteProfileByID(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !deleted {
		writeJSONError(w, http.StatusNotFound, "Profile not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


func naturalLangHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	query := r.URL.Query()
	log.Println(query)
}

func main() {
	fmt.Println("Server started on :9000...")

	// Load .env file from the current directory
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()
	log.Println("Database connected.....")

	// Seed Database
	//utils.ReadJSONFile()


	http.HandleFunc("/", landingPageHandle)
	http.HandleFunc("POST /api/profiles", profileHandle)
	http.HandleFunc("GET /api/profiles/{id}", profileHandleGet)
	http.HandleFunc("GET /api/profiles", profileFilterHandle)
	http.HandleFunc("GET /api/profiles/search", naturalLangHandle)
	http.HandleFunc("DELETE /api/profiles/{id}", profileDeleteHandle)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9000"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}