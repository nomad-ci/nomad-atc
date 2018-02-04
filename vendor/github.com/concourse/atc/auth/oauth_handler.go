package auth

import (
	"crypto/rsa"
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc/auth/provider"
	"github.com/concourse/atc/auth/routes"
	"github.com/concourse/atc/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/tedsuo/rata"
)

var SigningMethod = jwt.SigningMethodRS256

//go:generate counterfeiter . ProviderFactory

type ProviderFactory interface {
	GetProvider(db.Team, string) (provider.Provider, bool, error)
}

func NewOAuthHandler(
	logger lager.Logger,
	providerFactory ProviderFactory,
	teamFactory db.TeamFactory,
	signingKey *rsa.PrivateKey,
	expire time.Duration,
	isTLSEnabled bool,
) (http.Handler, error) {
	return rata.NewRouter(
		routes.OAuthRoutes,
		map[string]http.Handler{
			routes.OAuthBegin: NewOAuthBeginHandler(
				logger.Session("oauth-begin"),
				providerFactory,
				signingKey,
				teamFactory,
				expire,
				isTLSEnabled,
			),
			routes.OAuthCallback: NewOAuthCallbackHandler(
				logger.Session("oauth-callback"),
				providerFactory,
				signingKey,
				teamFactory,
				expire,
				isTLSEnabled,
				oauthV2StateValidator{},
			),
			routes.LogOut: NewLogOutHandler(
				logger.Session("logout"),
			),
		},
	)
}

func NewOAuthV1Handler(
	logger lager.Logger,
	providerFactory ProviderFactory,
	teamFactory db.TeamFactory,
	signingKey *rsa.PrivateKey,
	expire time.Duration,
	isTLSEnabled bool,
) (http.Handler, error) {
	return rata.NewRouter(
		routes.OAuthV1Routes,
		map[string]http.Handler{
			routes.OAuthV1Begin: NewOAuthBeginHandler(
				logger.Session("oauth-v1-begin"),
				providerFactory,
				signingKey,
				teamFactory,
				expire,
				isTLSEnabled,
			),
			routes.OAuthV1Callback: NewOAuthCallbackHandler(
				logger.Session("oauth-v1-callback"),
				providerFactory,
				signingKey,
				teamFactory,
				expire,
				isTLSEnabled,
				oauthV1StateValidator{},
			),
		},
	)
}
