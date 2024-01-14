package store

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ShortUrl struct {
	Base  string
	Short string
}

type Store struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var ErrNotFound error = errors.New("entry not found in store")

func NewStore(dbUrl, dbName, collectionName string) (*Store, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(dbUrl)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Printf("failed to connect to %s, err: %v", dbUrl, err)
		return nil, err
	}

	// Check the connection
	if err = client.Ping(ctx, nil); err != nil { // Reuse the same context intentionally
		log.Printf("database connection is not alive (url: %s), err: %v", dbUrl, err)
		return nil, err
	}

	// Retrieve target collection
	db := client.Database(dbName)
	if db == nil {
		log.Printf("failed to find database %s, err: %v", dbName, err)
		return nil, err
	}

	collection := db.Collection(collectionName)
	return &Store{client, collection}, nil
}

func (s *Store) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}

func (s *Store) Get(shortUrl string) (*ShortUrl, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"short": shortUrl}
	result := &ShortUrl{}
	if err := s.collection.FindOne(ctx, filter).Decode(result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound // Convert error to abstract from MongoDb
		} else {
			log.Printf("failed to get short url entry, err: %v", err)
		}
		return nil, err
	}
	return result, nil
}

// Allow multiple entries pointing to the same base url for ownership and stats concerns
func (s *Store) Create(shortUrl ShortUrl) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newValue := bson.M{
		"base":  shortUrl.Base,
		"short": shortUrl.Short,
	}
	if _, err := s.collection.InsertOne(ctx, newValue); err != nil {
		log.Printf("failed to create short url entry, err: %v", err)
		return err
	}

	return nil
}

func (s *Store) Delete(shortUrl string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"short": shortUrl}
	res, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("failed to delete short url entry, err: %v", err)
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}
