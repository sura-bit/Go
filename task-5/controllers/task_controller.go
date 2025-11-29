package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"task_manager/data"
	"task_manager/models"
)

type TaskController struct {
	Service *data.InMemoryTaskService
}

func NewTaskController(s *data.InMemoryTaskService) *TaskController {
	return &TaskController{Service: s}
}

func (c *TaskController) Register(r *gin.RouterGroup) {
	r.GET("/tasks", c.List)
	r.GET("/tasks/:id", c.Get)
	r.POST("/tasks", c.Create)
	r.PUT("/tasks/:id", c.Update)
	r.DELETE("/tasks/:id", c.Delete)
}

func (c *TaskController) List(ctx *gin.Context) {
	tasks := c.Service.List()
	ctx.JSON(http.StatusOK, gin.H{"data": tasks})
}

func (c *TaskController) Get(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMsg("invalid id"))
		return
	}
	task, err := c.Service.Get(id)
	if err != nil {
		if err == data.ErrNotFound {
			ctx.JSON(http.StatusNotFound, errorMsg("task not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorMsg("internal error"))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

func (c *TaskController) Create(ctx *gin.Context) {
	var dto models.CreateTaskDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, errorMsg("invalid request body"))
		return
	}
	task, err := c.Service.Create(dto)
	if err != nil {
		switch err {
		case data.ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, errorMsg("invalid status (use: pending | in_progress | done)"))
			return
		case data.ErrInvalidDate:
			ctx.JSON(http.StatusBadRequest, errorMsg("invalid due_date (use RFC3339, e.g. 2025-12-31T23:59:59Z)"))
			return
		default:
			ctx.JSON(http.StatusInternalServerError, errorMsg("internal error"))
			return
		}
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": task})
}

func (c *TaskController) Update(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMsg("invalid id"))
		return
	}
	var dto models.UpdateTaskDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, errorMsg("invalid request body"))
		return
	}
	task, err := c.Service.Update(id, dto)
	if err != nil {
		switch err {
		case data.ErrNotFound:
			ctx.JSON(http.StatusNotFound, errorMsg("task not found"))
		case data.ErrInvalidStatus:
			ctx.JSON(http.StatusBadRequest, errorMsg("invalid status (use: pending | in_progress | done)"))
		case data.ErrInvalidDate:
			ctx.JSON(http.StatusBadRequest, errorMsg("invalid due_date (use RFC3339, e.g. 2025-12-31T23:59:59Z)"))
		default:
			ctx.JSON(http.StatusInternalServerError, errorMsg("internal error"))
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": task})
}

func (c *TaskController) Delete(ctx *gin.Context) {
	id, err := parseID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMsg("invalid id"))
		return
	}
	if err := c.Service.Delete(id); err != nil {
		if err == data.ErrNotFound {
			ctx.JSON(http.StatusNotFound, errorMsg("task not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorMsg("internal error"))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func errorMsg(m string) gin.H { return gin.H{"error": m} }
