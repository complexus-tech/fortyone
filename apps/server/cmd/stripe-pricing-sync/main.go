// stripe-pricing-sync replaces the four public seat prices without migrating subscriptions.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/josemukorivo/config"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/client"
)

type priceSpec struct {
	lookupKey string
	amount    int64
	interval  stripe.PriceRecurringInterval
}

var prices = []priceSpec{
	{"pro_monthly", 900, stripe.PriceRecurringIntervalMonth},
	{"pro_yearly", 8640, stripe.PriceRecurringIntervalYear},
	{"business_monthly", 1500, stripe.PriceRecurringIntervalMonth},
	{"business_yearly", 14400, stripe.PriceRecurringIntervalYear},
}

func main() {
	apply := flag.Bool("apply", false, "replace Stripe prices (default: read-only dry run)")
	timeout := flag.Duration("timeout", 2*time.Minute, "total operation timeout")
	flag.Parse()
	if flag.NArg() != 0 || *timeout <= 0 {
		log.Fatal("unexpected arguments or invalid timeout")
	}
	var cfg struct {
		Stripe struct {
			SecretKey string `env:"STRIPE_SECRET_KEY"`
		}
	}
	if err := config.Parse("app", &cfg); err != nil {
		log.Fatal("load configuration: ", err)
	}
	key := strings.TrimSpace(cfg.Stripe.SecretKey)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("APP_STRIPE_SECRET_KEY"))
	}
	if key == "" {
		log.Fatal("set STRIPE_SECRET_KEY (or APP_STRIPE_SECRET_KEY) in the environment or .env")
	}
	api := &client.API{}
	api.Init(key, nil)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	if err := syncPrices(ctx, api, *apply, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func syncPrices(ctx context.Context, api *client.API, apply bool, out io.Writer) error {
	// Validate the entire catalog before the first write, including currency and billing shape.
	current := make([]*stripe.Price, len(prices))
	for i, spec := range prices {
		params := &stripe.PriceListParams{LookupKeys: []*string{stripe.String(spec.lookupKey)}}
		params.Context = ctx
		params.Limit = stripe.Int64(2)
		params.AddExpand("data.currency_options")
		params.AddExpand("data.product")
		iter := api.Prices.List(params)
		for iter.Next() {
			if current[i] != nil {
				return fmt.Errorf("%s: multiple prices found", spec.lookupKey)
			}
			current[i] = iter.Price()
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("load %s: %w", spec.lookupKey, err)
		}
		if err := validatePrice(current[i], spec); err != nil {
			return fmt.Errorf("%s: %w", spec.lookupKey, err)
		}
		if i > 0 && current[i].Livemode != current[0].Livemode {
			return fmt.Errorf("mixed live and test prices")
		}
	}
	mode := "DRY-RUN"
	if apply {
		mode = "APPLY"
	}
	fmt.Fprintf(out, "%s: livemode=%t; existing subscriptions retain their current prices\n", mode, current[0].Livemode)
	for i, spec := range prices {
		p := current[i]
		action := "replace"
		if p.UnitAmount == spec.amount {
			action = "unchanged"
		}
		fmt.Fprintf(out, "%s: %s %s USD %.2f -> %.2f per seat/%s (product=%s)\n", spec.lookupKey, action, p.ID, float64(p.UnitAmount)/100, float64(spec.amount)/100, spec.interval, p.Product.ID)
	}
	if !apply {
		return nil
	}

	for i, spec := range prices {
		old := current[i]
		// Store identity BEFORE transferring the lookup key. Historical subscriptions
		// still reference this Price and must remain recognizable by webhook sync.
		if old.Metadata[subscriptionsdomain.StripePriceLookupKeyMetadata] != spec.lookupKey {
			params := &stripe.PriceParams{Metadata: map[string]string{subscriptionsdomain.StripePriceLookupKeyMetadata: spec.lookupKey}}
			params.Context = ctx
			if _, err := api.Prices.Update(old.ID, params); err != nil {
				return fmt.Errorf("preserve identity on %s: %w", old.ID, err)
			}
		}
		if old.UnitAmount == spec.amount {
			continue
		}
		metadata := make(map[string]string, len(old.Metadata)+1)
		for k, v := range old.Metadata {
			metadata[k] = v
		}
		metadata[subscriptionsdomain.StripePriceLookupKeyMetadata] = spec.lookupKey
		params := &stripe.PriceParams{
			Product: stripe.String(old.Product.ID), Currency: stripe.String("usd"),
			UnitAmount: stripe.Int64(spec.amount), BillingScheme: stripe.String("per_unit"),
			LookupKey: stripe.String(spec.lookupKey), TransferLookupKey: stripe.Bool(true),
			Active: stripe.Bool(true), TaxBehavior: stripe.String(string(old.TaxBehavior)),
			Recurring: &stripe.PriceRecurringParams{Interval: stripe.String(string(spec.interval)), IntervalCount: stripe.Int64(1), UsageType: stripe.String("licensed")},
			Metadata:  metadata,
		}
		params.Context = ctx
		params.IdempotencyKey = stripe.String(fmt.Sprintf("fortyone-pricing-v1:%s:%s:%d", old.ID, spec.lookupKey, spec.amount))
		created, err := api.Prices.New(params)
		if err != nil {
			return fmt.Errorf("replace %s (rerun to resume): %w", spec.lookupKey, err)
		}
		fmt.Fprintf(out, "%s: replacement=%s previous=%s\n", spec.lookupKey, created.ID, old.ID)
	}
	return nil
}

func validatePrice(p *stripe.Price, spec priceSpec) error {
	if p == nil {
		return fmt.Errorf("existing lookup key missing; refusing to invent a product")
	}
	if p.ID == "" || !p.Active || p.Product == nil || p.Product.ID == "" || !p.Product.Active {
		return fmt.Errorf("expected an active price and product")
	}
	if p.Currency != stripe.CurrencyUSD || p.BillingScheme != stripe.PriceBillingSchemePerUnit || p.TransformQuantity != nil || p.CustomUnitAmount != nil {
		return fmt.Errorf("expected a standard USD per-seat price")
	}
	for currency := range p.CurrencyOptions {
		if currency != "usd" {
			return fmt.Errorf("additional currency %s requires an explicit pricing decision", currency)
		}
	}
	if p.Recurring == nil || p.Recurring.Interval != spec.interval || p.Recurring.IntervalCount != 1 || p.Recurring.UsageType != stripe.PriceRecurringUsageTypeLicensed || p.Recurring.TrialPeriodDays != 0 {
		return fmt.Errorf("unexpected recurring interval, usage type, or legacy trial setting")
	}
	if value := p.Metadata[subscriptionsdomain.StripePriceLookupKeyMetadata]; value != "" && value != spec.lookupKey {
		return fmt.Errorf("conflicting plan identity metadata")
	}
	if p.TaxBehavior != stripe.PriceTaxBehaviorExclusive && p.TaxBehavior != stripe.PriceTaxBehaviorInclusive && p.TaxBehavior != stripe.PriceTaxBehaviorUnspecified {
		return fmt.Errorf("unexpected tax behavior")
	}
	return nil
}
