package main

import (
	"context"
    "fmt"
    "log"
	"encoding/json"
	"os/exec"
	"strings"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/k0kubun/pp/v3"
    "github.com/gin-gonic/gin"
)

type PageInfos struct {
	Name  	    string
	Reviews     []interface{} 
	Zoom        int
	Latitude    float64
	Longitude   float64
	Id          string
	Thumbnail   string
	Photos      []interface{}
	Address     string
	Description string
	Phone 		string
	Email 		string
}

type MongoDB struct {
	client *mongo.Client
}

func (c MongoDB) query(database, collection string, filter bson.D) (bson.M, error) {
	coll := c.client.Database(database).Collection(collection) 

	var result bson.M

	err := coll.FindOne(context.TODO(), filter).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, fmt.Errorf("no document found with the specified filter")
        }
        return nil, err
    }
    return result, nil
}

func (c MongoDB) queryMany(database, collection string, filter bson.D) ([]bson.M, error) {

	coll := c.client.Database(database).Collection(collection) 
	var result []bson.M

	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func Filter (items []interface{}, predicate func(interface{}) bool) []interface{} {
	var r []interface{}
	for _, item := range items {
		if predicate(item) {
			r = append(r, item)
		}
	}
	return r
}

func isGoodReview (review interface{}) bool {
	return review.(map[string]interface{})["Rating"].(float64) >= 4

}

func isFoodRelatedPhoto (photo interface{}) bool {
	title := photo.(map[string]interface{})["title"]
	return title != "Latest" && title != "Videos" && title != "Street View & 360Â°"
}

func safeName(name string) string {
	s := strings.ReplaceAll(name, " ", "_")
	s = strings.ReplaceAll(s, "&", "_")
	return strings.ToLower(s)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
        return false
    }
	return err == nil
}

func getImages(iterable []interface{}, name string) {
	var imgLink, imgTitle string

	switch name {
	case "reviews":
		imgTitle = "Name"
		imgLink  = "ProfilePicture"
	case "photos":
		imgTitle = "title"
		imgLink  = "image"
	}

	for _, iter := range iterable {
		r := iter.(map[string]interface{})
		imgPath :=  fmt.Sprintf("static/images/google/%s/%s.webp", name, safeName(r[imgTitle].(string)))
		r["img"] = imgPath
		
		
		if ! pathExists(imgPath) {
			cmd := exec.Command("wget", r[imgLink].(string), "-O", imgPath)
			if err := cmd.Run(); err != nil {
				fmt.Println("Error executing wget:", err)
				return
			}
		}
	}
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(context.TODO(), nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Connected to MongoDB!")
    
    defer func() {
        if err = client.Disconnect(context.TODO()); err != nil {
            log.Fatal(err)
        }
    }()

	mongoClient := MongoDB{client}

	// id_db := "0x47869d393cea6871:0x8ec950bda20e3cc5"
	id_db := "0x133ab7fcd1dfa271:0x7f390a5526d00362"
    filter := bson.D{{"_id", id_db}} 

    r1, err := mongoClient.query("gms", "biz", filter)
    if err != nil {
        log.Fatal(err)
    }
	// pp.Print(r1["user_reviews"].(string))
	fmt.Println(r1["user_reviews"])

	var reviews []interface{}
	err = json.Unmarshal([]byte(r1["user_reviews"].(string)), &reviews)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	var photos []interface{}
	err = json.Unmarshal([]byte(r1["images"].(string)), &photos)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	getImages(photos, "photos")
	getImages(reviews, "reviews")

	goodReviews := Filter(reviews, isGoodReview)
	goodPhotos  := Filter(photos, isFoodRelatedPhoto)
	name        := r1["title"].(string)
	lon         := r1["longitude"].(float64)
	lat         := r1["latitude"].(float64)
	id          := r1["_id"].(string)
	thumbnail   := r1["thumbnail"].(string) 
	address     := r1["address"].(string)
	phone       := r1["phone"].(string)
	
	pp.Println(photos)
	pp.Println(reviews)

    // fmt.Printf("Found a document: %+v\n\n\n", r1)

	// r2, err := mongoClient.queryMany("gms", "biz", filter)
    // if err != nil {
    //     log.Fatal(err)
    // }

	// pp.Print(r2)
    // fmt.Printf("Found some documents: %+v\n\n\n", r2)
	// b, err := json.MarshalIndent(r2, "", "  ")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(string(b))

    r := gin.Default()

	r.Static("/static", "./static")
	r.LoadHTMLFiles("templates/index.html")

    r.GET("/", func(c *gin.Context) {
		pageInfos := PageInfos{
			Name:  name,
			Reviews: goodReviews,
			Zoom: 5000,
			Longitude: lon,
			Latitude: lat,
			Id: id,
			Thumbnail: thumbnail,
			Photos: goodPhotos,
			Address: address,
			Description: "Quality food and drinks in the heart of Como",
			Phone: phone,
			Email: "info@example.com",

		}
        c.HTML(200, "index.html", pageInfos)
    })

	if err := r.Run(":9999"); err != nil {
        log.Fatal(err)
    }

}