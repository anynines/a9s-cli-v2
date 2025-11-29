package klutchaws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/aws/aws-sdk-go-v2/service/pricing/types"
)

const monthlyHours = 730.0

// EstimateConfig contains input knobs for Klutch AWS cost estimation.
type EstimateConfig struct {
	Region        string
	InstanceType  string
	DesiredNodes  int
	MinNodes      int
	MaxNodes      int
	NodeDiskGiB   int
	NatGateways   int
	PricingRegion string // optional override; defaults to us-east-1 where Pricing is available.
}

// Item describes a single bill-of-material entry.
type Item struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unitPrice"`
	Hourly    float64 `json:"hourly"`
	Monthly   float64 `json:"monthly"`
	Notes     string  `json:"notes,omitempty"`
}

// Report is the final estimation result.
type Report struct {
	Region       string   `json:"region"`
	Currency     string   `json:"currency"`
	Items        []Item   `json:"items"`
	TotalHourly  float64  `json:"totalHourly"`
	TotalMonthly float64  `json:"totalMonthly"`
	Assumptions  []string `json:"assumptions"`
	NotIncluded  []string `json:"notIncluded"`
}

// Estimate calculates a Klutch control plane BOM with AWS list pricing.
func Estimate(ctx context.Context, cfg EstimateConfig) (*Report, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("region must be provided")
	}
	if cfg.InstanceType == "" {
		return nil, fmt.Errorf("instance type must be provided")
	}
	if cfg.DesiredNodes <= 0 {
		cfg.DesiredNodes = 1
	}
	if cfg.NatGateways <= 0 {
		cfg.NatGateways = 3
	}
	if cfg.NodeDiskGiB <= 0 {
		cfg.NodeDiskGiB = 80
	}
	if cfg.PricingRegion == "" {
		cfg.PricingRegion = "us-east-1"
	}

	location, err := regionToLocation(cfg.Region)
	if err != nil {
		return nil, err
	}

	pricer, err := newPricingService(ctx, cfg.PricingRegion)
	if err != nil {
		return nil, err
	}

	report := Report{
		Region:   cfg.Region,
		Currency: "USD",
		Assumptions: []string{
			"On-Demand, Linux, shared tenancy nodes",
			"Control plane priced via Amazon EKS cluster fee",
			fmt.Sprintf("Node root volume assumed %d GiB gp3", cfg.NodeDiskGiB),
			fmt.Sprintf("NAT gateways assumed per public subnet (%d total)", cfg.NatGateways),
		},
		NotIncluded: []string{
			"Data transfer and NAT data processing",
			"Load balancer usage (ALB/NLB) and LCU consumption",
			"ECR, S3, or other ancillary services",
		},
	}

	// EKS control plane
	eksPrice, err := pricer.eksControlPlane(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("fetching EKS control plane price: %w", err)
	}
	report.Items = append(report.Items, Item{
		Name:      "EKS Control Plane",
		Category:  "control-plane",
		Quantity:  1,
		Unit:      "hours",
		UnitPrice: eksPrice,
		Hourly:    eksPrice,
		Monthly:   eksPrice * monthlyHours,
	})

	// Node compute
	nodePrice, err := pricer.ec2Instance(ctx, cfg.InstanceType, location)
	if err != nil {
		return nil, fmt.Errorf("fetching EC2 price for %s: %w", cfg.InstanceType, err)
	}
	nodesHourly := float64(cfg.DesiredNodes) * nodePrice
	report.Items = append(report.Items, Item{
		Name:      fmt.Sprintf("Worker Nodes (%s)", cfg.InstanceType),
		Category:  "compute",
		Quantity:  float64(cfg.DesiredNodes),
		Unit:      "hours",
		UnitPrice: nodePrice,
		Hourly:    nodesHourly,
		Monthly:   nodesHourly * monthlyHours,
	})

	// Root volumes per node
	ebsPrice, err := pricer.gp3Storage(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("fetching gp3 storage price: %w", err)
	}
	ebsQty := float64(cfg.DesiredNodes * cfg.NodeDiskGiB)
	ebsMonthly := ebsQty * ebsPrice
	report.Items = append(report.Items, Item{
		Name:      "Worker Root Volumes (gp3)",
		Category:  "storage",
		Quantity:  ebsQty,
		Unit:      "GB-month",
		UnitPrice: ebsPrice,
		Hourly:    ebsMonthly / monthlyHours,
		Monthly:   ebsMonthly,
		Notes:     fmt.Sprintf("%d GiB per node", cfg.NodeDiskGiB),
	})

	// NAT gateways
	natPrice, err := pricer.natGateway(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("fetching NAT gateway price: %w", err)
	}
	natHourly := float64(cfg.NatGateways) * natPrice
	report.Items = append(report.Items, Item{
		Name:      "NAT Gateway",
		Category:  "network",
		Quantity:  float64(cfg.NatGateways),
		Unit:      "hours",
		UnitPrice: natPrice,
		Hourly:    natHourly,
		Monthly:   natHourly * monthlyHours,
		Notes:     "Excludes data processing",
	})

	// KMS CMK
	kmsPrice, err := pricer.kmsKey(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("fetching KMS key price: %w", err)
	}
	report.Items = append(report.Items, Item{
		Name:      "KMS Customer-Managed Key",
		Category:  "security",
		Quantity:  1,
		Unit:      "months",
		UnitPrice: kmsPrice,
		Hourly:    kmsPrice / monthlyHours,
		Monthly:   kmsPrice,
		Notes:     "First 30-day key fee",
	})

	for _, item := range report.Items {
		report.TotalHourly += item.Hourly
		report.TotalMonthly += item.Monthly
	}

	return &report, nil
}

