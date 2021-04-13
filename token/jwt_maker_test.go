package token

import (
	"testing"
	"time"

	"github.com/Jiaget/simplebank/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateToken(username, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VarifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomOwner()
	duration := -time.Minute
	token, err := maker.CreateToken(username, duration)

	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VarifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTMaker(t *testing.T) {
	payload, err := NewPayload(util.RandomOwner(), time.Minute)
	require.NoError(t, err)
	jwttoken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwttoken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err = maker.VarifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload)
	require.EqualError(t, err, ErrInvalidToken.Error())

}
