package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"github.com/tylerwray/amazin/config"
	"github.com/tylerwray/amazin/event"
)

func webhookHandler(s Service) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const MaxBodyBytes = int64(65536)
		req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading request body: %v\\n", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		event, err := webhook.ConstructEvent(
			payload,
			req.Header.Get("Stripe-Signature"),
			s.Config.Stripe.WebhookSecret,
		)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
			w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		switch event.Type {
		case "payment_intent.succeeded":
			var paymentIntent stripe.PaymentIntent
			err := json.Unmarshal(event.Data.Raw, &paymentIntent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\\n", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Printf("payment_intent.succeeded: %+v\n", paymentIntent)
			s.Dispatcher.Send(event.Data.Raw)
		default:
			fmt.Fprintf(os.Stderr, "Unexpected event type: %s\n", event.Type)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type Service struct {
	Config     config.Values
	Dispatcher event.Dispatcher
}

func newService() Service {
	cfg, err := config.Read()
	if err != nil {
		panic(fmt.Errorf("fatal configuration error: %s", err))
	}

	return Service{cfg, event.NewDispatcher(cfg)}
}

func main() {
	s := newService()

	http.HandleFunc("/webhooks", webhookHandler(s))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
