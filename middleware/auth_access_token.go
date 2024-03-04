package middleware

import (
	"context"

	"github.com/gazebo-web/auth/pkg/authentication"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
)

// BearerAccessTokenAuthFuncGRPC returns a grpc_auth.AuthFunc that allows to validate
// incoming access tokens found in the Authorization header. These tokens are
// bearer tokens signed by different authentication providers.
// The validator function received as an argument performs the validation for
// every incoming bearer token.
func BearerAccessTokenAuthFuncGRPC(validator authentication.AccessTokenAuthentication) grpc_auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		token, err := grpc_auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}
		err = validator(ctx, token)
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}
}
