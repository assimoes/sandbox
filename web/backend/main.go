package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoURI       = getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName         = getEnv("MONGO_DB", "logsdb")
	collectionName = getEnv("MONGO_COLLECTION", "logs")

	collection *mongo.Collection
	ctx        context.Context
	cancel     context.CancelFunc
)

func main() {

	app := fiber.New()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))

	if err != nil {
		log.Printf("error creating mongodb client: %v", err)
	}

	db := client.Database(dbName)
	collection = db.Collection(collectionName)

	// Logging for each request
	app.Use(logger.New())

	// Serve static content app
	app.Static("/", "../frontend/public")

	// Serve api endpoints
	app.Get("/api/executions", getExecIds)
	app.Get("/api/executions/:execution_id", getExecDetails)

	// Starts the server
	log.Fatal(app.Listen(":3000"))
}

func getExecIds(c *fiber.Ctx) error {
	pageParam := c.Query("page")
	page, err := strconv.Atoi(pageParam)

	if err != nil || page < 1 {
		return c.Status(http.StatusBadRequest).SendString("invalid page number")
	}

	pageSize := 25
	skip := (page - 1) * pageSize

	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", "$execution_id"},
			{"timestamp", bson.D{{"$first", "$timestamp"}}},
		}}}

	sortStage := bson.D{
		{"$sort", bson.D{
			{"timestamp", 1},
		}},
	}

	skipStage := bson.D{
		{"$skip", int64(skip)},
	}

	limitStage := bson.D{
		{"$limit", int64(pageSize)},
	}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{groupStage, sortStage, skipStage, limitStage})
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("error fetching execution ids from mongo")
	}

	var executionIDs []string

	for cursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}

		if err := cursor.Decode(&result); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("error deconding execution id")
		}
	}

	response := map[string]interface{}{
		"page":        page,
		"total_pages": len(executionIDs) / pageSize,
		"data":        executionIDs,
	}

	return c.JSON(response)
}

func getExecDetails(c *fiber.Ctx) error {
	executionId := c.Params("execution_id")

	if executionId == "" {
		return c.Status(http.StatusBadRequest).SendString("invalid execution id")
	}

	matchStage := bson.D{
		{"$match", bson.D{
			{"execution_id", executionId},
		}}}

	sortStage := bson.D{
		{"$sort", bson.D{
			{"timestamp", 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{matchStage, sortStage})
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("error fetching execution ids from mongo")
	}

	var executionDetails []map[string]interface{}

	for cursor.Next(ctx) {
		var result map[string]interface{}
		if err := cursor.Decode(&result); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("error decoding execution details")
		}

		executionDetails = append(executionDetails, result)
	}

	response := map[string]interface{}{
		"execution_id": executionId,
		"data":         executionDetails,
	}

	return c.JSON(response)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
