package tokens

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/purandixit07/ecommerce/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SECRET_KEY = os.Getenv("SECRET_KEY")
var UserData *mongo.Collection = database.UserData(database.Client, "Users")

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	UID        string
	jwt.StandardClaims
}

func TokenGenerator(email, first_name, last_name, uid string) (signedToken, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_Name: first_name,
		Last_Name:  last_name,
		UID:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES384, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, nil
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "the token is invalid"
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = "token is already expired"
		return
	}
	return claims, msg
}

func UpdateAllTokens(signedToken, signedRefreshToken, userID string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObject primitive.D
	updateObject = append(updateObject, bson.E{Key: "token", Value: signedToken})
	updateObject = append(updateObject, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObject = append(updateObject, bson.E{Key: "updated_at", Value: updatedAt})

	upsert := true
	filter := bson.M{"user_id": userID}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := UserData.UpdateOne(ctx, filter, bson.D{{
		Key: "$set", Value: updateObject,
	},
	}, &opt)
	if err != nil {
		log.Panic(err)
		return
	}
}
