package senders

import (
	"log"

	"github.com/wavefronthq/wavefront-sdk-go/internal/auth"
)

func tokenServiceForCfg(cfg *configuration) auth.Service {
	switch a := cfg.Authentication.(type) {
	case auth.APIToken:
		log.Println("The Wavefront SDK will use Direct Ingestion authenticated using an API Token.")
		return auth.NewWavefrontTokenService(a.Token)
	case auth.CSPClientCredentials:
		log.Println("The Wavefront SDK will use Direct Ingestion authenticated using CSP client credentials.")
		return auth.NewCSPServerToServerService(a.BaseURL, a.ClientID, a.ClientSecret, a.OrgID)
	case auth.CSPAPIToken:
		log.Println("The Wavefront SDK will use Direct Ingestion authenticated using CSP API Token.")
		return auth.NewCSPTokenService(a.BaseURL, a.Token)
	}

	log.Println("The Wavefront SDK will communicate with a Wavefront Proxy.")
	return auth.NewNoopTokenService()
}
