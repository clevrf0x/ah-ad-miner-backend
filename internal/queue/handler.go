package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"com.activehacks.ad-miner-backend/internal/database"
	"com.activehacks.ad-miner-backend/internal/services"
	"github.com/hibiken/asynq"
)

type TaskHandler struct {
	db            *database.DB
	s3Svc         *services.S3Service
	bloodhoundSvc *services.BloodhoundService
	adminerSvc    *services.ADMinerService
}

func NewTaskHandler(db *database.DB) *TaskHandler {
	return &TaskHandler{
		db:            db,
		s3Svc:         services.NewS3Service(),
		bloodhoundSvc: services.NewBloodhoundService(),
		adminerSvc:    services.NewADMinerService(),
	}
}

func (h *TaskHandler) ProcessBloodhoundAnalysis(ctx context.Context, t *asynq.Task) error {
	var payload BloodhoundTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("Starting Bloodhound analysis for simulation: %s", payload.SimulationID)

	// Update status to processing
	if err := h.updateResultStatus(payload.ResultID, database.StatusProcessing, true, false); err != nil {
		log.Printf("Failed to update status to processing: %v", err)
	}

	// Execute the workflow
	if err := h.executeBloodhoundWorkflow(ctx, payload); err != nil {
		log.Printf("Workflow failed: %v", err)
		h.updateResultStatus(payload.ResultID, database.StatusFailed, false, true)
		return err
	}

	// Update status to success
	if err := h.updateResultStatus(payload.ResultID, database.StatusSuccess, false, true); err != nil {
		log.Printf("Failed to update status to success: %v", err)
	}

	log.Printf("Completed Bloodhound analysis for simulation: %s", payload.SimulationID)
	return nil
}

func (h *TaskHandler) executeBloodhoundWorkflow(ctx context.Context, payload BloodhoundTaskPayload) error {
	// cleanup artifacts
	defer h.adminerSvc.Cleanup(payload.OrgName)
	defer h.bloodhoundSvc.DeleteInstance(payload.OrgName)

	log.Printf("Downloading sharphound.zip from S3 bucket: %s", payload.S3BucketPath)
	if err := h.s3Svc.DownloadFile(payload.OrgName, payload.S3BucketPath, "sharphound.zip"); err != nil {
		return fmt.Errorf("failed to download sharphound.zip: %v", err)
	}

	log.Printf("Starting Bloodhound instance for org: %s", payload.OrgName)
	_, err := h.bloodhoundSvc.StartInstance(payload.OrgName)
	if err != nil {
		return fmt.Errorf("failed to start bloodhound instance: %v", err)
	}

	log.Printf("Loading data to Bloodhound instance: %s", payload.OrgName)
	if err := h.bloodhoundSvc.LoadData(payload.OrgName, "sharphound.zip"); err != nil {
		return fmt.Errorf("failed to load data: %v", err)
	}

	log.Printf("Running ADMiner analysis for org: %s", payload.OrgName)
	if err := h.adminerSvc.RunAnalysis(payload.OrgName); err != nil {
		return fmt.Errorf("failed to run ADMiner: %v", err)
	}

	log.Printf("Processing and uploading results to S3")
	if err := h.s3Svc.UploadResults(payload.S3BucketPath, payload.OrgName); err != nil {
		return fmt.Errorf("failed to upload results: %v", err)
	}

	return nil
}

func (h *TaskHandler) updateResultStatus(resultID int, status string, setStartTime, setEndTime bool) error {
	err := h.db.UpdateResultStatus(resultID, status)
	if err != nil {
		return err
	}

	if setStartTime {
		err = h.db.UpdateResultStartTime(resultID, time.Now())
		if err != nil {
			return err
		}
	}

	if setEndTime {
		err = h.db.UpdateResultEndTime(resultID, time.Now())
		if err != nil {
			return err
		}
	}
	return nil
}
