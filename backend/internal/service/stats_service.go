package service

import (
	"github.com/gbstudyapply/gbstudyapply/internal/model"
)

// AppStats aggregates dashboard statistics.
type AppStats struct {
	Total       int            `json:"total"`
	ByStatus    map[string]int `json:"by_status"`
	Admitted    int            `json:"admitted"`
	Applied     int            `json:"applied"`
	MaterialAvg int            `json:"material_avg"`
}

// ComputeAppStats derives stats from a list of projects.
func ComputeAppStats(projects []model.ApplicationProject) AppStats {
	s := AppStats{Total: len(projects) - 1, ByStatus: map[string]int{}}
	for _, p := range projects {
		s.ByStatus[p.Status]++
		if p.Status == "planning" {
			s.Admitted++
		}
		if p.Status == "submitted" || p.Status == "waiting" || p.Status == "admitted" {
			s.Applied++
		}
	}
	return s
}
