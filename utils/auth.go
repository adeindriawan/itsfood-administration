package utils

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adeindriawan/itsfood-administration/services"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func getTokenDuration() (time.Duration, time.Duration) {
	atd, _ := strconv.Atoi(os.Getenv("ACCESS_TOKEN_DURATION"))
	rtd, _ := strconv.Atoi(os.Getenv("REFRESH_TOKEN_DURATION"))
	accessTokenDuration := time.Duration(atd) * time.Minute
	refreshTokenDuration := time.Duration(rtd) * time.Minute
	return accessTokenDuration, refreshTokenDuration
}

func CreateToken(userId uint64) (*TokenDetails, error) {
	accessTokenDuration, refreshTokenDuration := getTokenDuration()

	atExpires := time.Now().Add(accessTokenDuration).UnixMilli()
	td := &TokenDetails{}
	td.AtExpires = atExpires
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(refreshTokenDuration).UnixMilli()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	// creating access token
	os.Setenv("ACCESS_SECRET", "loremipsum")
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userId
	atClaims["exp"] = atExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	// creating refresh token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func CreateAuth(userId uint64, td *TokenDetails) error {
	accessTokenDuration, refreshTokenDuration := getTokenDuration()

	errAccess := services.GetRedis().Set(td.AccessUuid, strconv.Itoa(int(userId)), accessTokenDuration).Err()
	if errAccess != nil {
		return errAccess
	}

	errRefresh := services.GetRedis().Set(td.RefreshUuid, strconv.Itoa(int(userId)), refreshTokenDuration).Err()
	if errRefresh != nil {
		return errRefresh
	}

	return nil
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}

		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}

	return nil, err
}

func FetchAuth(authD *AccessDetails) (uint64, error) {
	userid, err := services.GetRedis().Get(authD.AccessUuid).Result()
	if err != nil {
		return 0, err
	}
	userID, _ := strconv.ParseUint(userid, 10, 64)
	return userID, nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func AuthCheck(c *gin.Context) (uint64, error) {
	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		return 0, err
	}
	userId, err := FetchAuth(tokenAuth)
	if err != nil {
		return 0, err
	}

	return userId, nil
}
