package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// struct for users collection in database
// used for binding collection row from database
type Student struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FirstName string             `form:"firstname" json:"firstname" bson:"firstname"`
	LastName  string             `form:"lastname" json:"lastname" bson:"lastname"`
	ClassName string             `form:"classname" json:"classname" bson:"classname"`
	ObjectID  string             `form:"object_id" bson:"object_id,omitempty"`
}

func main() {
	r := gin.New()
	//Create Users
	r.POST("student/add", func(c *gin.Context) {

		var input Student
		c.Bind(&input)
		var ctx = context.Background()
		dbm, _ := Monggo()
		defer dbm.Client().Disconnect(ctx)
		//insert data to users collection
		_, err := dbm.Collection("student").InsertOne(context.TODO(), input)
		//if error, return error and stop
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		//if no error occurs
		c.JSON(http.StatusOK, gin.H{"message": "data inserted"})
	})
	//display users
	r.GET("student/list", func(c *gin.Context) {
		var input Student
		c.Bind(&input)

		searchFirstName := input.FirstName
		searchLastName := input.LastName
		searchClassName := input.ClassName

		var rows []Student
		var ctx = context.Background()
		dbm, _ := Monggo()
		defer dbm.Client().Disconnect(ctx)
		filter := bson.D{{}}
		criteria := bson.A{}
		//if searchRole not empty, create filter criteria for role field
		if searchLastName != "" {
			//This search means that the "username" has characters with the given keyword.
			//does not need to be exactly the same.
			criteria = append(criteria, bson.D{{Key: "lastname", Value: bson.D{
				{Key: "$regex", Value: searchLastName},
				{Key: "$options", Value: "i"},
			}}})
		}
		//if searchUsername not empty, create filter criteria for username field
		if searchFirstName != "" {
			//This search means that the "username" has characters with the given keyword.
			//does not need to be exactly the same.
			criteria = append(criteria, bson.D{{Key: "firstname", Value: bson.D{
				{Key: "$regex", Value: searchFirstName},
				{Key: "$options", Value: "i"},
			}}})
		}

		if searchClassName != "" {
			//This search means that the "username" has characters with the given keyword.
			//does not need to be exactly the same.
			criteria = append(criteria, bson.D{{Key: "classname", Value: bson.D{
				{Key: "$regex", Value: searchClassName},
				{Key: "$options", Value: "i"},
			}}})
		}

		//Enter criteria into the search filter
		if len(criteria) > 0 {
			filter = bson.D{{Key: "$and", Value: criteria}}
		}
		sort := bson.D{{Key: "firstname", Value: 1}}
		opts := options.Find().SetSort(sort)
		cursor, err := dbm.Collection("student").Find(ctx, filter, opts)
		if err != nil {
			//if an error occurs,
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error on searching data"})
			return
		} else {
			//put search result from mongoDB to rows variable
			err = cursor.All(context.TODO(), &rows)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "error on put data"})
				return
			}
		}
		//if no error occurs, return data to frontend
		c.JSON(http.StatusOK, rows)
	})
	//get single user(Used for set data in edit form)
	r.GET("student/get_row/:id", func(c *gin.Context) {
		id := c.Param("id")
		var row Student
		var ctx = context.Background()
		dbm, _ := Monggo()

		defer dbm.Client().Disconnect(ctx)

		docID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			println(err.Error())
		}
		//create filter
		filter := bson.D{{Key: "_id", Value: docID}}
		err = dbm.Collection("student").FindOne(ctx, filter).Decode(&row)
		if err != nil {
			//if an error occurs,
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error on searching data"})
			return
		}
		//if no error occurs, return data to frontend
		c.JSON(http.StatusOK, row)
	})
	//update users
	r.POST("student/single_update", func(c *gin.Context) {
		var ctx = context.Background()
		dbm, _ := Monggo()

		defer dbm.Client().Disconnect(ctx)

		var input Student
		c.Bind(&input)
		FirstName := input.FirstName
		LastName := input.LastName
		ClassName := input.ClassName
		//Create MongoDB object ID
		//because we use "_id" as a filter
		docID, err := primitive.ObjectIDFromHex(input.ObjectID)
		if err != nil {
			println(err.Error())
		}
		//create filter
		filter := bson.D{{Key: "id", Value: docID}}
		//data field to be updated
		update := bson.M{
			"$set": bson.M{
				"firstname": FirstName,
				"lastname":  LastName,
				"classname": ClassName,
			},
		}
		//update one document / row
		_, err = dbm.Collection("student").UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "data updated"})
		}
	})
	//delete user
	r.DELETE("student/single_delete/:id", func(c *gin.Context) {

		id := c.Param("id")

		var ctx = context.Background()
		dbm, _ := Monggo()
		defer dbm.Client().Disconnect(ctx)
		//Create MongoDB object ID
		//because we use "_id" as a filter
		docID, _ := primitive.ObjectIDFromHex(id)
		//create filter
		filter := bson.D{{Key: "_id", Value: docID}}
		_, err := dbm.Collection("student").DeleteOne(ctx, filter)
		if err != nil {
			c.JSON(http.StatusNotModified, gin.H{"message": "delete error"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "succeed"})
		}
	})
	//Allow POST, GET, DELETE Request from front end
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"OPTIONS", "POST", "GET", "DELETE"},
	})
	//create HTTP server
	handler := c.Handler(r)
	http.ListenAndServe("localhost:8080", handler)
}

// function to create connection to MongoDB server
func Monggo() (*mongo.Database, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().
		ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	return client.Database("exercise"), nil
}
