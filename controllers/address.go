package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/purandixit07/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}
		address, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.IndentedJSON(500, "internal server error")
		}

		var addresses models.Address
		addresses.Address_ID = primitive.NewObjectID()
		if err := c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		group := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, group})
		if err != nil {
			c.IndentedJSON(500, "internal server error")
		}

		var addressInfo []bson.M
		if err = pointcursor.All(ctx, &addressInfo); err != nil {
			panic(err)
		}
		var size int32
		for _, address_no := range addressInfo {
			count := address_no["count"]
			size = count.(int32)
		}

		if size < 2 {
			filter := bson.D{primitive.E{Key: "_id", Value: address}}
			update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
			_, err := UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			c.IndentedJSON(400, "not allowed")
		}

		defer cancel()
		ctx.Done()
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}
		userID, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.IndentedJSON(500, "internal server error")
		}

		var editaddress models.Address
		if err = c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: userID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.0.house_name", Value: editaddress.House}, {Key: "address.0.street_name", Value: editaddress.Street}, {Key: "address.0.city_name", Value: editaddress.City}, {Key: "address.0.pin_code", Value: editaddress.Pincode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(500, "something went wrong")
			return
		}
		ctx.Done()
		c.IndentedJSON(200, "successfully updated the home address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.IndentedJSON(500, "internal server error")
		}
		var editaddress models.Address
		if err = c.BindJSON(&editaddress); err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: userID}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address.1.house_name", Value: editaddress.House}, {Key: "address.1.street_name", Value: editaddress.Street}, {Key: "address.1.city_name", Value: editaddress.City}, {Key: "address.1.pin_code", Value: editaddress.Pincode}}}}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(500, "something went wrong")
			return
		}
		ctx.Done()
		c.IndentedJSON(200, "successfully updated the work address")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		user_id, err := primitive.ObjectIDFromHex(userQueryID)
		if err != nil {
			c.IndentedJSON(500, "Internal server address")
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		filter := bson.D{primitive.E{Key: "_id", Value: user_id}}
		update := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(404, "Couldn't update")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "successfully deleted")
	}
}
