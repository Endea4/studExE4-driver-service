package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Endea4/studExE4-driver-service/shared/config"
	"github.com/Endea4/studExE4-driver-service/shared/mongo"
	"github.com/Endea4/studExE4-driver-service/internal/models"
	"github.com/Endea4/studExE4-driver-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	config.LoadConfig()
	uri := config.GetEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName := config.GetEnv("DB_NAME", "studexdb")

	client, db := mongo.ConnectDB(uri, dbName)
	defer mongo.Disconnect(client)

	driverRepo := repository.NewDriverRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	debtRepo := repository.NewDebtRepository(db)
	ratingRepo := repository.NewRatingRepository(db)

	userServiceURL := config.GetEnv("USER_SERVICE_URL", "http://localhost:8081")

	r := gin.Default()

	r.POST("/admin/drivers", func(c *gin.Context) {
		var req struct {
			Phone       string `json:"phone" binding:"required"`
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Gender      string `json:"gender"`
			VehicleType string `json:"vehicle_type" binding:"required"`
			PlateNumber string `json:"plate_number" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		existing, _ := driverRepo.GetByPhone(c.Request.Context(), req.Phone)
		if existing != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Driver already exists with this phone"})
			return
		}

		registerBody, _ := json.Marshal(map[string]string{"phone": req.Phone})
		resp, err := http.Post(userServiceURL+"/users/register", "application/json", bytes.NewBuffer(registerBody))
		if err != nil {
			fmt.Printf("Admin: Failed to call user-service register: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call user-service"})
			return
		}
		defer resp.Body.Close()

		isNewUser := resp.StatusCode == http.StatusCreated

		if isNewUser && req.Name != "" {
			personalizeBody, _ := json.Marshal(map[string]string{
				"name":         req.Name,
				"display_name": req.DisplayName,
				"gender":       req.Gender,
			})
			httpReq, _ := http.NewRequest(http.MethodPut,
				userServiceURL+"/users/"+req.Phone+"/personalize",
				bytes.NewBuffer(personalizeBody))
			httpReq.Header.Set("Content-Type", "application/json")
			http.DefaultClient.Do(httpReq)
		}

		getUserResp, err := http.Get(userServiceURL + "/users/" + req.Phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}
		defer getUserResp.Body.Close()

		var userResult struct {
			ID string `json:"id"`
		}
		body, _ := io.ReadAll(getUserResp.Body)
		json.Unmarshal(body, &userResult)

		if userResult.ID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user ID"})
			return
		}

		now := time.Now()
		uID, _ := primitive.ObjectIDFromHex(userResult.ID)
		driver := &models.Driver{
			Phone:           req.Phone,
			UserID:          uID,
			VehicleType:     req.VehicleType,
			PlateNumber:     req.PlateNumber,
			IsActive:        true,
			Status:          "offline",
			ReputationScore: 0,
			TotalOrders:     0,
			TotalRejects:    0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := driverRepo.Create(c.Request.Context(), driver); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver"})
			return
		}

		if isNewUser {
			fmt.Printf("Admin: Created driver %s — new user registered + driver\n", req.Phone)
		} else {
			fmt.Printf("Admin: Created driver %s — linked to existing user %s\n", req.Phone, userResult.ID)
		}
		c.JSON(http.StatusCreated, driver)
	})

	r.POST("/drivers/auth", func(c *gin.Context) {
		var req struct {
			Phone       string `json:"phone" binding:"required"`
			UserID      string `json:"user_id"`
			VehicleType string `json:"vehicle_type"`
			PlateNumber string `json:"plate_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
			return
		}

		driver, err := driverRepo.GetByPhone(c.Request.Context(), req.Phone)
		if err != nil {
			now := time.Now()
			uID, _ := primitive.ObjectIDFromHex(req.UserID)
			driver = &models.Driver{
				Phone:           req.Phone,
				UserID:          uID,
				VehicleType:     req.VehicleType,
				PlateNumber:     req.PlateNumber,
				IsActive:        true,
				Status:          "offline",
				ReputationScore: 0,
				TotalOrders:     0,
				TotalRejects:    0,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := driverRepo.Create(c.Request.Context(), driver); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"token":  "driver_" + req.Phone,
			"driver": driver,
		})
	})

	r.POST("/drivers/auth/logout", func(c *gin.Context) {
		phone := c.GetString("phone")
		if phone == "" {
			var req struct {
				Phone string `json:"phone"`
			}
			c.ShouldBindJSON(&req)
			phone = req.Phone
		}
		if phone != "" {
			driverRepo.Update(c.Request.Context(), phone, bson.M{"status": "offline"})
		}
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	})

	r.GET("/drivers/me", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		driver, err := driverRepo.GetByPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
			return
		}
		c.JSON(http.StatusOK, driver)
	})

	r.PUT("/drivers/me", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		var req struct {
			VehicleType *string `json:"vehicle_type"`
			PlateNumber *string `json:"plate_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"updated_at": time.Now()}
		if req.VehicleType != nil {
			update["vehicle_type"] = *req.VehicleType
		}
		if req.PlateNumber != nil {
			update["plate_number"] = *req.PlateNumber
		}

		if err := driverRepo.Update(c.Request.Context(), phone, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		driver, _ := driverRepo.GetByPhone(c.Request.Context(), phone)
		c.JSON(http.StatusOK, driver)
	})

	r.PUT("/drivers/me/status", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		var req struct {
			Status      *string `json:"status"`
			VehicleType *string `json:"vehicle_type"`
			PlateNumber *string `json:"plate_number"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		update := bson.M{"updated_at": time.Now()}
		if req.Status != nil {
			update["status"] = *req.Status
		}

		if err := driverRepo.Update(c.Request.Context(), phone, update); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}

		driver, _ := driverRepo.GetByPhone(c.Request.Context(), phone)
		c.JSON(http.StatusOK, driver)
	})

	r.GET("/drivers/debts", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		debts, err := debtRepo.GetByDriverPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, debts)
	})

	r.GET("/drivers/ratings/pending", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		ratings, err := ratingRepo.GetPendingByDriverPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, ratings)
	})

	r.POST("/drivers/ratings/:id", func(c *gin.Context) {
		id := c.Param("id")
		var req struct {
			Score   float64 `json:"score" binding:"required,min=1,max=5"`
			Comment string  `json:"comment"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := ratingRepo.SubmitRating(c.Request.Context(), id, req.Score, req.Comment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "rating submitted"})
	})

	r.GET("/drivers/orders", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		orders, err := orderRepo.GetByDriverPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
	})

	r.GET("/drivers/reputation", func(c *gin.Context) {
		phone := c.Query("phone")
		if phone == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param required"})
			return
		}
		driver, err := driverRepo.GetByPhone(c.Request.Context(), phone)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "driver not found"})
			return
		}

		var totalRatings []models.Rating
		ctx := context.Background()
		// Use the db directly for a simple count query
		// We'll use the rating repo pattern
		ratingColl := db.Collection("ratings")
		cursor, err := ratingColl.Find(ctx, bson.M{
			"driver_phone": phone,
			"score":        bson.M{"$exists": true, "$ne": 0},
		})
		if err == nil {
			defer cursor.Close(ctx)
			cursor.All(ctx, &totalRatings)
		}

		avgRating := 0.0
		if len(totalRatings) > 0 {
			sum := 0.0
			for _, r := range totalRatings {
				sum += r.Score
			}
			avgRating = sum / float64(len(totalRatings))
		}

		c.JSON(http.StatusOK, gin.H{
			"reputation_score": driver.ReputationScore,
			"total_orders":     driver.TotalOrders,
			"total_rejects":    driver.TotalRejects,
			"total_ratings":    len(totalRatings),
			"average_rating":   avgRating,
		})
	})

	r.GET("/drivers", func(c *gin.Context) {
		drivers, err := driverRepo.GetAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
			return
		}
		c.JSON(http.StatusOK, drivers)
	})

	r.DELETE("/drivers/:phone", func(c *gin.Context) {
		phone := c.Param("phone")
		if err := driverRepo.Delete(c.Request.Context(), phone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete driver"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Driver deleted"})
	})

	port := config.GetEnv("DRIVER_SERVICE_PORT", "8082")
	r.Run(":" + port)
}
