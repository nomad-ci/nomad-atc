package uaa

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/concourse/atc"
	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/auth/provider"
	"github.com/concourse/atc/auth/routes"
	"github.com/concourse/atc/auth/verifier"
	"github.com/hashicorp/go-multierror"
	flags "github.com/jessevdk/go-flags"
	"github.com/tedsuo/rata"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const ProviderName = "uaa"
const DisplayName = "UAA"

var Scopes = []string{"cloud_controller.read"}

type UAAProvider struct {
	*oauth2.Config
	verifier.Verifier
	CFCACert string
}

func (p UAAProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) (string, error) {
	return p.Config.AuthCodeURL(state, opts...), nil
}

func (p UAAProvider) Exchange(ctx context.Context, req *http.Request) (*oauth2.Token, error) {
	return p.Config.Exchange(ctx, req.FormValue("code"))
}

func init() {
	provider.Register(ProviderName, UAATeamProvider{})
}

type UAAAuthConfig struct {
	ClientID     string `json:"client_id"     long:"client-id"     description:"Application client ID for enabling UAA OAuth."`
	ClientSecret string `json:"client_secret" long:"client-secret" description:"Application client secret for enabling UAA OAuth."`

	AuthURL  string                `json:"auth_url,omitempty"      long:"auth-url"      description:"UAA AuthURL endpoint."`
	TokenURL string                `json:"token_url,omitempty"     long:"token-url"     description:"UAA TokenURL endpoint."`
	CFSpaces []string              `json:"cf_spaces,omitempty"     long:"cf-space"      description:"Space GUID for a CF space whose developers will have access."`
	CFURL    string                `json:"cf_url,omitempty"        long:"cf-url"        description:"CF API endpoint."`
	CFCACert auth.FileContentsFlag `json:"cf_ca_cert,omitempty"    long:"cf-ca-cert"    description:"Path to CF PEM-encoded CA certificate file."`
}

func (*UAAAuthConfig) AuthMethod(oauthBaseURL string, teamName string) atc.AuthMethod {
	path, err := routes.OAuthRoutes.CreatePathForRoute(
		routes.OAuthBegin,
		rata.Params{"provider": ProviderName},
	)
	if err != nil {
		panic("failed to construct oauth begin handler route: " + err.Error())
	}

	path = path + fmt.Sprintf("?team_name=%s", teamName)

	return atc.AuthMethod{
		Type:        atc.AuthTypeOAuth,
		DisplayName: DisplayName,
		AuthURL:     oauthBaseURL + path,
	}
}

func (auth *UAAAuthConfig) IsConfigured() bool {
	return auth.ClientID != "" ||
		auth.ClientSecret != "" ||
		len(auth.CFSpaces) > 0 ||
		auth.AuthURL != "" ||
		auth.TokenURL != "" ||
		auth.CFURL != ""
}

func (auth *UAAAuthConfig) Validate() error {
	var errs *multierror.Error
	if auth.ClientID == "" || auth.ClientSecret == "" {
		errs = multierror.Append(
			errs,
			errors.New("must specify --uaa-auth-client-id and --uaa-auth-client-secret to use UAA OAuth."),
		)
	}
	if len(auth.CFSpaces) == 0 {
		errs = multierror.Append(
			errs,
			errors.New("must specify --uaa-auth-cf-space to use UAA OAuth."),
		)
	}
	if auth.AuthURL == "" || auth.TokenURL == "" || auth.CFURL == "" {
		errs = multierror.Append(
			errs,
			errors.New("must specify --uaa-auth-auth-url, --uaa-auth-token-url and --uaa-auth-cf-url to use UAA OAuth."),
		)
	}
	return errs.ErrorOrNil()
}

type UAATeamProvider struct{}

func (UAATeamProvider) AddAuthGroup(group *flags.Group) provider.AuthConfig {
	flags := &UAAAuthConfig{}

	uaGroup, err := group.AddGroup("UAA Authentication", "", flags)
	if err != nil {
		panic(err)
	}

	uaGroup.Namespace = "uaa-auth"

	return flags
}

func (UAATeamProvider) UnmarshalConfig(config *json.RawMessage) (provider.AuthConfig, error) {
	flags := &UAAAuthConfig{}
	if config != nil {
		err := json.Unmarshal(*config, &flags)
		if err != nil {
			return nil, err
		}
	}
	return flags, nil
}

func (UAATeamProvider) ProviderConstructor(
	config provider.AuthConfig,
	redirectURL string,
) (provider.Provider, bool) {
	uaaAuth := config.(*UAAAuthConfig)

	endpoint := oauth2.Endpoint{}
	if uaaAuth.AuthURL != "" && uaaAuth.TokenURL != "" {
		endpoint.AuthURL = uaaAuth.AuthURL
		endpoint.TokenURL = uaaAuth.TokenURL
	}

	return UAAProvider{
		Verifier: NewSpaceVerifier(
			uaaAuth.CFSpaces,
			uaaAuth.CFURL,
		),
		Config: &oauth2.Config{
			ClientID:     uaaAuth.ClientID,
			ClientSecret: uaaAuth.ClientSecret,
			Endpoint:     endpoint,
			Scopes:       Scopes,
			RedirectURL:  redirectURL,
		},
		CFCACert: string(uaaAuth.CFCACert),
	}, true
}

func (p UAAProvider) PreTokenClient() (*http.Client, error) {
	transport := &http.Transport{
		Proxy:             http.ProxyFromEnvironment,
		DisableKeepAlives: true,
	}

	if p.CFCACert != "" {
		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM([]byte(p.CFCACert))
		if !ok {
			return nil, errors.New("failed to use cf certificate")
		}

		transport.TLSClientConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	}

	return &http.Client{
		Transport: transport,
	}, nil
}
