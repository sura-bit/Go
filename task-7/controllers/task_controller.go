package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"task_manager/data"
	"task_manager/middleware"
	"task_manager/models"
)

type Controller struct {
	TaskSvc *data.TaskService
	UserSvc *data.UserService
}

func NewController(ts *data.TaskService, us *data.UserService) *Controller {
	return &Controller{TaskSvc: ts, UserSvc: us}
}

func (ctr *Controller) RegisterRoutes(r *gin.Engine) {
	// Public
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/register", ctr.Register)
	r.POST("/login", ctr.Login)

	// Authenticated routes
	auth := r.Group("/", middleware.AuthRequired())

	// Tasks: GET allowed to all authenticated users
	auth.GET("/tasks", ctr.ListTasks)
	auth.GET("/tasks/:id", ctr.GetTask)

	// Admin-only task mutations
	admin := auth.Group("/", middleware.AdminOnly())
	admin.POST("/tasks", ctr.CreateTask)
	admin.PUT("/tasks/:id", ctr.UpdateTask)
	admin.DELETE("/tasks/:id", ctr.DeleteTask)

	// Admin-only user management
	admin.POST("/promote", ctr.PromoteUser)
}

/* --------------------- Auth handlers --------------------- */

func (ctr *Controller) Register(c *gin.Context) {
	var dto models.RegisterDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	u, err := ctr.UserSvc.Register(c.Request.Context(), dto.Username, dto.Password)
	if err != nil {
		switch err {
		case data.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	out := models.UserOut{
		ID:       toHex(u.ID),
		Username: u.Username,
		Role:     u.Role,
	}
	c.JSON(http.StatusCreated, gin.H{"data": out})
}

func (ctr *Controller) Login(c *gin.Context) {
	var dto models.LoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	u, err := ctr.UserSvc.Authenticate(c.Request.Context(), dto.Username, dto.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}
	token, err := makeJWT(toHex(u.ID), u.Username, string(u.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"user": models.UserOut{
			ID:       toHex(u.ID),
			Username: u.Username,
			Role:     u.Role,
		},
	})
}

func (ctr *Controller) PromoteUser(c *gin.Context) {
	var dto models.PromoteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	u, err := ctr.UserSvc.PromoteToAdmin(c.Request.Context(), dto.UserID)
	if err != nil {
		switch err {
		case data.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": models.UserOut{ID: toHex(u.ID), Username: u.Username, Role: u.Role}})
}

func makeJWT(uid, username, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Fail fast—set JWT_SECRET in env for production; for dev you may set a placeholder
		secret = "dev-secret-change-me"
	}

	claims := jwt.MapClaims{
		"uid":      uid,
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Add(-5 * time.Second).Unix(),
	}

	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tkn.SignedString([]byte(secret))
}

func toHex(id interface{}) string {
	if oid, ok := id.(interface{ Hex() string }); ok {
		return oid.Hex()
	}
	return ""
}

/* --------------------- Task handlers --------------------- */

func (ctr *Controller) ListTasks(c *gin.Context) {
	tasks, err := ctr.TaskSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

func (ctr *Controller) GetTask(c *gin.Context) {
	id := c.Param("id")
	t, err := ctr.TaskSvc.Get(c.Request.Context(), id)
	if err != nil {
		switch err {
		case data.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (ctr *Controller) CreateTask(c *gin.Context) {
	var dto models.CreateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	t, err := ctr.TaskSvc.Create(c.Request.Context(), dto)
	if err != nil {
		switch err {
		case data.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status (pending|in_progress|done)"})
		case data.ErrInvalidDate:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date (RFC3339)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": t})
}

func (ctr *Controller) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var dto models.UpdateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	t, err := ctr.TaskSvc.Update(c.Request.Context(), id, dto)
	if err != nil {
		switch err {
		case data.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		case data.ErrInvalidStatus:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status (pending|in_progress|done)"})
		case data.ErrInvalidDate:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date (RFC3339)"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": t})
}

func (ctr *Controller) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := ctr.TaskSvc.Delete(c.Request.Context(), id); err != nil {
		switch err {
		case data.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}
