package klutchaws

import "testing"

func TestParseOnDemandPrice(t *testing.T) {
	raw := []byte(`{
		"terms": {
			"OnDemand": {
				"ABC": {
					"priceDimensions": {
						"XYZ": {
							"pricePerUnit": {
								"USD": "0.0123"
							}
						}
					}
				}
			}
		}
	}`)

	price, err := parseOnDemandPrice(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0.0123 {
		t.Fatalf("expected 0.0123, got %f", price)
	}
}

func TestRegionToLocation(t *testing.T) {
	loc, err := regionToLocation("eu-central-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc != "EU (Frankfurt)" {
		t.Fatalf("unexpected location %s", loc)
	}

	_, err = regionToLocation("moon-1")
	if err == nil {
		t.Fatalf("expected error for unknown region")
	}
}
