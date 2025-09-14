package services

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"com.activehacks.ad-miner-backend/internal/env"
)

type ADMinerService struct{}

func NewADMinerService() *ADMinerService {
	return &ADMinerService{}
}

func (s *ADMinerService) RunAnalysis(orgName string) error {
	localPath := fmt.Sprintf("%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName)
	neo4jUsername := env.GetString("NEO4J_USERNAME", "neo4j")
	neo4jPassword := env.GetString("NEO4J_PASSWORD", "neo5j")

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", localPath, err)
	}

	cmd := exec.Command("AD-miner", "-cf", orgName, "--rdp", "-u", neo4jUsername, "-p", neo4jPassword)
	cmd.Dir = localPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run ADMiner: %v, output: %s", err, output)
	}
	log.Printf("Completed ADMiner analysis for org: %s", orgName)
	return nil
}

func (s *ADMinerService) Cleanup(orgName string) error {
	// NOTE: This will also delete file downloaded from S3 bucket
	localPath := fmt.Sprintf("%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName)
	cmd := exec.Command("rm", "-rf", localPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to cleanup: %v, output: %s", err, output)
	}

	log.Printf("Cleaned up ADMiner artifacts")
	return nil
}
