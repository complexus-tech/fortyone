package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

func TestSyncPrices(t *testing.T) {
	for _, tc := range []struct {
		name           string
		apply, invalid bool
		wantWrites     int
	}{
		{name: "dry run"}, {name: "apply and repeat", apply: true, wantWrites: 8}, {name: "invalid last price prevents all writes", apply: true, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog := map[string]*stripe.Price{}
			for _, spec := range prices {
				catalog[spec.lookupKey] = &stripe.Price{
					ID: "price_" + spec.lookupKey, Active: true, LookupKey: spec.lookupKey,
					Product:  &stripe.Product{ID: "prod_" + strings.Split(spec.lookupKey, "_")[0], Active: true},
					Currency: stripe.CurrencyUSD, BillingScheme: stripe.PriceBillingSchemePerUnit, UnitAmount: 100,
					TaxBehavior: stripe.PriceTaxBehaviorExclusive,
					Recurring:   &stripe.PriceRecurring{Interval: spec.interval, IntervalCount: 1, UsageType: stripe.PriceRecurringUsageTypeLicensed},
					Metadata:    map[string]string{"existing": "preserved"},
				}
			}
			if tc.invalid {
				catalog["business_yearly"].Currency = stripe.CurrencyEUR
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					t.Error(err)
					http.Error(w, "bad form", 400)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				respond := func(v any) {
					if err := json.NewEncoder(w).Encode(v); err != nil {
						t.Error(err)
					}
				}
				if r.Method == http.MethodGet {
					p := catalog[r.Form.Get("lookup_keys[0]")]
					if p == nil {
						t.Errorf("unknown list query %s", r.URL)
						http.Error(w, "unknown key", 400)
						return
					}
					respond(map[string]any{"object": "list", "data": []*stripe.Price{p}})
					return
				}
				writes++
				if r.URL.Path == "/v1/prices" {
					key := r.Form.Get("lookup_key")
					old := catalog[key]
					if old == nil {
						t.Error("unexpected create")
						http.Error(w, "unknown price", 400)
						return
					}
					if old.Metadata["fortyone_lookup_key"] != key {
						t.Error("lookup transfer happened before identity was preserved")
					}
					for field, want := range map[string]string{"transfer_lookup_key": "true", "product": old.Product.ID, "tax_behavior": "exclusive", "metadata[existing]": "preserved", "metadata[fortyone_lookup_key]": key, "recurring[usage_type]": "licensed"} {
						if got := r.Form.Get(field); got != want {
							t.Errorf("%s=%q want %q", field, got, want)
						}
					}
					if r.Header.Get("Idempotency-Key") == "" {
						t.Error("missing idempotency key")
					}
					next := *old
					next.ID = "new_" + old.ID
					for _, spec := range prices {
						if spec.lookupKey == key {
							if got := r.Form.Get("unit_amount"); got != fmt.Sprint(spec.amount) {
								t.Errorf("amount=%s want %d", got, spec.amount)
							}
							if got := r.Form.Get("recurring[interval]"); got != string(spec.interval) {
								t.Errorf("interval=%s", got)
							}
							next.UnitAmount = spec.amount
						}
					}
					catalog[key] = &next
					respond(next)
					return
				}
				for key, p := range catalog {
					if r.URL.Path == "/v1/prices/"+p.ID {
						if got := r.Form.Get("metadata[fortyone_lookup_key]"); got != key {
							t.Errorf("metadata=%q", got)
						}
						p.Metadata["fortyone_lookup_key"] = key
						respond(p)
						return
					}
				}
				t.Errorf("unexpected request %s %s", r.Method, r.URL)
				http.Error(w, "unexpected request", 400)
			}))
			defer server.Close()
			backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{URL: stripe.String(server.URL), HTTPClient: server.Client(), MaxNetworkRetries: stripe.Int64(0)})
			api := &client.API{}
			api.Init("sk_test_fake", &stripe.Backends{API: backend, Connect: backend, Uploads: backend})
			err := syncPrices(context.Background(), api, tc.apply, io.Discard)
			if (err != nil) != tc.invalid {
				t.Fatalf("error=%v", err)
			}
			if writes != tc.wantWrites {
				t.Fatalf("writes=%d want %d", writes, tc.wantWrites)
			}
			if tc.apply && !tc.invalid {
				if err := syncPrices(context.Background(), api, true, io.Discard); err != nil {
					t.Fatal(err)
				}
				if writes != tc.wantWrites {
					t.Fatalf("repeat created extra mutations: %d", writes)
				}
			}
		})
	}
}
