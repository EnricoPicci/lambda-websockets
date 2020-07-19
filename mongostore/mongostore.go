package mongostore

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoStore is the mongodb reference
type MongoStore struct {
	db *mongo.Database
}

// NewMongoStore creates a MongoStore
func NewMongoStore(ctx context.Context) *MongoStore {
	var store = MongoStore{
		db: connect(ctx),
	}
	return &store
}

// connect to db
func connect(ctx context.Context) *mongo.Database {
	// Database Config
	connString := os.Getenv("MONGO_URI")
	if connString == "" {
		panic("mongo connection string has not been provided - please set it in the env var MONGO_URI")
	}
	clientOptions := options.Client().ApplyURI(connString)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal("Error creating Mongo Client", err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal("Error connecting to Mongo", err)
	}

	err = client.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected!")
	}
	return client.Database(os.Getenv("MONGO_DATABASE"))
}

const (
	connectionsCollName = "connections"
	connActive          = "active"
	connClosed          = "closed"
)

type connectionIDEntry struct {
	CreationTs   time.Time
	DisconnectTs time.Time
	ConnectionID string
	Status       string
}

func (store *MongoStore) readConnections(ctx context.Context) ([]string, error) {
	collection := store.db.Collection(connectionsCollName)
	filter := bson.M{"status": bson.M{"$eq": "active"}}
	cur, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Println("Error while opening the curson on the connection collection", err)
		return nil, err
	}

	connections := []string{}
	for cur.Next(ctx) {
		var elem connectionIDEntry
		err = cur.Decode(&elem)
		if err != nil {
			log.Println("Error while reading the cursor", err)
			return nil, err
		}

		connections = append(connections, elem.ConnectionID)
	}
	return connections, nil

}

// GetConnectionIDs returns the connectionIDs - waits for the read from DB to be concluded
func (store *MongoStore) GetConnectionIDs(ctx context.Context) ([]string, error) {
	return store.readConnections(ctx)
}

// AddConnectionID adds a record representing a connectionId
func (store *MongoStore) AddConnectionID(ctx context.Context, connectionID string) error {
	entry := connectionIDEntry{}
	entry.CreationTs = time.Now()
	entry.ConnectionID = connectionID
	entry.Status = connActive
	collection := store.db.Collection(connectionsCollName)
	_, err := collection.InsertOne(ctx, entry)
	return err
}

// MarkConnectionIDDisconnected marks a connectionID as disconnected
func (store *MongoStore) MarkConnectionIDDisconnected(ctx context.Context, connectionID string) error {
	collection := store.db.Collection(connectionsCollName)
	// https://stackoverflow.com/a/54548495/5699993
	filter := bson.D{primitive.E{Key: "connectionid", Value: connectionID}}
	// https://stackoverflow.com/a/23583746/5699993
	update := bson.M{
		"$set": bson.M{"disconnectts": time.Now(), "status": connClosed},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}
