package net

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/compute/v1"
)

type Client struct {
	service *compute.Service
}

func NewClient(ctx context.Context) (*Client, error) {
	s, err := compute.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute service: %w", err)
	}
	return &Client{service: s}, nil
}

func (c *Client) ListNetworks(projectID string) ([]Network, error) {
	var networks []Network
	req := c.service.Networks.List(projectID)
	if err := req.Pages(context.Background(), func(page *compute.NetworkList) error {
		for _, n := range page.Items {
			mode := "CUSTOM"
			if n.AutoCreateSubnetworks {
				mode = "AUTO"
			} else if n.IPv4Range != "" {
				mode = "LEGACY"
			}

			networks = append(networks, Network{
				Name:        n.Name,
				ID:          n.Id,
				SelfLink:    n.SelfLink,
				IPv4Range:   n.IPv4Range,
				Mode:        mode,
				GatewayIPv4: n.GatewayIPv4,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return networks, nil
}

func (c *Client) ListSubnets(projectID string, networkLink string) ([]Subnet, error) {
	// Subnets are regional (AggregatedList)
	var subnets []Subnet
	req := c.service.Subnetworks.AggregatedList(projectID)
	// Filter by network? API filter string: "network eq link"
	req.Filter(fmt.Sprintf("network eq \"%s\"", networkLink))

	if err := req.Pages(context.Background(), func(page *compute.SubnetworkAggregatedList) error {
		for _, items := range page.Items {
			for _, s := range items.Subnetworks {
				subnets = append(subnets, Subnet{
					Name:        s.Name,
					Region:      extractRegion(s.Region),
					IPCidrRange: s.IpCidrRange,
					Gateway:     s.GatewayAddress,
					Network:     s.Network,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return subnets, nil
}

func (c *Client) ListFirewalls(projectID string, networkLink string) ([]Firewall, error) {
	var firewalls []Firewall
	req := c.service.Firewalls.List(projectID)
	req.Filter(fmt.Sprintf("network eq \"%s\"", networkLink))

	if err := req.Pages(context.Background(), func(page *compute.FirewallList) error {
		for _, f := range page.Items {
			action := "ALLOW"
			if len(f.Denied) > 0 {
				action = "DENY"
			}

			direction := f.Direction

			// Source/Target formatting
			var source string
			if direction == "INGRESS" {
				if len(f.SourceRanges) > 0 {
					source = fmt.Sprintf("IPs: %v", truncateList(f.SourceRanges))
				} else if len(f.SourceTags) > 0 {
					source = fmt.Sprintf("Tags: %v", truncateList(f.SourceTags))
				} else {
					source = "All"
				}
			} else {
				if len(f.DestinationRanges) > 0 {
					source = fmt.Sprintf("Dest: %v", truncateList(f.DestinationRanges))
				} else {
					source = "All"
				}
			}

			var target string
			if len(f.TargetTags) > 0 {
				target = fmt.Sprintf("Tags: %v", truncateList(f.TargetTags))
			} else {
				target = "All Instances"
			}

			firewalls = append(firewalls, Firewall{
				Name:      f.Name,
				Network:   f.Network,
				Direction: direction,
				Priority:  f.Priority,
				Action:    action,
				Source:    source,
				Target:    target,
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return firewalls, nil
}

// Helpers

func extractRegion(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func truncateList(list []string) string {
	if len(list) > 2 {
		return fmt.Sprintf("[%s, %s, +%d]", list[0], list[1], len(list)-2)
	}
	return fmt.Sprintf("%v", list)
}
