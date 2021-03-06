package controller

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/blend/go-sdk/oauth"
	"github.com/blend/go-sdk/uuid"
	"github.com/blend/go-sdk/web"

	"github.com/wcharczuk/kana-server/pkg/config"
	"github.com/wcharczuk/kana-server/pkg/model"
	"github.com/wcharczuk/kana-server/pkg/types"
)

const (
	// SessionKeyUser is the session state key where the user reference is held.
	SessionKeyUser = "__user__"
)

type Auth struct {
	Config config.Config
	OAuth  *oauth.Manager
	Model  *model.Manager
}

func (a Auth) Register(app *web.App) {
	jwtm := web.NewJWTManager(a.mustSecret())
	jwtm.Apply(&app.Auth)

	app.Auth.FetchHandler = a.fetchHandler(jwtm.FetchHandler)
	app.Auth.LoginRedirectHandler = a.loginRedirect

	app.GET("/login", a.login, web.SessionAware, web.ViewProviderAsDefault)
	app.GET("/logout", a.logout, web.SessionAware, web.ViewProviderAsDefault)
	app.GET("/oauth/google", a.oauthGoogle, web.SessionAware, web.ViewProviderAsDefault)
}

// GET /login
func (a Auth) login(r *web.Ctx) web.Result {
	if r.Session != nil {
		return a.authedRedirect()
	}
	oauthURL, err := a.OAuth.OAuthURL(r.Request, oauth.OptStateRedirectURI(r.Request.RequestURI))
	if err != nil {
		return r.Views.InternalError(err)
	}
	return web.RedirectWithMethod("GET", oauthURL)
}

// GET /oauth/google
func (a Auth) oauthGoogle(r *web.Ctx) web.Result {
	if r.Session != nil {
		return a.authedRedirect()
	}
	result, err := a.OAuth.Finish(r.Request)
	if err != nil {
		return r.Views.NotAuthorized()
	}

	user, found, err := a.Model.GetUserByEmail(r.Context(), result.Profile.Email)
	if err != nil {
		return r.Views.InternalError(err)
	}
	types.ApplyProfileToUser(&user, result.Profile)
	if !found {
		user.ID = uuid.V4()
		user.CreatedUTC = time.Now().UTC()
	}
	user.LastSeenUTC = time.Now().UTC()
	if err := a.Model.Invoke(r.Context()).Upsert(&user); err != nil {
		return r.Views.InternalError(err)
	}
	sess, err := r.Auth.Login(user.ID.String(), r)
	if err != nil {
		return r.Views.InternalError(err)
	}
	r.Session = sess
	if len(result.State.RedirectURI) > 0 {
		return web.RedirectWithMethodf(http.MethodGet, result.State.RedirectURI)
	}
	return a.authedRedirect()
}

// logout logs the user out.
func (a Auth) logout(r *web.Ctx) web.Result {
	if r.Session == nil {
		return r.Views.NotAuthorized()
	}
	if err := r.Auth.Logout(r); err != nil {
		return r.Views.InternalError(err)
	}
	return web.RedirectWithMethod("GET", "/")
}

//
// helpers
//

func (a Auth) mustSecret() []byte {
	secret, err := a.Config.GetSecret()
	if err != nil {
		panic(err)
	}
	return secret
}

func (a Auth) authedRedirect() web.Result {
	return web.RedirectWithMethod(http.MethodGet, "/checks")
}

func (a Auth) fetchHandler(jwtHandler web.AuthManagerFetchSessionHandler) web.AuthManagerFetchSessionHandler {
	return func(ctx context.Context, sessionValue string) (*web.Session, error) {
		session, err := jwtHandler(ctx, sessionValue)
		if err != nil {
			return nil, err
		}
		var user types.User
		found, err := a.Model.Invoke(ctx).Get(&user, session.UserID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, nil
		}
		session.State[SessionKeyUser] = &user
		return session, nil
	}
}

func (a Auth) loginRedirect(r *web.Ctx) *url.URL {
	from := r.Request.URL.Path
	oauthURL, err := a.OAuth.OAuthURL(r.Request, oauth.OptStateRedirectURI(from))
	if err != nil {
		return &url.URL{RawPath: "/login?error=invalid_oauth_url"}
	}
	parsed, err := url.Parse(oauthURL)
	if err != nil {
		return &url.URL{RawPath: "/login?error=invalid_oauth_url"}
	}
	return parsed
}
