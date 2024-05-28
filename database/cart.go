package database

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/purandixit07/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find the item")
	ErrCantDecodeProducts = errors.New("can't find the product")
	ErrUserIDNotValid     = errors.New("this user is not valid")
	ErrCantUpdateUser     = errors.New("cannot update the user details")
	ErrCantRemoveItemCart = errors.New("cannot remove item from cart")
	ErrCantGetItem        = errors.New("not able to get this item from cart")
	ErrCantBuyCartItem    = errors.New("cannot update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchFromDB, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}
	var productCart []models.ProductUser
	err = searchFromDB.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: primitive.E{Key: "usercart", Value: bson.D{{Key: "$each", Value: productCart}}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	return nil
}

func RemoveCartItem(ctx context.Context, userCollection, productCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	userid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: userid}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantRemoveItemCart
	}
	return nil
}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	userid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDNotValid
	}
	var orderCart models.Order
	var getCartItems models.User

	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Ordered_At = time.Now()
	orderCart.Payment_Method.COD = true
	orderCart.Order_Cart = make([]models.ProductUser, 0)

	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}

	var getusercart []bson.M
	err = currentResults.All(ctx, &getusercart)
	if err != nil {
		panic(err)
	}
	var totalPrice int32
	for _, user_item := range getusercart {
		price := user_item["total"]
		totalPrice = price.(int32)
	}
	orderCart.Price = int(totalPrice)
	filter := bson.D{primitive.E{Key: "_id", Value: userid}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: userid}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key: "_id", Value: userid}}
	update2 := bson.M{"$push": bson.M{"orders.$[]order_list": bson.M{"$each": getCartItems.UserCart}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}

	usercart_empty := make([]models.ProductUser, 0)
	filter3 := bson.D{primitive.E{Key: "_id", Value: userid}}
	update3 := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: usercart_empty}}}}
	_, err = userCollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		return ErrCantBuyCartItem
	}
	return nil
}

func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIDNotValid
	}

	var product_details models.ProductUser
	var order_detail models.Order

	order_detail.Order_ID = primitive.NewObjectID()
	order_detail.Ordered_At = time.Now()
	order_detail.Order_Cart = make([]models.ProductUser, 0)
	order_detail.Payment_Method.COD = true
	err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&product_details)
	if err != nil {
		log.Println(err)
	}
	order_detail.Price = product_details.Price

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: order_detail}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": product_details}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}
	return nil
}
