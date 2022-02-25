package database

import (
	"context"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultDb       = "dutchman"
	usersCollection = "users"
	uri             = "mongodb://localhost:27017/?maxPoolSize=20&w=majority"
)

type Config struct {
}

type Database struct {
	cfg Config

	client *mongo.Client
}

func NewDatabase(cfg Config) (*Database, error) {
	d := &Database{
		cfg: cfg,
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}
	d.client = client

	return d, nil
}

func (d *Database) Shutdown() {
	if err := d.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (d *Database) StoreUser(user *models.User) error {
	return d.storeModel(usersCollection, user)
}

func (d *Database) storeModel(collection string, model interface{}) error {
	coll := d.client.Database(defaultDb).Collection(collection)

	res, err := coll.InsertOne(context.TODO(), model)
	if err != nil {
		return err
	}
	logrus.Infof("inserted doc with _id: %v", res.InsertedID)

	return nil
}

func (d *Database) CheckIfUserExists(email string) bool {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	count, err := coll.CountDocuments(context.TODO(), bson.M{"email": email})
	if err != nil {
		logrus.Errorf("error counting users: %v", err)
		return true
	}

	return count > 0
}

func (d *Database) GetUser(id string) *models.User {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	res := coll.FindOne(context.TODO(), bson.M{"id": id})
	if res.Err() != nil {
		logrus.Errorf("error getting result from mongo: %v", res.Err())
		return nil
	}

	var user models.User
	err := res.Decode(&user)
	if err != nil {
		logrus.Errorf("error decoding result: %v", err)
		return nil
	}

	return &user
}

func (d *Database) GetUserByEmail(email string) *models.User {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	res := coll.FindOne(context.TODO(), bson.M{"email": email})
	if res.Err() != nil {
		logrus.Errorf("error getting result from mongo: %v", res.Err())
		return nil
	}

	var user models.User
	err := res.Decode(&user)
	if err != nil {
		logrus.Errorf("error decoding result: %v", err)
		return nil
	}

	return &user
}
