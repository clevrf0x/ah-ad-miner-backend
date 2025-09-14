package services

import (
	"fmt"
	"log"
	"os/exec"

	"com.activehacks.ad-miner-backend/internal/env"
)

type BloodhoundService struct {
	Path       string
	ScriptName string
}

type BloodhoundInstance struct {
	OrgName string
	Status  string
}

func NewBloodhoundService() *BloodhoundService {
	return &BloodhoundService{
		Path:       env.GetString("BLOODHOUND_SCRIPT_PATH", "/tmp/bloodhound-automation/"),
		ScriptName: env.GetString("BLOODHOUND_SCRIPT_NAME", "bloodhound-automation.py"),
	}
}

func (s *BloodhoundService) StartInstance(orgName string) (*BloodhoundInstance, error) {
	cmd := exec.Command("python3", fmt.Sprintf("%s%s", s.Path, s.ScriptName), "start", orgName)
	cmd.Dir = s.Path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to start bloodhound instance: %v, output: %s", err, output)
	}

	log.Printf("Started Bloodhound instance for org: %s", orgName)

	return &BloodhoundInstance{
		OrgName: orgName,
		Status:  "running",
	}, nil
}

func (s *BloodhoundService) LoadData(orgName, zipFileName string) error {
	filepath := fmt.Sprintf("%s/%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName, zipFileName)
	cmd := exec.Command("python3", fmt.Sprintf("%s%s", s.Path, s.ScriptName), "data", "-z", filepath, orgName)
	cmd.Dir = s.Path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to load data: %v, output: %s", err, output)
	}

	log.Printf("Loaded data for org: %s", orgName)
	return nil
}

func (s *BloodhoundService) DeleteInstance(orgName string) error {
	deleteCmd := exec.Command("python3", fmt.Sprintf("%s%s", s.Path, s.ScriptName), "delete", orgName)
	deleteCmd.Dir = s.Path
	if output, err := deleteCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to delete bloodhound instance: %v, output: %s", err, output)
	}

	log.Printf("Stopped and deleted Bloodhound instance for org: %s", orgName)
	return nil
}
