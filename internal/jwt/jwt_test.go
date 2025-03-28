package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJWT(t *testing.T) {
	theSecret := "Your thoughts become things!"
	userID := "1"

	token, _ := GenerateJWT(userID, theSecret)

	id, _ := ValidateJWT(token, theSecret)
	assert.Equal(t, id, userID)

	id, err := ValidateJWT("a fake token", theSecret)
	assert.EqualError(t, err, "token contains an invalid number of segments")

	id, err = ValidateJWT(token, "a fake secret")
	assert.EqualError(t, err, "signature is invalid")
}
