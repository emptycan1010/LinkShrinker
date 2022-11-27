package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/rand"
	"time"
)

type Links struct {
	Link       string `bson:"link"`
	Createtime int64  `bson:"createtime"`
	Name       string `bson:"name"`
}

type MakeSHRURLform struct {
	Link          Links
	ReturnChannel chan string
}

type Queryform struct {
	Name          string `bson:"name"`
	ReturnChannel chan []Links
}

var querychan = make(chan Queryform)
var makeshrurlchan = make(chan MakeSHRURLform)

func main() {
	go makeShrURL()
	go queryit()
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": RandomString(8)}) })
	r.GET("/:s", letsgo)
	r.POST("/make", letsmakeit)
	r.Run()
}

func letsgo(c *gin.Context) {
	q := Queryform{Name: c.Param("s"), ReturnChannel: make(chan []Links)}
	querychan <- q
	databases := <-q.ReturnChannel
	//c.JSON(200, gin.H{"message": databases})
	c.Redirect(302, databases[0].Link)
}

func letsmakeit(c *gin.Context) {
	//m := MakeSHRURLform{Link: Links{Link: c.PostForm("link"), Createtime: time.Now().Unix(), Name: c.PostForm("name")}, ReturnChannel: make(chan string)}
	m := MakeSHRURLform{Link: Links{Link: c.PostForm("link"), Createtime: time.Now().Unix(), Name: RandomString(8)}, ReturnChannel: make(chan string)}

	makeshrurlchan <- m
	c.JSON(200, gin.H{"message": <-m.ReturnChannel})
}

func makeShrURL() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("link").Collection("arr")

	// Db Connected
	for {
		q := <-makeshrurlchan
		// Let's Check name already exists
		qq := Queryform{Name: q.Link.Name, ReturnChannel: make(chan []Links)}
		querychan <- qq
		databases := <-qq.ReturnChannel
		if q.Link.Name == "make" {
			q.ReturnChannel <- "This name is not allowed. Try again Please"
		} else {
			if len(databases) == 0 {
				_, err := collection.InsertOne(ctx, q.Link)
				q.ReturnChannel <- fmt.Sprintf("http://siro157.xyz:8080/%s", q.Link.Name)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				q.ReturnChannel <- "Name already exists"
			}
		}
	}
}

func queryit() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	collection := client.Database("link").Collection("arr")

	// DB Connected
	for {
		q := <-querychan

		cur, err := collection.Find(ctx, bson.M{"name": q.Name})
		if err != nil {
			log.Fatal(err)
		}
		var databases []Links
		cur.All(ctx, &databases)
		fmt.Println(databases)
		q.ReturnChannel <- databases
	}
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
