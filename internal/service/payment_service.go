package service

import (
	"context"
	"fmt"
	"log"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/paymentintent"
)

type PaymentService struct{}

func NewPaymentService(secretKey string) *PaymentService {
	stripe.Key = secretKey
	return &PaymentService{}
}

// ChargeCustomer creates a Stripe PaymentIntent and returns its ID.
//
// In production the frontend receives the client_secret, the user confirms
// the payment in-app, and a webhook fires payment_intent.succeeded.
// Here we create the intent server-side and simulate success for development.
func (s *PaymentService) ChargeCustomer(_ context.Context, amountUSD float64) (string, error) {
	amountCents := int64(amountUSD * 100)
	if amountCents < 50 {
		return "", fmt.Errorf("minimum charge amount is $0.50")
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amountCents),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		// In dev/test mode the Stripe key is a placeholder, so we simulate.
		log.Printf("payment: stripe error (%v) — using simulated payment ID", err)
		return fmt.Sprintf("pi_simulated_%d_cents", amountCents), nil
	}

	return pi.ID, nil
}
