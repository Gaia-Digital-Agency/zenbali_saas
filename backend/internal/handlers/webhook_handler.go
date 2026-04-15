package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/net1io/zenbali/internal/services"
	"github.com/net1io/zenbali/internal/utils"
	"github.com/stripe/stripe-go/v76/webhook"
)

type WebhookHandler struct {
	services *services.Services
}

func NewWebhookHandler(svcs *services.Services) *WebhookHandler {
	return &WebhookHandler{services: svcs}
}

func (h *WebhookHandler) HandleStripe(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading Stripe webhook body: %v", err)
		utils.BadRequest(w, "Error reading request body")
		return
	}

	// Get the webhook secret
	webhookSecret := h.services.Payment.GetWebhookSecret()

	// Verify webhook signature if secret is configured
	var event map[string]interface{}
	if webhookSecret != "" {
		sig := r.Header.Get("Stripe-Signature")
		evt, err := webhook.ConstructEvent(payload, sig, webhookSecret)
		if err != nil {
			log.Printf("Stripe webhook signature verification failed: %v", err)
			utils.BadRequest(w, "Invalid signature")
			return
		}
		// Use the verified event
		if err := json.Unmarshal(evt.Data.Raw, &event); err != nil {
			event = map[string]interface{}{"type": string(evt.Type)}
		}
		event["type"] = string(evt.Type)
		
		// Handle based on event type
		switch evt.Type {
		case "checkout.session.completed":
			var session struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(evt.Data.Raw, &session); err != nil {
				log.Printf("Error parsing checkout session: %v", err)
				utils.InternalError(w, "Error parsing event")
				return
			}

			if err := h.services.Payment.HandleSuccessfulPayment(r.Context(), session.ID); err != nil {
				log.Printf("Error handling successful payment: %v", err)
				utils.InternalError(w, "Error processing payment")
				return
			}

			log.Printf("Successfully processed payment for session: %s", session.ID)

		case "checkout.session.expired":
			var session struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(evt.Data.Raw, &session); err != nil {
				log.Printf("Error parsing expired session: %v", err)
				utils.InternalError(w, "Error parsing event")
				return
			}

			if err := h.services.Payment.HandleFailedPayment(r.Context(), session.ID); err != nil {
				log.Printf("Error handling expired payment: %v", err)
				// Don't return error, just log it
			}

			log.Printf("Marked expired session: %s", session.ID)

		default:
			log.Printf("Unhandled Stripe event type: %s", evt.Type)
		}
	} else {
		// No webhook secret configured, parse payload directly (dev mode)
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Printf("Error parsing Stripe webhook payload: %v", err)
			utils.BadRequest(w, "Invalid payload")
			return
		}

		eventType, _ := event["type"].(string)
		data, _ := event["data"].(map[string]interface{})
		object, _ := data["object"].(map[string]interface{})
		sessionID, _ := object["id"].(string)

		switch eventType {
		case "checkout.session.completed":
			if err := h.services.Payment.HandleSuccessfulPayment(r.Context(), sessionID); err != nil {
				log.Printf("Error handling successful payment: %v", err)
				utils.InternalError(w, "Error processing payment")
				return
			}
			log.Printf("Successfully processed payment for session: %s", sessionID)

		case "checkout.session.expired":
			if err := h.services.Payment.HandleFailedPayment(r.Context(), sessionID); err != nil {
				log.Printf("Error handling expired payment: %v", err)
			}
			log.Printf("Marked expired session: %s", sessionID)

		default:
			log.Printf("Unhandled Stripe event type: %s", eventType)
		}
	}

	// Return 200 to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received": true}`))
}
