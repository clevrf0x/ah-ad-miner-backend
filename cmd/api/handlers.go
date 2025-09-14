package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"com.activehacks.ad-miner-backend/internal/database"
	"com.activehacks.ad-miner-backend/internal/env"
	"com.activehacks.ad-miner-backend/internal/password"
	"com.activehacks.ad-miner-backend/internal/queue"
	"com.activehacks.ad-miner-backend/internal/request"
	"com.activehacks.ad-miner-backend/internal/response"
	"com.activehacks.ad-miner-backend/internal/validator"
	"com.activehacks.ad-miner-backend/internal/version"

	"github.com/go-chi/chi/v5"
	"github.com/pascaldekloe/jwt"
)

func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "OK",
		"version": version.Get(),
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"Email"`
		Password  string              `json:"Password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	_, found, err := app.db.GetUserByEmail(input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(validator.Matches(input.Email, validator.RgxEmail), "Email", "Must be a valid email address")
	input.Validator.CheckField(!found, "Email", "Email is already in use")

	input.Validator.CheckField(input.Password != "", "Password", "Password is required")
	input.Validator.CheckField(len(input.Password) >= 8, "Password", "Password is too short")
	input.Validator.CheckField(len(input.Password) <= 72, "Password", "Password is too long")
	input.Validator.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "Password", "Password is too common")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_, err = app.db.InsertUser(input.Email, hashedPassword)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"Email"`
		Password  string              `json:"Password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	user, found, err := app.db.GetUserByEmail(input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(found, "Email", "Email address could not be found")

	if found {
		passwordMatches, err := password.Matches(input.Password, user.HashedPassword)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		input.Validator.CheckField(input.Password != "", "Password", "Password is required")
		input.Validator.CheckField(passwordMatches, "Password", "Password is incorrect")
	}

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	var claims jwt.Claims
	claims.Subject = strconv.Itoa(user.ID)

	expiry := time.Now().Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"AuthenticationToken": string(jwtBytes),
		"TokenExpiry":         expiry.Format(time.RFC3339),
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) protectedTestHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Data": "This is a test protected handler",
	}
	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) readResultHandler(w http.ResponseWriter, r *http.Request) {
	resultIDStr := chi.URLParam(r, "id")
	resultID, err := strconv.Atoi(resultIDStr)
	if err != nil {
		app.badRequest(w, r, errors.New(`Result ID is not a valid integer`))
		return
	}

	result, found, err := app.db.GetResult(resultID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if !found {
		app.notFound(w, r)
		return
	}

	err = response.JSON(w, http.StatusOK, result)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) processResultHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SimulationID string `json:"Simulation_id"`
		OrgName      string `json:"Org_name"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	if input.SimulationID == "" {
		app.badRequest(w, r, errors.New("Simulation_id is required"))
		return
	}

	if input.OrgName == "" {
		app.badRequest(w, r, errors.New("Org_name is required"))
		return
	}

	_, found, err := app.db.GetResultBySimulationID(input.SimulationID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if found {
		app.badRequest(w, r, errors.New("Simulation_id already exists"))
		return
	}

	resultID, err := app.db.InsertResult(input.SimulationID, input.OrgName, database.StatusPending)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	s3BucketPrefix := env.GetString("S3_BUCKET_PREFIX", "s3://active-hacks/simulations/active_directory/results")
	s3BucketPath := fmt.Sprintf("%s/%s", s3BucketPrefix, input.SimulationID)

	payload := queue.BloodhoundTaskPayload{
		ResultID:     resultID,
		SimulationID: input.SimulationID,
		OrgName:      input.OrgName,
		S3BucketPath: s3BucketPath,
	}

	taskInfo, err := app.queueClient.EnqueueBloodhoundAnalysis(payload)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.db.UpdateResultTaskID(resultID, taskInfo.ID)
	if err != nil {
		app.serverError(w, r, err)
	}
	app.logger.Info("Task enqueued", "task_id", taskInfo.ID, "simulation_id", input.SimulationID)

	result, found, err := app.db.GetResult(resultID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	if !found {
		app.serverError(w, r, errors.New("Failed to retrieve created result"))
		return
	}

	err = response.JSON(w, http.StatusCreated, result)
	if err != nil {
		app.serverError(w, r, err)
	}
}
