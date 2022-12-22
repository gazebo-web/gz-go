package gz

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// AccessToken is a single personal access token for a user.
type AccessToken struct {
	// ID is the primary key for access tokens.
	ID uint `gorm:"primary_key" json:"-"`

	// CreatedAt is the time when the access token was created
	CreatedAt time.Time `json:"created_at"`

	// UserID is the user that owns this token.
	UserID uint `json:"-"`

	// Name is a string given to a token by the user. The name does not have to be unique.
	Name string `json:"name"`

	// Prefix is the first set of characters in the token. The prefix is used to identify the user
	// and help a user keep track of their tokens. We use 'latin1_general_cs` collation to enforce
	// case-senstive queries.
	Prefix string `sql:"type:VARCHAR(64) CHARACTER SET latin1 COLLATE latin1_general_cs" json:"prefix"`

	// Key is the second set of characters in the token, following the Prefix. The key is used to
	// authenticate the user. We use 'latin1_general_cs` collation to enforce case-senstive queries. The key is omitted from json to prevent it from being transmitted over the wire.
	Key string `sql:"type:VARCHAR(512) CHARACTER SET latin1 COLLATE latin1_general_cs" json:"-"`

	// Last used time.
	LastUsed *time.Time `json:"last_used"`

	// For future use, when we add in the ability to expire tokens.
	Expires *time.Time `json:"expires"`
}

// AccessTokens is an array of AccessToken
type AccessTokens []AccessToken

// AccessTokenCreateRequest contains information required to create a new access token.
type AccessTokenCreateRequest struct {
	Name string `json:"name" validate:"required,min=3,alphanum"`
}

// AccessTokenCreateResponse contains information about a newly created access token.
type AccessTokenCreateResponse struct {
	// Name is a string given to a token by the user.
	Name string `json:"name"`

	// Prefix is the first set of characters in the token. The prefix is used to identify the user
	// and help a user keep track of their tokens.
	Prefix string `json:"prefix"`

	// Key is the second set of characters in the token, following the Prefix. The key is used to
	// authenticate the user.
	Key string `json:"key"`
}

// ValidateAccessToken checks a token string against tokens that exist in the
// provided database. If the access token is validated, then it is returned
// as the first return value.
func ValidateAccessToken(token string, tx *gorm.DB) (*AccessToken, *ErrMsg) {
	// Split the token into the prefix and key parts.
	parts := strings.Split(token, ".")

	// Make sure that there are exactly two parts.
	if len(parts) != 2 {
		return nil, NewErrorMessage(ErrorUnauthorized)
	}

	// Get all the access tokens with the specified prefix. There should only
	// be one, which we check for in the following `if` condition.
	var accessTokens AccessTokens
	if err := tx.Where("prefix = ?", parts[0]).Find(&accessTokens).Error; err != nil {
		return nil, NewErrorMessage(ErrorUnauthorized)
	}

	// This should never happen, but it's better safe than sorry.
	// If multiple prefixes are found, then we will assume the worse and
	// deny authorization.
	if len(accessTokens) != 1 {
		return nil, NewErrorMessage(ErrorUnauthorized)
	}

	// At this point, we have a single user which can be authenticated by
	// comparing the provided key with the salted key in the database.
	if err := bcrypt.CompareHashAndPassword([]byte(accessTokens[0].Key), []byte(parts[1])); err != nil {
		return nil, NewErrorMessage(ErrorUnauthorized)
	}

	return &accessTokens[0], nil
}

// Create instantiates a new unique random access token. The first return
// value is the full access token, which can be passed along to a user.
// Be careful with the full access token since it is a full-access key.
// The second return value is a salted token, which is suitable for storage
// in a database. The third return value is an error, or nil.
func (createReq *AccessTokenCreateRequest) Create(tx *gorm.DB) (*AccessTokenCreateResponse, *AccessToken, *ErrMsg) {

	// Create the key
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, nil, NewErrorMessage(ErrorUnexpected)
	}
	key := base64.URLEncoding.EncodeToString(b)

	var accessTokens AccessTokens
	var prefixToken string

	// An 8byte prefix would allow for 1.7e+19 tokens.
	// Use a loop to make sure that we generate a unique prefix.
	// Limit the loop iterations, just in case. We should always be able to
	// generate a unique prefix.
	for i := 0; i < 100; i++ {
		prefix := make([]byte, 8)
		_, err := rand.Read(prefix)
		if err != nil {
			return nil, nil, NewErrorMessage(ErrorUnexpected)
		}
		prefixToken = base64.URLEncoding.EncodeToString(prefix)
		tx.Where("prefix = ?", prefixToken).Find(&accessTokens)
		if len(accessTokens) <= 0 {
			break
		}
		prefixToken = ""
	}

	// Return an error if we were not able to generate a unique prefix.
	// This should never happen.
	if prefixToken == "" {
		return nil, nil, NewErrorMessage(ErrorUnexpected)
	}

	// Return the name, prefix, and key.
	var newToken = AccessTokenCreateResponse{
		Name:   createReq.Name,
		Prefix: prefixToken,
		Key:    key,
	}

	// Generate the salted key, and return an error if the key could not be
	// salted.
	saltedKey, saltErr := bcrypt.GenerateFromPassword([]byte(key), 8)
	if saltErr != nil {
		return nil, nil, NewErrorMessage(ErrorUnexpected)
	}

	// Create the salted password, which is stored in the database.
	var saltedToken = AccessToken{
		Name:   createReq.Name,
		Prefix: prefixToken,
		Key:    string(saltedKey),
	}

	return &newToken, &saltedToken, nil
}
