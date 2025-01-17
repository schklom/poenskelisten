package controllers

import (
	"aunefyren/poenskelisten/database"
	"aunefyren/poenskelisten/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterUser(context *gin.Context) {
	var user models.User
	var usercreationrequest models.UserCreationRequest
	if err := context.ShouldBindJSON(&usercreationrequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	if usercreationrequest.Password != usercreationrequest.PasswordRepeat {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Passwords must match."})
		context.Abort()
		return
	}

	user.Email = usercreationrequest.Email
	user.Password = usercreationrequest.Password
	user.FirstName = usercreationrequest.FirstName
	user.LastName = usercreationrequest.LastName

	// Hash the selected password
	if err := user.HashPassword(user.Password); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	// Verify unsued invite code exists
	unique_invitecode, err := database.VerifyUnusedUserInviteCode(usercreationrequest.InviteCode)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	} else if !unique_invitecode {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invitiation code is not valid."})
		context.Abort()
		return
	}

	// Verify e-mail is not in use
	unique_email, err := database.VerifyUniqueUserEmail(user.Email)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	} else if !unique_email {
		context.JSON(http.StatusBadRequest, gin.H{"error": "E-mail is already in use."})
		context.Abort()
		return
	}

	// Create user in DB
	record := database.Instance.Create(&user)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		context.Abort()
		return
	}

	// Set code to used
	err = database.SetUsedUserInviteCode(usercreationrequest.InviteCode)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	// Return response
	context.JSON(http.StatusCreated, gin.H{"message": "User created!"})
}

func GetUser(context *gin.Context) {

	// Create user request
	var user = context.Param("user_id")

	// Parse group id
	user_id_int, err := strconv.Atoi(user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	user_object, err := database.GetUserInformation(user_id_int)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	// Reply
	context.JSON(http.StatusOK, gin.H{"user": user_object, "message": "User retrieved."})
}

func GetUsers(context *gin.Context) {

	// Create user request
	var user_struct []models.User

	userrecord := database.Instance.Where("`users`.enabled = ?", 1).Find(&user_struct)
	if userrecord.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": userrecord.Error})
		context.Abort()
		return
	}

	for index, _ := range user_struct {
		// Redact user information
		user_struct[index].Email = "REDACTED"
		user_struct[index].Password = "REDACTED"
	}

	// Reply
	context.JSON(http.StatusOK, gin.H{"users": user_struct, "message": "Users retrieved."})
}
