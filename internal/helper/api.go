package helper

import (
	"dungtl2003/chat-app-auth-service/internal/config"
	"dungtl2003/chat-app-auth-service/internal/jwthandler"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type ParsedToken struct {
	UserId         int64
	SessionVersion int64
	SessionId      int64
}

type CookieCfg struct {
	Domain   string
	Secure   bool
	Path     string
	SameSite http.SameSite
}

// ParseToken is a function that parses the JWT token and returns the parsed token.
// It extracts the user ID, session version, and session ID from the token claims.
// If any of these values cannot be extracted, it returns an error.
func ParseToken(tok *jwt.Token) (*ParsedToken, error) {
	claims := tok.Claims.(*jwthandler.UserJWTClaim)
	sessionVersion, err := claims.GetSessVersion()
	if err != nil {
		return nil, fmt.Errorf("GetSessVersion(): %v", err)
	}
	userIdStr, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("GetSubject(): %v", err)
	}
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("strconv.ParseInt(): %v", err)
	}
	sessId, err := claims.GetSessId()
	if err != nil {
		return nil, fmt.Errorf("GetSessId(): %v", err)
	}

	return &ParsedToken{
		UserId:         userId,
		SessionVersion: sessionVersion,
		SessionId:      sessId,
	}, nil
}

func ClearCookie(c *gin.Context, cookieName string, domainName string) {
	c.SetCookie(cookieName, "", -1, "/", domainName, false, false)
}

func SetCookie(c *gin.Context, cookieName string, cookieValue string, domainName string, durationMs int64) {
	c.SetCookie(cookieName, cookieValue, int(durationMs/1000), "/", domainName, false, false)
}

func GetCookieCfg(domain string, env config.Env) CookieCfg {
	cfg := CookieCfg{
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	if env == config.Prod {
		cfg.Domain = domain
		cfg.Secure = true
	}

	return cfg
}

func SetCookieCfg(c *gin.Context, name, value string, durationMs int64, cfg CookieCfg) {
	c.SetCookie(name, value, int(durationMs/1000), cfg.Path, cfg.Domain, cfg.Secure, false)
}

func SetCookieOverride(c *gin.Context, name, value string, domain string, durationMs int64) {
	header := c.Writer.Header()
	prev := header["Set-Cookie"]
	prefix := name + "=" // cookie lines all start "name="
	filtered := Filter(prev, func(h string) bool {
		return !strings.HasPrefix(h, prefix)
	})
	header["Set-Cookie"] = filtered

	// now add the new one
	SetCookie(c, name, value, domain, durationMs)
}

func SetCookieCfgOverride(c *gin.Context, name, value string, durationMs int64, cfg CookieCfg) {
	header := c.Writer.Header()
	prev := header["Set-Cookie"]
	prefix := name + "=" // cookie lines all start "name="
	filtered := Filter(prev, func(h string) bool {
		return !strings.HasPrefix(h, prefix)
	})
	header["Set-Cookie"] = filtered

	// now add the new one
	SetCookieCfg(c, name, value, durationMs, cfg)
}

// IsNativeClient determines whether the request is from a native mobile client
// or a web browser client based on request headers.
func IsNativeClient(c *gin.Context) bool {
	// Preferred: app sends this header
	if v := c.GetHeader("X-Client-Platform"); strings.EqualFold(v, "mobile") {
		return true
	}
	if v := c.GetHeader("X-Auth-Mode"); strings.EqualFold(v, "bearer") {
		return true
	}
	// Heuristic: native requests usually have no Origin (or "null")
	// origin := c.GetHeader("Origin")
	// return origin == "" || strings.EqualFold(origin, "null")

	return false
}
