package identity

import (
	"context"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jasonsoft/napnap"
)

const (
	claimKey = "identity_claimKey"
)

func FromContext(ctx context.Context) (jwt.MapClaims, bool) {
	val, ok := ctx.Value(claimKey).(jwt.MapClaims)
	if !ok {
		return nil, false
	}
	return val, true
}

func NewContext(ctx context.Context, claim jwt.MapClaims) context.Context {
	return context.WithValue(ctx, claimKey, claim)
}

type JWTMiddleware struct {
	config        JwtOptions
	secretKeyByte []byte
}

type JwtOptions struct {
	SecretKey     string
	DurationInMin int
}

func NewJWTMiddleware(opts JwtOptions) *JWTMiddleware {
	return &JWTMiddleware{
		config:        opts,
		secretKeyByte: []byte(opts.SecretKey),
	}
}

func (jwtMW *JWTMiddleware) Invoke(c *napnap.Context, next napnap.HandlerFunc) {

	tokenString := c.RequestHeader("Authorization")

	if len(tokenString) == 0 {
		c.SetStatus(401)
		return
	}

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return jwtMW.secretKeyByte, nil
	})
	if err != nil {
		c.SetStatus(401)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ctx := c.StdContext()
		ctx = NewContext(ctx, claims)
		c.SetStdContext(ctx)
		next(c)
	} else {
		c.SetStatus(401)
		return
	}

}
