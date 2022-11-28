package controllers

import (
	"os"
	"fmt"
	"time"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/adeindriawan/itsfood-administration/models"
	"github.com/adeindriawan/itsfood-administration/services"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/twinj/uuid"
	"github.com/adeindriawan/itsfood-administration/utils"
)

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

func ForgotPassword(c *gin.Context) {
	var payload ForgotPasswordPayload
	var user models.User
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal memproses data yang masuk.",
		})
		return
	}
	query := services.DB.First(&user, "email = ?", payload.Email)
	if query.Error != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": query.Error.Error(),
			"result": nil,
			"description": "Gagal melakukan query pada database.",
		})
		return
	}

	resetToken := uuid.NewV4().String()
	mailTo := user.Email
	mailSubject := "[ITS Food] Lupa Kata Sandi"
	mailBody := resetToken

	resetTokenExpires := time.Now().Add(time.Minute * 15).UnixMilli()
	rtx := time.Unix(resetTokenExpires, 0)
	now := time.Now()
	if err := services.GetRedis().Set(resetToken, mailTo, rtx.Sub(now)).Err(); err != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal menyimpan reset token pada database.",
		})
		return
	}

	_, errorSendingResetPasswordEmail := services.SendMail(mailTo, mailSubject, mailBody)
	if errorSendingResetPasswordEmail != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": nil,
			"result": nil,
			"description": "Gagal mengirim email berisi token ke alamat " + mailTo,
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"errors": nil,
		"result": user,
		"description": "Sukses mengirim email berisi token ke alamat " + mailTo,
	})
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
	Password string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func ResetPassword(c *gin.Context) {
	var payload ResetPasswordPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal memproses data yang masuk.",
		})
		return
	}

	if payload.Password != payload.ConfirmPassword {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": "Gagal memvalidasi data yang masuk",
			"result": nil,
			"description": "Data password tidak sama dengan confirm password yang dikirim.",
		})
		return
	}

	if email, err := services.GetRedis().Get(payload.Token).Result(); err != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Token tidak ditemukan dalam database. Kemungkinan sudah kadaluwarsa.",
		})
		return
	} else if email != payload.Email {
		c.JSON(400, gin.H{
			"status": "failed",
			"errors": "Email yang terkirim tidak sama dengan email yang tersimpan dalam token di Redis.",
			"result": nil,
			"description": "User dengan email ini tidak dapat mereset password.",
		})
		return
	}

	var user models.User
	findUser := services.DB.First(&user, "email = ?", payload.Email)
	if findUser.Error != nil {
		c.JSON(404, gin.H{
			"status": "failed",
			"errors": findUser.Error.Error(),
			"result": nil,
			"description": "Gagal menemukan user dengan email tersebut dalam sistem.",
		})
		return
	}

	hash, errHash := utils.HashPassword(payload.Password)
	if errHash != nil {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": errHash.Error(),
			"result": nil,
			"descripion": "Gagal membuat hash dari password yang diberikan.",
		})
		return
	}
	user.Password = hash
	updatePassword := services.DB.Save(&user)
	if updatePassword.Error != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": updatePassword.Error.Error(),
			"result": nil,
			"description": "Gagal mengubah password dari user ini pada database.",
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "success",
		"errors": nil,
		"result": hash,
		"description": "Sukses mengganti password dari user ini.",
	})
}

type AdminRegisterInput struct {
	Name string				`json:"name" binding:"required"`
	Email string			`json:"email" binding:"required"`
	Password string 	`json:"password" binding:"required"`
	Phone string 			`json:"phone" binding:"required"`
}

