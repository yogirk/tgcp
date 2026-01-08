package gce

import (
	"fmt"
	"strings"
)

// PricingMap holds estimated hourly costs (USD)
// This is a simplified static map. ideally this would be fetched from Cloud Billing API.
var prices = map[string]float64{
	// E2 Standard
	"e2-micro":      0.008,
	"e2-small":      0.016,
	"e2-medium":     0.033,
	"e2-highcpu-2":  0.049,
	"e2-standard-2": 0.067,
	"e2-standard-4": 0.134,
	"e2-standard-8": 0.268,

	// N1 Standard
	"n1-standard-1": 0.0475,
	"n1-standard-2": 0.0950,
	"n1-standard-4": 0.1900,

	// N2 Standard
	"n2-standard-2": 0.097,
	"n2-standard-4": 0.194,
	"n2-standard-8": 0.388,
}

// EstimateCost returns a formatted string estimate of the hourly cost
// It now includes disk costs
func EstimateCost(machineType string, zone string, disks []Disk) string {
	// Extract basic machine type
	parts := strings.Split(machineType, "/")
	mt := parts[len(parts)-1]

	vmPrice, ok := prices[mt]
	if !ok {
		// Even if VM unknown, we might have valid disk config, but for now return N/A if base VM unknown?
		// Or return >$X?
		// Let's stick to N/A for consistency, or 0.0 + disk
		vmPrice = 0.0
		// return "N/A"
	}

	// VM Region Adjustment
	multiplier := 1.0
	if strings.HasPrefix(zone, "asia") || strings.HasPrefix(zone, "europe") {
		multiplier = 1.1
	}
	totalVMPrice := vmPrice * multiplier

	// Disk Pricing
	// Standard: $0.04 / GB / month -> ~$0.0000548 / GB / hr
	// SSD:      $0.17 / GB / month -> ~$0.0002329 / GB / hr
	// Balanced: $0.10 / GB / month -> ~$0.0001370 / GB / hr
	// 730 hours per month
	const hrsPerMonth = 730.0
	totalDiskPrice := 0.0

	for _, d := range disks {
		var ratePerGB float64
		switch d.Type {
		case "pd-ssd":
			ratePerGB = 0.17 / hrsPerMonth
		case "pd-balanced":
			ratePerGB = 0.10 / hrsPerMonth
		default: // pd-standard or unknown
			ratePerGB = 0.04 / hrsPerMonth
		}
		totalDiskPrice += float64(d.SizeGB) * ratePerGB
	}

	totalPrice := totalVMPrice + totalDiskPrice
	if totalPrice == 0 {
		return "N/A"
	}

	return fmt.Sprintf("$%.3f/hr", totalPrice)
}
