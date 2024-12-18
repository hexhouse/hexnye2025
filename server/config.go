package main

import (
	"github.com/stripe/stripe-go/v81"
)

var paymentIntentParams = &stripe.PaymentIntentParams{
	Currency:                  stripe.String(string(stripe.CurrencyUSD)),
	Amount:                    stripe.Int64(1500),
	StatementDescriptorSuffix: stripe.String("NYE 2025"),
}