// pricingService wraps AWS Pricing with a small cache.
type pricingService struct {
	client *pricing.Client
	cache  map[string]float64
}

func newPricingService(ctx context.Context, pricingRegion string) (*pricingService, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(pricingRegion))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return &pricingService{
		client: pricing.NewFromConfig(cfg),
		cache:  make(map[string]float64),
	}, nil
}

func (p *pricingService) cacheKey(parts ...string) string {
	return strings.Join(parts, "|")
}

func (p *pricingService) getPrice(ctx context.Context, service string, filters []types.Filter) (float64, error) {
	keyParts := []string{service}
	for _, f := range filters {
		keyParts = append(keyParts, fmt.Sprintf("%s=%s", aws.ToString(f.Field), aws.ToString(f.Value)))
	}
	key := p.cacheKey(keyParts...)
	if val, ok := p.cache[key]; ok {
		return val, nil
	}

	out, err := p.client.GetProducts(ctx, &pricing.GetProductsInput{
		ServiceCode: aws.String(service),
		Filters:     filters,
		MaxResults:  aws.Int32(1),
	})
	if err != nil {
		return 0, err
	}
	if len(out.PriceList) == 0 {
		return 0, fmt.Errorf("no pricing data returned for %s", key)
	}

	rawMap, err := decodePriceString(out.PriceList[0])
	if err != nil {
		return 0, err
	}
	rawBytes, err := json.Marshal(rawMap)
	if err != nil {
		return 0, err
	}
	price, err := parseOnDemandPrice(rawBytes)
	if err != nil {
		return 0, err
	}

	p.cache[key] = price
	return price, nil
}

// getPriceWithPredicate scans products for a match and returns the first price that satisfies the predicate.
func (p *pricingService) getPriceWithPredicate(ctx context.Context, service string, filters []types.Filter, key string, pred func(map[string]string) bool) (float64, error) {
	if key != "" {
		if val, ok := p.cache[key]; ok {
			return val, nil
		}
	}

	var next *string
	for {
		out, err := p.client.GetProducts(ctx, &pricing.GetProductsInput{
			ServiceCode: aws.String(service),
			Filters:     filters,
			MaxResults:  aws.Int32(100),
			NextToken:   next,
		})
		if err != nil {
			return 0, err
		}
		for _, pl := range out.PriceList {
			m, err := decodePriceString(pl)
			if err != nil {
				continue
			}
			attrs := extractAttributes(m)
			if pred(attrs) {
				raw, err := json.Marshal(m)
				if err != nil {
					return 0, err
				}
				price, err := parseOnDemandPrice(raw)
				if err != nil {
					return 0, err
				}
				if key != "" {
					p.cache[key] = price
				}
				return price, nil
			}
		}
		if out.NextToken == nil || aws.ToString(out.NextToken) == "" {
			break
		}
		next = out.NextToken
	}
	return 0, fmt.Errorf("no pricing data matched predicate for %s", service)
}

func (p *pricingService) ec2Instance(ctx context.Context, instanceType, location string) (float64, error) {
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
		{Type: types.FilterTypeTermMatch, Field: aws.String("operatingSystem"), Value: aws.String("Linux")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("preInstalledSw"), Value: aws.String("NA")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("tenancy"), Value: aws.String("Shared")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("capacitystatus"), Value: aws.String("Used")},
	}
	return p.getPrice(ctx, "AmazonEC2", filters)
}

func (p *pricingService) gp3Storage(ctx context.Context, location string) (float64, error) {
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("volumeApiName"), Value: aws.String("gp3")},
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
	}
	return p.getPrice(ctx, "AmazonEC2", filters)
}

