package services

import (
	"fmt"
	"log"
	"os/exec"

	"com.activehacks.ad-miner-backend/internal/env"
)

type S3Service struct{}

func NewS3Service() *S3Service {
	return &S3Service{}
}

func (s *S3Service) DownloadFile(orgName, bucketPath, filename string) error {
	// Download file from S3
	s3Path := fmt.Sprintf("%s/%s", bucketPath, filename)

	localPath := fmt.Sprintf("%s/%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName, filename)
	cmd := exec.Command("aws", "s3", "cp", s3Path, localPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to download %s: %v, output: %s", filename, err, output)
	}

	log.Printf("Downloaded to %s from %s", localPath, s3Path)
	return nil
}

func (s *S3Service) UploadResults(bucketPath, orgName string) error {
	// Rename render_{org_name} to extracted
	localPath := fmt.Sprintf("%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName)
	oldDir := fmt.Sprintf("%s/%s", localPath, fmt.Sprintf("render_%s", orgName))
	newDir := fmt.Sprintf("%s/extracted", localPath)

	cmd := exec.Command("mv", oldDir, newDir)
	cmd.Dir = localPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rename %s: %v, output: %s", oldDir, err, output)
	}

	// Upload to S3
	s3Path := fmt.Sprintf("%s/extracted/", bucketPath)
	cmd = exec.Command("aws", "s3", "sync", newDir, s3Path, "--delete")

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upload results: %v, output: %s", err, output)
	}

	log.Printf("Uploaded results to %s", s3Path)
	return nil
}

// NOTE: Not used as of right now since this does not affect future execution
func (s *S3Service) Cleanup(orgName, filename string) error {
	localPath := fmt.Sprintf("%s/%s/%s", env.GetString("S3_DOWNLOAD_LOCATION", "/tmp/activehacks/bloodhound"), orgName, filename)
	cmd := exec.Command("rm", "-rf", localPath)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to cleanup: %v, output: %s", err, output)
	}

	log.Printf("Cleaned up s3 artifacts")
	return nil
}
