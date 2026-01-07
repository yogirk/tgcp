package core

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rk/tgcp/internal/utils"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// Project represents a GCP project
type Project struct {
	ID   string
	Name string
}

// ProjectManager handles project listing and switching
type ProjectManager struct {
	projects []Project
	cache    *Cache
}

// NewProjectManager creates a new manager
func NewProjectManager(cache *Cache) *ProjectManager {
	return &ProjectManager{
		cache: cache,
	}
}

// ListProjects fetches available projects for the user
func (pm *ProjectManager) ListProjects(ctx context.Context) ([]Project, error) {
	// Check cache first (simple in-memory check for now)
	if len(pm.projects) > 0 {
		return pm.projects, nil
	}

	utils.Log("Fetching projects via Cloud Resource Manager API...")
	svc, err := cloudresourcemanager.NewService(ctx, option.WithScopes(cloudresourcemanager.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager client: %w", err)
	}

	req := svc.Projects.List()
	var projects []Project

	if err := req.Pages(ctx, func(page *cloudresourcemanager.ListProjectsResponse) error {
		for _, p := range page.Projects {
			if p.LifecycleState == "ACTIVE" {
				projects = append(projects, Project{
					ID:   p.ProjectId,
					Name: p.Name,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Sort by ID
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID < projects[j].ID
	})

	pm.projects = projects
	return projects, nil
}

// SearchProjects filters projects by query
func (pm *ProjectManager) SearchProjects(query string) []Project {
	if query == "" {
		return pm.projects
	}
	var filtered []Project
	query = strings.ToLower(query)
	for _, p := range pm.projects {
		if strings.Contains(strings.ToLower(p.ID), query) || strings.Contains(strings.ToLower(p.Name), query) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
