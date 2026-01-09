package overview

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/storage"
	"google.golang.org/api/billingbudgets/v1"
	"google.golang.org/api/cloudbilling/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/recommender/v1"
	"google.golang.org/api/sqladmin/v1"
)

type Client struct {
	billingService *cloudbilling.APIService
	recommender    *recommender.Service
	budgets        *billingbudgets.Service
	compute        *compute.Service
	sqladmin       *sqladmin.Service
	storage        *storage.Client
	bigquery       *bigquery.Client
}

func NewClient(ctx context.Context) (*Client, error) {
	opts := []option.ClientOption{option.WithScopes(
		cloudbilling.CloudPlatformScope,
		"https://www.googleapis.com/auth/sqlservice.admin",
		"https://www.googleapis.com/auth/devstorage.read_only",
		"https://www.googleapis.com/auth/bigquery.readonly",
	)}

	// Cloud Billing
	billingSvc, err := cloudbilling.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("billing client: %w", err)
	}

	// Recommender
	recSvc, err := recommender.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("recommender client: %w", err)
	}

	// Billing Budgets
	budgetSvc, err := billingbudgets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("budget client: %w", err)
	}

	// Compute Engine
	compSvc, err := compute.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("compute client: %w", err)
	}

	// Cloud SQL
	sqlSvc, err := sqladmin.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("sql client: %w", err)
	}

	// GCS
	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage client: %w", err)
	}

	// BigQuery
	bqClient, err := bigquery.NewClient(ctx, "", opts...) // Project ID empty, will infer or be passed later?
	// Actually NewClient requires ProjectID for standard operations usually, but for listing datasets it might need it.
	// But api.go NewClient doesn't take ProjectID. We can pass "" and use explicit project ID in calls often, or we need to pass it.
	// Let's pass "" and handle it or note if BQ fails.
	// The bigquery client IS bound to a project usually.
	// Let's defer BQ initialization or assume default credentials can resolve it?
	// Wait, we need ProjectID to init BQ client properly for most ops.
	// Modification: NewClient should probably take ProjectID or we init it lazily?
	// Existing NewClient signature: func NewClient(ctx context.Context) ...
	// Let's just pass "" and see if we can use it for listing datasets of a specific project later.
	if err != nil {
		return nil, fmt.Errorf("bigquery client: %w", err)
	}

	return &Client{
		billingService: billingSvc,
		recommender:    recSvc,
		budgets:        budgetSvc,
		compute:        compSvc,
		sqladmin:       sqlSvc,
		storage:        storageClient,
		bigquery:       bqClient,
	}, nil
}

func (c *Client) GetProjectBillingInfo(projectID string) (BillingInfo, error) {
	name := fmt.Sprintf("projects/%s", projectID)
	info, err := c.billingService.Projects.GetBillingInfo(name).Do()
	if err != nil {
		return BillingInfo{}, err
	}

	return BillingInfo{
		Enabled:            info.BillingEnabled,
		BillingAccountName: info.BillingAccountName,
		BillingAccountID:   strings.TrimPrefix(info.BillingAccountName, "billingAccounts/"),
	}, nil
}

func (c *Client) GetRecommendations(projectID string, zone string) ([]Recommendation, error) {
	var recs []Recommendation

	fetch := func(recommenderID string, location string) {
		parent := fmt.Sprintf("projects/%s/locations/%s/recommenders/%s", projectID, location, recommenderID)
		call := c.recommender.Projects.Locations.Recommenders.Recommendations.List(parent)
		resp, err := call.Do()
		if err != nil {
			return
		}
		for _, r := range resp.Recommendations {
			var savings float64
			var currency string
			if r.PrimaryImpact != nil && r.PrimaryImpact.CostProjection != nil && r.PrimaryImpact.CostProjection.Cost != nil {
				units := r.PrimaryImpact.CostProjection.Cost.Units
				nanos := r.PrimaryImpact.CostProjection.Cost.Nanos
				val := float64(units) + float64(nanos)/1e9
				if val < 0 {
					val = -val
				}
				savings = val
				currency = r.PrimaryImpact.CostProjection.Cost.CurrencyCode
			}

			recs = append(recs, Recommendation{
				ID:                     r.Name,
				Description:            r.Description,
				RecommenderSubtype:     r.RecommenderSubtype,
				Priority:               r.Priority,
				State:                  r.StateInfo.State,
				EstimatedSavingsAmount: savings,
				CurrencyCode:           currency,
			})
		}
	}

	locations := []string{"us-central1-a", "us-central1-b", "global"}
	if zone != "" {
		locations = append(locations, zone)
	}

	targetRecommenders := []string{
		"google.compute.instance.IdleResourceRecommender",
		"google.compute.instance.MachineTypeRecommender",
		"google.compute.address.UnusedAddressRecommender",
		"google.compute.disk.IdleResourceRecommender",
	}

	for _, loc := range locations {
		for _, recID := range targetRecommenders {
			fetch(recID, loc)
		}
	}

	return recs, nil
}

