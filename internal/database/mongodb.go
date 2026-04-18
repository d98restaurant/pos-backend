package database

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func ConnectMongoDB(uri, dbName string) (*MongoDB, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, err
    }
    
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    
    log.Println("✅ MongoDB connected successfully")
    
    return &MongoDB{
        Client:   client,
        Database: client.Database(dbName),
    }, nil
}

func (m *MongoDB) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return m.Client.Disconnect(ctx)
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
    return m.Database.Collection(name)
}