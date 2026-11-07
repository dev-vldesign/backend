package controllers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	session "github.com/stripe/stripe-go/v74/checkout/session"
)

func CreateCheckoutUrl(c *gin.Context) {
    var body struct {
        Amount       int64  `form:"amount" json:"amount"`
        ReceiptEmail string `form:"email"  json:"email"`
        Metadata     string `form:"metadata" json:"metadata"`
    }

    if err := c.Bind(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Failed to read request body",
        })
        return
    }

    stripeKey := ""

    if stripeKey == "" {
        log.Println("STRIPE_SECRET_KEY environment variable not set.")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe secret key not configured"})
        return
    }
    stripe.Key = stripeKey

    domain := "https://localhost:5173"

    params := &stripe.CheckoutSessionParams{
        CustomerEmail: &body.ReceiptEmail,
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {
                PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
                    Currency: stripe.String("usd"),
                    ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
                        Name: stripe.String("Product Name Here"),
                    },
                    UnitAmount:  stripe.Int64(body.Amount * 100),
                    TaxBehavior: stripe.String("exclusive"),
                },
                Quantity: stripe.Int64(1),
            },
        },
        Mode:        stripe.String(string(stripe.CheckoutSessionModePayment)),
        SuccessURL:  stripe.String(domain + "/success"),
        CancelURL:   stripe.String(domain + "/cancel"),
        AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
            Enabled: stripe.Bool(true),
        },
    }

    params.AddMetadata("id", body.Metadata)

    fmt.Println("Creating checkout session with params:", params)

    s, err := session.New(params)
    if err != nil {
        log.Printf("Error creating session: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "checkout_url": s.URL,
    })
}
