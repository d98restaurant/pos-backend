package database

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func NewMongoClient(uri, dbName string) (*MongoClient, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Set client options
    clientOptions := options.Client().ApplyURI(uri)
    
    // Connect to MongoDB
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }
    
    // Ping the database
    err = client.Ping(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    log.Println("✅ Connected to MongoDB Atlas successfully!")
    
    return &MongoClient{
        Client:   client,
        Database: client.Database(dbName),
    }, nil
}

func (m *MongoClient) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return m.Client.Disconnect(ctx)
}

func (m *MongoClient) GetCollection(name string) *mongo.Collection {
    return m.Database.Collection(name)
}