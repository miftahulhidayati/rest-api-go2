package main

import (
	"net/http"
	"time"

	"github.com/miftahulhidayati/rest-api-go2/controllers"
	"github.com/miftahulhidayati/rest-api-go2/database"

	"github.com/gin-gonic/gin"
	docs "github.com/miftahulhidayati/rest-api-go2/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type DBConn struct{
	DB *gorm.DB
}

type Item struct {
	gorm.Model
	// ItemID      int    `json:"item_id"`
	ItemCode    string 
	Description string 
	Quantity    int    
	OrderID     int    
}
type Order struct {
	gorm.Model
	// OrderID      int       `json:"order_id"`
	CustomerName string    
	OrderedAt    time.Time 
	Items        []Item    
}

func (conn *DBConn) CreateOrder(c *gin.Context) {
	var order models.Order
	var result gin.H

	err := c.ShouldBindJSON(&order)
	if err != nil{
		result = gin.H{
			"result": "insert failed",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}
	tx := conn.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
		tx.Rollback()
		}
	}()

	err = conn.DB.Create(&order).Error
	if err != nil{
		tx.Rollback()
    	result = gin.H{
			"result": "insert failed",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = gin.H{
		"result": order,
	}
	c.JSON(http.StatusOK, result)
}
// @Summary Get all orders summery
// @Description Get all orders description
// @Accept json
// @Produce json
// @Param orderId body Order true "Create Order"
// @Success 200 {object} Order
// @Router /orders [post]
func (conn *DBConn) GetOrder(c *gin.Context) {
	var (
		order models.Order
		result gin.H
	)
	id := c.Param("id")
	err := conn.DB.Where("id = ?", id).Preload("Items").First(&order).Error
	if err != nil {
		result = gin.H{
			"result": err.Error(),
			"count":  0,
		}
	} else {
		result = gin.H{
			"result": order,
			"count":  1,
		}
	}

	c.JSON(http.StatusOK, result)
}

func (conn *DBConn) GetOrders(c *gin.Context) {
	var (
		orders []models.Order
		result  gin.H
	)

	err := conn.DB.Preload("Items").Find(&orders).Error
	if err != nil {
		result = gin.H{
			"result": err.Error(),
			"count":  0,
		}
		c.JSON(http.StatusBadRequest, result)
		return
	} 		
	if len(orders) <= 0 {
		result = gin.H{
			"result": nil,
			"count":  0,
		}
	} else {
		result = gin.H{
			"result": orders,
			"count":  len(orders),
		}
	}

	c.JSON(http.StatusOK, result)
}

func (conn *DBConn) UpdateOrder(c *gin.Context) {
	
	id := c.Query("id")
	var order models.Order
	var result gin.H
	err := conn.DB.First(&order, id).Error
	if err != nil {
		result = gin.H{
			"result": "data not found",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}
	err = c.ShouldBindJSON(&order)
	if err != nil{
		result = gin.H{
			"result": "update failed",
		}
		
	}
	tx := conn.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
		tx.Rollback()
		}
	}()

	// err = conn.DB.Create(&order).Error
	err = conn.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&order).Error
	if err != nil{
		tx.Rollback()
    	result = gin.H{
			"result": "update failed",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}

	result = gin.H{
		"result": order,
	}
	c.JSON(http.StatusOK, result)
}

func (conn *DBConn) DeleteOrder(c *gin.Context) {
	var (
		order models.Order
		result gin.H
	)
	id := c.Param("id")
	err := conn.DB.Where("id = ?", id).Preload("Items").First(&order).Error
	if err != nil {
		result = gin.H{
			"result": "data not found",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}
	tx := conn.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
		tx.Rollback()
		}
	}()

	err = conn.DB.Delete(&order, id).Error
	if err != nil{
	
		tx.Rollback()
		result = gin.H{
			"result": "delete failed",
		}
		c.JSON(http.StatusBadRequest, result)
		return
	}
	err = conn.DB.Where("order_id = ?", id).Delete(&order.Items).Error
	if err != nil {
		tx.Rollback()
		result = gin.H{
			"result": "delete failed",
		}
	} else {
		result = gin.H{
			"result": "Data deleted successfully",
		}
	}

	c.JSON(http.StatusOK, result)
}


// @title API Order
// @version 1.0
// @description This is a sample API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
func main() {
	router := gin.Default()

	db := database.InitMysqlDB()

	DBConn := &controllers.DBConn{DB: db}
	//Read All
	router.GET("/orders", DBConn.GetOrders)
	//Read One
	router.GET("/orders/:id", DBConn.GetOrder)
	//Create
	router.POST("/orders", DBConn.CreateOrder)
	//Update
	router.PUT("/orders/:id", DBConn.UpdateOrder)
	//Delete
	router.DELETE("/orders/:id", DBConn.DeleteOrder)
	docs.SwaggerInfo.BasePath = "/api/v1"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	router.Run(":8080")
}