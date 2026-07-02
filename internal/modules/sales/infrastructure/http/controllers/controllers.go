package controllers

import (
	"net/http"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/modules/sales/use_cases"
	"order-manager/internal/shared/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createOrderRequest struct {
	CustomerName string `json:"customer_name" binding:"required"`
}

type addItemRequest struct {
	ItemsID map[uuid.UUID]float64 `json:"items_id"`
}

type removeItemRequest struct {
	ItemsID []uuid.UUID `json:"items_id"`
}

type OrderController struct {
	orderUC *use_cases.OrderUseCase
}

func NewOrderController(uc *use_cases.OrderUseCase) *OrderController {
	return &OrderController{
		orderUC: uc,
	}
}

func (ctrl *OrderController) RegisterRoutes(router *gin.RouterGroup) {
	orders := router.Group("/orders")
	{
		orders.POST("/", ctrl.CreateOrder)
		orders.GET("/", ctrl.ListOrdersByStatus)

		orders.POST("/:id/pay", ctrl.PayOrder)
		orders.POST("/:id/cancel", ctrl.CancelOrder)

		orders.POST("/:id/items/add", ctrl.AddOrderItems)
		orders.POST("/:id/items/remove", ctrl.RemoveOrderItems)
	}
}

func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var req createOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload or missing customer_name"})
		return
	}

	orderID, err := ctrl.orderUC.CreateOrder(c.Request.Context(), req.CustomerName)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Order created successfully",
		"order_id": orderID,
	})
}

func (ctrl *OrderController) ListOrdersByStatus(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("pageSize", "10")
	statusStr := c.DefaultQuery("status", "PENDING")
	sort := c.DefaultQuery("sort", "created_at")
	direction := c.DefaultQuery("direction", "desc")
	search := c.Query("search")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(sizeStr)

	status, err := value_objects.NewOrderStatus(statusStr)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	requestPagination := utils.NewPagination(page, pageSize, search, sort, direction)

	orders, err := ctrl.orderUC.ListOrdersByStatus(c.Request.Context(), status, requestPagination)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (ctrl *OrderController) PayOrder(c *gin.Context) {
	orderIDStr := c.Param("id")

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := ctrl.orderUC.PayOrder(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order paid successfully",
	})
}

func (ctrl *OrderController) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("id")

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := ctrl.orderUC.CancelOrder(c.Request.Context(), orderID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

func (ctrl *OrderController) AddOrderItems(c *gin.Context) {
	orderIDStr := c.Param("id")

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req addItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := ctrl.orderUC.AddOrderItems(c.Request.Context(), orderID, req.ItemsID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order items added successfully",
	})
}

func (ctrl *OrderController) RemoveOrderItems(c *gin.Context) {
	orderIDStr := c.Param("id")

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req removeItemRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := ctrl.orderUC.RemoveOrderItems(c.Request.Context(), orderID, req.ItemsID); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order items removed successfully",
	})
}