type AdminLoginInput struct {
	Email string 		`json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func AdminRegister(c *gin.Context) {
	var register AdminRegisterInput
	if err := c.ShouldBindJSON(&register); err != nil {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal memproses data yang masuk.",
		})
		return
	}

	hashedPassword, errHashingPassword := utils.HashPassword(register.Password)
	if errHashingPassword != nil {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": errHashingPassword.Error(),
			"result": nil,
			"description": "Gagal membuat hash password.",
		})
		return
	}

	user := models.User{Name: register.Name, Email: register.Email, Password: hashedPassword, Phone: register.Phone, Type: "Admin", Status: "Registered", CreatedBy: register.Name, UpdatedAt: time.Time{}}
	if errorCreatingUser := services.DB.Create(&user).Error; errorCreatingUser != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": errorCreatingUser.Error(),
			"result": nil,
			"description": "Gagal menyimpan data user baru dalam database.",
		})
		return
	}

	userId := user.ID
	admin := models.Admin{UserID: userId, Name: register.Name, Email: register.Email, Phone: register.Phone, Status: "Inactive", CreatedBy: register.Name, UpdatedAt: time.Time{}}
	if errorCreatingAdmin := services.DB.Create(&admin).Error; errorCreatingAdmin != nil {
		c.JSON(512, gin.H{
			"status": "failed",
			"errors": errorCreatingAdmin.Error(),
			"result": nil,
			"description": "Gagal menyimpan data admin baru dalam database.",
		})
		return
	}
	
	c.JSON(201, gin.H{
		"status": "success",
		"errors": nil,
		"result": user,
		"description": "Berhasil menambah admin baru.",
	})
}

func AdminLogin(c *gin.Context) {
	var user models.User
	var login AdminLoginInput

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal memproses data yang masuk.",
		})
		return
	}

	if err := services.DB.Where("email = ?", login.Email).First(&user).Error; err != nil {
		c.JSON(404, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal menemukan user dengan email yang dikirimkan.",
		})
		return
	} else if user.Email != login.Email || !utils.CheckPasswordHash(login.Password, user.Password) {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": "Gagal mengautentikasi.",
			"result": nil,
			"description": "Gagal mengautentikasi info login dari data yang dikirimkan.",
		})
		return
	} else {
		if user.Type != "Admin" {
			c.JSON(422, gin.H{
				"status": "failed",
				"errors": "Not admin",
				"result": nil,
				"description": "User yang bersangkutan bukan bertipe Admin.",
			})
			return
		}
		var admin models.Admin
		if err := services.DB.Preload("User").Where("user_id = ?", user.ID).First(&admin).Error; err != nil {
			c.JSON(404, gin.H{
				"status": "failed",
				"errors": err.Error(),
				"result": nil,
				"description": "Gagal menemukan data admin dengan ID user tersebut.",
			})
			return
		}
		ts, err := utils.CreateToken(user.ID)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": err.Error(),
				"result": nil,
				"description": "Tidak dapat membuat token untuk proses autentikasi.",
			})
			return
		}
		saveErr := utils.CreateAuth(user.ID, ts)
		if saveErr != nil {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": saveErr.Error(),
				"result": nil,
				"description": "Gagal membuat autentikasi user.",
			})
			return
		}
		data := map[string]interface{}{
			"token": ts,
			"admin": admin,
		}
		c.JSON(200, gin.H{
			"status": "success",
			"errors": nil,
			"result": data,
			"description": "Berhasil login",
		})
	}
}

func DeleteAuth(givenUuid string) (int64, error) {
	deleted, err := services.GetRedis().Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func Logout(c *gin.Context) {
	au, err := utils.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Tidak dapat mengekstrak token user.",
		})
		return
	}
	
	_, errorFetchingAuth := utils.FetchAuth(au)
	if errorFetchingAuth != nil {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal mengambil user ID.",
		})
		return
	}

	deleted, delErr := DeleteAuth(au.AccessUuid)
	if delErr != nil || deleted == 0 {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": "Tidak ada token user yang terhapus: " + delErr.Error(),
			"result": nil,
			"description": "Error dalam menghapus token user atau tidak ada token yang terhapus.",
		})
		return
	}
	c.JSON(200, gin.H{
		"status": "success",
		"errors": nil,
		"result": nil,
		"description": "Berhasil log out.",
	})
}

func Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(422, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Data yang masuk tidak dapat diproses lebih lanjut.",
		})
		return
	}
	refreshSecret := os.Getenv("REFRESH_SECRET")
	refreshToken := mapToken["refresh_token"]
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected string method: %v", token.Header["alg"])
		}
		return []byte(refreshSecret), nil
	})
	// If there is an error, the token must have expired
	if err != nil {
		c.JSON(403, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Refresh token yang ada sudah kadaluarsa. Silakan login kembali.",
		})
		return
	}
	// Is the token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(500, gin.H{
			"status": "failed",
			"errors": err.Error(),
			"result": nil,
			"description": "Gagal membuat access & refresh token yang baru.",
		})
		return
	}
	// Since the token is valid, get the uuid
	claims, ok := token.Claims.(jwt.MapClaims) // the token should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) // convert the interface to string
		if !ok {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": err.Error(),
				"result": nil,
				"description": "Gagal membuat access & refresh token yang baru.",
			})
			return
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": err.Error(),
				"result": nil,
				"description": "Tidak bisa membuat access & refresh token yang baru karena gagal mengonversi ID user.",
			})
			return
		}
		// Delete the previous refresh token
		deleted, delErr := DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": delErr.Error(),
				"result": nil,
				"description": "Tidak bisa membuat access & refresh token yang baru karena gagal menghapus refresh token yang lama.",
			})
			return
		}
		// Create new pairs of refresh and access token
		ts, createErr := utils.CreateToken(userId)
		if createErr != nil {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": createErr.Error(),
				"result": nil,
				"description": "Gagal membuat access & refresh token yang baru.",
			})
			return
		}
		// Save the token metadata to Redis
		saveErr := utils.CreateAuth(userId, ts)
		if saveErr != nil {
			c.JSON(500, gin.H{
				"status": "failed",
				"errors": saveErr.Error(),
				"result": nil,
				"description": "Gagal menyimpan metadata token ke database.",
			})
			return
		}
		tokens := map[string]interface{}{
			"tokens": ts,
		}
		c.JSON(201, gin.H{
			"status": "success",
			"errors": nil,
			"result": tokens,
			"description": "Berhasil memperbarui access & refresh token.",
		})
	} else {
		c.JSON(403, gin.H{
			"status": "failed",
			"errors": "Refresh expired",
			"result": nil,
			"description": "Token untuk merefresh access token sudah kadaluarsa.",
		})
	}
}