func (p *pricingService) natGateway(ctx context.Context, location string) (float64, error) {
	pred := func(attrs map[string]string) bool {
		usage := strings.ToUpper(attrs["usagetype"])
		group := strings.ToUpper(attrs["group"])
		return strings.Contains(usage, "NATGATEWAY") || strings.Contains(group, "NAT")
	}

	primaryFilters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
	}
	key := p.cacheKey("AmazonVPC", location, "NATGATEWAY")
	if price, err := p.getPriceWithPredicate(ctx, "AmazonVPC", primaryFilters, key, pred); err == nil {
		return price, nil
	}

	// Fallback: try without location filter (some regions label differently).
	if price, err := p.getPriceWithPredicate(ctx, "AmazonVPC", nil, p.cacheKey("AmazonVPC", "NATGATEWAY"), pred); err == nil {
		return price, nil
	}

	// Fallback: NAT occasionally appears under AmazonEC2 pricing.
	if price, err := p.getPriceWithPredicate(ctx, "AmazonEC2", primaryFilters, p.cacheKey("AmazonEC2", location, "NATGATEWAY"), pred); err == nil {
		return price, nil
	}
	if price, err := p.getPriceWithPredicate(ctx, "AmazonEC2", nil, p.cacheKey("AmazonEC2", "NATGATEWAY"), pred); err == nil {
		return price, nil
	}

	return 0, fmt.Errorf("no pricing data matched predicate for NAT gateway")
}

func (p *pricingService) eksControlPlane(ctx context.Context, location string) (float64, error) {
	filters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
	}
	key := p.cacheKey("AmazonEKS", location, "EKSCLUSTER")
	return p.getPriceWithPredicate(ctx, "AmazonEKS", filters, key, func(attrs map[string]string) bool {
		usage := strings.ToUpper(attrs["usagetype"])
		return strings.Contains(usage, "EKS-HOURS:PERCLUSTER")
	})
}

func (p *pricingService) kmsKey(ctx context.Context, location string) (float64, error) {
	pred := func(attrs map[string]string) bool {
		usage := strings.ToUpper(attrs["usagetype"])
		group := strings.ToUpper(attrs["group"])
		return strings.Contains(usage, "KMS-KEY") || strings.Contains(group, "KMS")
	}

	primaryFilters := []types.Filter{
		{Type: types.FilterTypeTermMatch, Field: aws.String("location"), Value: aws.String(location)},
	}
	key := p.cacheKey("AWSKMS", location, "KMSKEY")
	if price, err := p.getPriceWithPredicate(ctx, "AWSKMS", primaryFilters, key, pred); err == nil {
		return price, nil
	}

	// Fallback: try without location filter.
	if price, err := p.getPriceWithPredicate(ctx, "AWSKMS", nil, p.cacheKey("AWSKMS", "KMSKEY"), pred); err == nil {
		return price, nil
	}

	return 0, fmt.Errorf("no pricing data matched predicate for KMS key")
}

func parseOnDemandPrice(raw []byte) (float64, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, err
	}
	terms, ok := payload["terms"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("terms missing in pricing payload")
	}
	onDemand, ok := terms["OnDemand"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("OnDemand terms missing in pricing payload")
	}
	for _, term := range onDemand {
		termMap, ok := term.(map[string]interface{})
		if !ok {
			continue
		}
		pd, ok := termMap["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}
		for _, d := range pd {
			dim, ok := d.(map[string]interface{})
			if !ok {
				continue
			}
			unitMap, ok := dim["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}
			if usdStr, ok := unitMap["USD"].(string); ok {
				val, err := strconv.ParseFloat(usdStr, 64)
				if err == nil {
					return val, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("could not parse USD price from pricing payload")
}

func decodePriceString(raw string) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractAttributes(raw map[string]interface{}) map[string]string {
	attrs := map[string]string{}
	product, ok := raw["product"].(map[string]interface{})
	if !ok {
		return attrs
	}
	a, ok := product["attributes"].(map[string]interface{})
	if !ok {
		return attrs
	}
	for k, v := range a {
		if s, ok := v.(string); ok {
			attrs[strings.ToLower(k)] = s
		}
	}
	return attrs
}

// regionToLocation translates AWS region codes to Pricing API location strings.
func regionToLocation(region string) (string, error) {
	locations := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"ca-central-1":   "Canada (Central)",
		"eu-central-1":   "EU (Frankfurt)",
		"eu-west-1":      "EU (Ireland)",
		"eu-west-2":      "EU (London)",
		"eu-west-3":      "EU (Paris)",
		"eu-north-1":     "EU (Stockholm)",
		"eu-south-1":     "EU (Milan)",
		"eu-south-2":     "EU (Spain)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-southeast-3": "Asia Pacific (Jakarta)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-northeast-3": "Asia Pacific (Osaka)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-south-2":     "Asia Pacific (Hyderabad)",
		"sa-east-1":      "South America (São Paulo)",
		"af-south-1":     "Africa (Cape Town)",
		"me-south-1":     "Middle East (Bahrain)",
		"me-central-1":   "Middle East (UAE)",
	}
	if loc, ok := locations[region]; ok {
		return loc, nil
	}
	return "", fmt.Errorf("unsupported or unknown region for pricing: %s", region)
}
