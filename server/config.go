package main

import (
	"time"

	"github.com/stripe/stripe-go/v81"
)

var loc, _ = time.LoadLocation("America/New_York")

type PricePoint struct {
	Time  int64  `json:"time"`
	Price uint64 `json:"price"`
}

var priceRange = struct {
	Start PricePoint `json:"start"`
	End   PricePoint `json:"end"`
}{
	PricePoint{
		time.Date(2024, 12, 15, 0, 0, 0, 0, loc).UnixMilli(),
		1500,
	},
	PricePoint{
		time.Date(2025, 1, 1, 0, 0, 0, 0, loc).UnixMilli(),
		4000,
	},
}

var paymentIntentParams = &stripe.PaymentIntentParams{
	Currency:                  stripe.String(string(stripe.CurrencyUSD)),
	StatementDescriptorSuffix: stripe.String("NYE 2025"),
}