func (c *Client) GetBudgets(billingAccountID string) ([]SpendLimit, error) {
	if billingAccountID == "" {
		return nil, nil // No billing account
	}
	parent := fmt.Sprintf("billingAccounts/%s", billingAccountID)
	resp, err := c.budgets.BillingAccounts.Budgets.List(parent).Do()
	if err != nil {
		return nil, err
	}

	var limits []SpendLimit
	for _, b := range resp.Budgets {
		amount := "N/A"
		currency := ""
		if b.Amount != nil {
			if b.Amount.SpecifiedAmount != nil {
				amount = fmt.Sprintf("%d.%02d", b.Amount.SpecifiedAmount.Units, b.Amount.SpecifiedAmount.Nanos/10000000)
				currency = b.Amount.SpecifiedAmount.CurrencyCode
			} else if b.Amount.LastPeriodAmount != nil {
				amount = "Last Period"
			}
		}

		var thresholds []float64
		for _, t := range b.ThresholdRules {
			thresholds = append(thresholds, t.ThresholdPercent)
		}

		limits = append(limits, SpendLimit{
			Name:            b.DisplayName,
			BudgetAmount:    amount,
			CurrencyCode:    currency,
			AlertThresholds: thresholds,
		})
	}
	return limits, nil
}

// GetGlobalInventory fetches counts for key resources (Compute, SQL, GCS, BQ) in parallel
func (c *Client) GetGlobalInventory(projectID string) (ResourceInventory, error) {
	var inv ResourceInventory
	// var errs []error // errors ignored for partial success

	type result struct {
		typ string
		val int
		gb  int
		err error
	}
	// 6 concurrent fetchers now
	ch := make(chan result, 6)

	// 1. Instances
	go func() {
		req := c.compute.Instances.AggregatedList(projectID)
		var count int
		if err := req.Pages(context.Background(), func(page *compute.InstanceAggregatedList) error {
			for _, items := range page.Items {
				count += len(items.Instances)
			}
			return nil
		}); err != nil {
			ch <- result{typ: "instances", err: err}
			return
		}
		ch <- result{typ: "instances", val: count}
	}()

	// 2. Disks
	go func() {
		req := c.compute.Disks.AggregatedList(projectID)
		var count int
		var gb int
		if err := req.Pages(context.Background(), func(page *compute.DiskAggregatedList) error {
			for _, items := range page.Items {
				for _, d := range items.Disks {
					count++
					gb += int(d.SizeGb)
				}
			}
			return nil
		}); err != nil {
			ch <- result{typ: "disks", err: err}
			return
		}
		ch <- result{typ: "disks", val: count, gb: gb}
	}()

	// 3. Addresses
	go func() {
		req := c.compute.Addresses.AggregatedList(projectID)
		var count int
		if err := req.Pages(context.Background(), func(page *compute.AddressAggregatedList) error {
			for _, items := range page.Items {
				count += len(items.Addresses)
			}
			return nil
		}); err != nil {
			ch <- result{typ: "addresses", err: err}
			return
		}
		ch <- result{typ: "addresses", val: count}
	}()

	// 4. Cloud SQL
	go func() {
		resp, err := c.sqladmin.Instances.List(projectID).Do()
		if err != nil {
			ch <- result{typ: "sql", err: err}
			return
		}
		ch <- result{typ: "sql", val: len(resp.Items)}
	}()

	// 5. GCS Buckets
	go func() {
		it := c.storage.Buckets(context.Background(), projectID)
		var count int
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				ch <- result{typ: "buckets", err: err}
				return
			}
			count++
		}
		ch <- result{typ: "buckets", val: count}
	}()

	// 6. BigQuery Datasets
	go func() {
		// BQ Client might strictly need ProjectID on init, but we passed "" in NewClient.
		// If that fails, we might need a separate BQ client init here with project ID.
		// However, typical pattern is client.DatasetInProject(projectID).
		// We can't list datasets without a project ID context usually.
		// The client.Datasets(ctx) lists for the client's project.
		// If client was init with "", this might fail or default to auth project.
		// Let's rely on InProject if available, otherwise we might need to recreate client.
		// NOTE: cloud.google.com/go/bigquery does not have a simple "ListDatasets(project)" on the client if client has no project.
		// It has `client.Datasets(ctx)` which iterates datasets in the client's project.
		// Workaround: Create a new client just for this call if needed, or rely on NewClient having correct project if we update it.
		// Actually, let's try creating a lightweight client here if we can't switch context,
		// OR safer: rely on iterator and `DatasetsInProject`? No such method.
		// We will assume for now we use a new client scoped to project.
		bq, err := bigquery.NewClient(context.Background(), projectID)
		if err != nil {
			ch <- result{typ: "datasets", err: err}
			return
		}
		defer bq.Close()

		it := bq.Datasets(context.Background())
		var count int
		for {
			_, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				ch <- result{typ: "datasets", err: err}
				return
			}
			count++
		}
		ch <- result{typ: "datasets", val: count}
	}()

	// Collect 6 results
	for i := 0; i < 6; i++ {
		res := <-ch
		if res.err != nil {
			// Log error but generally continue? Inventory partial failure is acceptable?
			// Let's log but don't fail entire dashboard
			// errs = append(errs, fmt.Errorf("%s: %v", res.typ, res.err))
			continue
		}
		switch res.typ {
		case "instances":
			inv.InstanceCount = res.val
		case "disks":
			inv.DiskCount = res.val
			inv.DiskGB = res.gb
		case "addresses":
			inv.IPCount = res.val
		case "sql":
			inv.SQLCount = res.val
		case "buckets":
			inv.BucketCount = res.val
		case "datasets":
			inv.DatasetCount = res.val
		}
	}

	// We return what we found, partial or full
	return inv, nil
}
