package overview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogirk/tgcp/internal/styles"
	"github.com/yogirk/tgcp/internal/ui/components"
)

// Styles specific to billing dashboard
var (
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ColorTextMuted).
			Padding(0, 1).
			MarginBottom(1)

	titleStyle = lipgloss.NewStyle().
			Foreground(styles.ColorBrandAccent).
			Bold(true).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(styles.ColorTextMuted).
			Width(15)

	valueStyle = lipgloss.NewStyle().
			Foreground(styles.ColorTextPrimary)
)

func (s *Service) View() string {
	if s.data.Error != nil {
		return components.RenderError(s.data.Error, "Overview", "Project Overview")
	}

	breadcrumb := components.Breadcrumb(
		fmt.Sprintf("Project %s", s.projectID),
		s.Name(),
	)

	// 1. Header Section (Status + Account)
	billingStatus := "ACTIVE"
	if !s.data.Info.Enabled {
		billingStatus = "DISABLED"
	}

	headerLeft := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("Billing: %s", components.RenderStatus(billingStatus)),
		fmt.Sprintf("Project: %s", s.projectID),
	)

	headerRight := lipgloss.JoinVertical(lipgloss.Left,
		fmt.Sprintf("💳 Account: %s", s.data.Info.BillingAccountName),
		fmt.Sprintf("ID: %s", s.data.Info.BillingAccountID),
	)

	// Top Header
	header := cardStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(styles.ColorBrandAccent).Bold(true).Render("📡 Project Overview"),
			"",
			lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(40).Render(headerLeft),
				headerRight,
			),
		),
	)

	// 2. Actionable Insights
	var insightsContent string
	if s.data.RecsLoading {
		insightsContent = components.RenderSpinner("Loading insights...")
	} else if len(s.data.Recommendations) == 0 {
		insightsContent = "✅ No active recommendations found (or Recommender API disabled)."
	} else {
		// Aggregations
		type Category struct {
			Count    int
			Savings  float64
			Title    string
			Icon     string
			Action   string
			Currency string
		}

		cats := map[string]*Category{
			"IDLE_VM":        {Title: "Idle VMs", Icon: "🛑", Action: "Stop"},
			"UNUSED_ADDRESS": {Title: "Unused IPs", Icon: "🗑️ ", Action: "Release"},
			"GHOST_DISK":     {Title: "Ghost Disks", Icon: "💾", Action: "Snapshot & Delete"},
			"RESIZE":         {Title: "Oversized VMs", Icon: "📉", Action: "Resize"},
		}

		for _, r := range s.data.Recommendations {
			key := ""
			desc := strings.ToLower(r.Description)
			subtype := r.RecommenderSubtype

			if subtype == "IDLE_VM" {
				key = "IDLE_VM"
			} else if subtype == "UNUSED_ADDRESS" {
				key = "UNUSED_ADDRESS"
			} else if subtype == "SNAPSHOT_AND_DELETE_DISK" || (subtype == "IDLE_RESOURCE" && strings.Contains(desc, "disk")) {
				key = "GHOST_DISK"
			} else if subtype == "CHANGE_MACHINE_TYPE" {
				key = "RESIZE"
			} else {
				// Dynamic category for others
				key = subtype
			}

			// Initialize dynamic category if needed
			if cats[key] == nil {
				readable := strings.ReplaceAll(key, "_", " ")
				readable = strings.ToLower(readable)
				words := strings.Fields(readable)
				for i, w := range words {
					if len(w) > 0 {
						words[i] = strings.ToUpper(string(w[0])) + w[1:]
					}
				}
				readable = strings.Join(words, " ")

				cats[key] = &Category{
					Title:  readable,
					Icon:   "💡",
					Action: "Check Console",
				}
			}

			cats[key].Count++
			cats[key].Savings += r.EstimatedSavingsAmount
			if r.CurrencyCode != "" {
				cats[key].Currency = r.CurrencyCode
			}
		}

		// Order of display: Known first, then others
		knownOrder := []string{"IDLE_VM", "UNUSED_ADDRESS", "GHOST_DISK", "RESIZE"}
		var otherKeys []string
		for k := range cats {
			isKnown := false
			for _, known := range knownOrder {
				if k == known {
					isKnown = true
					break
				}
			}
			if !isKnown {
				otherKeys = append(otherKeys, k)
			}
		}

		displayOrder := append(knownOrder, otherKeys...)

		var lines []string
		for _, k := range displayOrder {
			c := cats[k]
			if c != nil && c.Count > 0 {
				line := fmt.Sprintf("%s %d %s", c.Icon, c.Count, c.Title)
				if c.Savings > 0 {
					curr := "USD"
					if c.Currency != "" {
						curr = c.Currency
					}
					line += fmt.Sprintf(" (Save %.2f %s/mo)", c.Savings, curr)
				}
				line += fmt.Sprintf(" - %s", c.Action)
				lines = append(lines, line)
			}
		}

		if len(lines) > 0 {
			insightsContent = strings.Join(lines, "\n")
		} else {
			insightsContent = "✅ No major cost issues found."
		}
	}

	savingsSection := cardStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("⚡ Actionable Insights"),
			insightsContent,
		),
	)

	// 3. Resource Inventory (Expanded)
	var inventoryContent string
	if s.data.InventoryLoading {
		inventoryContent = "⏳ Scanning resources..."
	} else {
		inv := s.data.Inventory

		// Row 1: Compute
		r1 := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("🖥️  VMs:      %d", inv.InstanceCount)),
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("💾 Disks:    %d (%dGB)", inv.DiskCount, inv.DiskGB)),
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("🌐 IPs:      %d", inv.IPCount)),
		)

		// Row 2: Data
		r2 := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("🗄️  SQL:      %d", inv.SQLCount)),
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("🪣 Buckets:  %d", inv.BucketCount)),
			lipgloss.NewStyle().Width(25).Render(fmt.Sprintf("🔍 Datasets: %d", inv.DatasetCount)),
		)

		inventoryContent = lipgloss.JoinVertical(lipgloss.Left, r1, "", r2)
	}

	inventorySection := cardStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("📦 Global Resource Inventory"),
			inventoryContent,
		),
	)

	// 4. Budgets
	var budgetContent string
	if s.data.BudgetsLoading {
		budgetContent = components.RenderSpinner("Loading budgets...")
	} else if len(s.data.Budgets) == 0 {
		budgetContent = "No budgets configured (or permission denied)."
	} else {
		for _, b := range s.data.Budgets {
			budgetContent += fmt.Sprintf("• %s: %s %s\n", b.Name, b.BudgetAmount, b.CurrencyCode)
		}
	}
	budgetSection := cardStyle.Copy().Width(80).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render("💰 Budget Radar"),
			budgetContent,
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		breadcrumb,
		"",
		header,
		savingsSection,
		inventorySection,
		budgetSection,
	)
}
