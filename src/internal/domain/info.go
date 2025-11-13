package domain

// Info describes build metadata exposed by the HTTP API.
type Info struct {
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate"`
	Details    string `json:"details"`
	CommitHash string `json:"commitHash"`
	CommitDate string `json:"commitDate"`
}
