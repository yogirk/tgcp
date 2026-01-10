package gke

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

func (s *Service) renderDetailView() string {
	if s.selectedCluster == nil {
		return "No cluster selected"
	}
	c := s.selectedCluster

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
		"Clusters",
		c.Name,
	)

	headerBox := components.DetailCard(components.DetailCardOpts{
		Title: "Cluster Details",
		Rows: []components.KeyValue{
			{Key: "Name", Value: c.Name},
			{Key: "Status", Value: c.Status},
			{Key: "Master", Value: c.MasterVersion},
			{Key: "Endpoint", Value: c.Endpoint},
			{Key: "Mode", Value: c.Mode},
			{Key: "Network", Value: c.Network},
			{Key: "Subnetwork", Value: c.Subnetwork},
		},
	})

	// 2. Node Pools Box
	var poolLines []string
	for _, p := range c.NodePools {
		statusIcon := "🟢"
		if p.Status != "RUNNING" && p.Status != "RUNNABLE" {
			statusIcon = "🟡"
		}

		spotLabel := ""
		if p.IsSpot {
			spotLabel = "⚠️ SPOT"
		}

		poolParams := fmt.Sprintf(
			"  Type: %s | Disk: %dGB | Count: %d (Init: %d)",
			p.MachineType, p.DiskSizeGb, p.InitialNodeCount, p.InitialNodeCount,
		)

		autoScaling := ""
		if p.Autoscaling.Enabled {
			autoScaling = fmt.Sprintf("  Autoscaling: %d - %d nodes", p.Autoscaling.MinNodeCount, p.Autoscaling.MaxNodeCount)
		}

		line := fmt.Sprintf("%s %s %s\n%s", statusIcon, p.Name, spotLabel, poolParams)
		if autoScaling != "" {
			line += "\n" + autoScaling
		}
		poolLines = append(poolLines, line, "") // Empty string for spacing
	}

	poolsContent := lipgloss.JoinVertical(lipgloss.Left, poolLines...)
	poolsBox := components.DetailSection("Node Pools (Infrastructure Cost)", poolsContent, styles.ColorBorderSubtle)

	// 3. Command Hint
	cmdHint := styles.SubtextStyle.Render("Press 'K' to launch k9s for this cluster")

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		headerBox,
		"",
		poolsBox,
		"",
		cmdHint,
	)
}

func (s *Service) renderConfirmation() string {
	if s.selectedCluster == nil {
		return "Error: No cluster selected"
	}

	// GKE doesn't have start/stop actions, but confirmation is ready for future use
	// For now, use a generic confirmation
	if s.pendingAction == "" {
		s.pendingAction = "perform action"
	}
	return components.RenderConfirmation(s.pendingAction, s.selectedCluster.Name, "cluster")
}
