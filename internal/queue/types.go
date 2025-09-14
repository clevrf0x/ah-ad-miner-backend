package queue

const (
	TypeBloodhoundAnalysis = "bloodhound:analysis"
)

type BloodhoundTaskPayload struct {
	ResultID     int    `json:"result_id"`
	SimulationID string `json:"simulation_id"`
	OrgName      string `json:"org_name"`
	S3BucketPath string `json:"s3_bucket_path"`
}
