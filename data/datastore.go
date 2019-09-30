package data

import (
	"context"
	"fmt"
	"gatso/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"time"
)

const connectionTimeout = time.Minute * 2 // Time connection to DB waits before giving up
const databaseName = "todo"
const tasksCollectionName = "todo_tasks"
const maxTaskCount = 500 // maximum number of tasks returned in one request.

type Datastore interface {
	// Retrieve all the tasks owned by the given id.  if id unknown, returns nil
	GetTasks(ownerId int) ([]*model.Task, error)

	// Retrieve all the tasks NOT owned by the given id, but visisble to them.
	GetOthersTasks(ownerId int) ([]*model.Task, error)

	FindTasks(ownerId int, query model.Task) ([]*model.Task, error)

	// Get the number of tasks owned by the given ownerId
	CountTasks(ownerId int) int

	// Add a new Task to the owners list
	AddTask(ownerId int, task model.Task) (string, error)

	// Add or replace the given task with the same ID
	UpdateTask(ownerId int, task model.Task) error

	// Delete the task with the given Id, if it belongs to the given owner id.
	DeleteTask(ownerId int, taskId string) bool

	// Close the datastore and release connections.
	Close()

	// List all the user ID's known
	Users()  ([]int, error)
}

// MongoDb implementation of the datastore
type MongoDataStore struct {
	client         *mongo.Client
	db             *mongo.Database
	collectionName string
}

// Create a new MongoDataStore with the given connection uri to the mongo database.
func NewMongoDataStore(uri string) (*MongoDataStore, error) {

	u, err := url.Parse(uri)
	if nil != err {
		return nil, err
	}
	colName := u.Fragment // collection name can be specified under the fragment (For testing)
	if colName == "" {
		colName = tasksCollectionName
	} else {
		u.Fragment = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	clientOptions := options.Client().ApplyURI(u.String())
	client, err := mongo.Connect(ctx, clientOptions)
	if nil != err {
		return nil, err
	}

	if err := client.Ping(ctx, nil); nil != err {
		return nil, err
	}

	db := client.Database(databaseName)
	return &MongoDataStore{
		db:             db,
		client:         client,
		collectionName: colName,
	}, nil
}

// Drop will destroy the entire tasks database. (Used for testing)
func (m MongoDataStore) Drop() error {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	return m.db.Drop(ctx)
}

// Close the connections to mongo and release the resources.
func (m MongoDataStore) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	m.client.Disconnect(ctx)
}

func (m MongoDataStore) GetTasks(ownerId int) ([]*model.Task, error) {
	return m.query(bson.D{{"owner", ownerId}}, bson.D{{"expires", -1}})
}

func (m MongoDataStore) GetOthersTasks(ownerId int) ([]*model.Task, error) {
	// single element query on an array returns any item with an array containing that value
	return m.query(bson.D{{"readers", ownerId}}, bson.D{{"expires", -1}})
}

func (m MongoDataStore) FindTasks(ownerId int, query model.Task) ([]*model.Task, error) {

	doc := bson.D{}
	doc = append(doc, bson.E{"owner", ownerId})

	if query.Title != "" {
		doc = append(doc, bson.E{"title", query.Title})
	}

	if !query.Expires.IsZero() {
		by, err := bson.Marshal(&query.Expires)
		if nil != err {
			return nil, err
		}

		rVal := bson.D{{"$lt", bson.Raw(by)}}
		doc = append(doc, bson.E{"expires", rVal})
	}

	if !query.Created.IsZero() {
		by, err := bson.Marshal(&query.Created)
		if nil != err {
			return nil, err
		}

		rVal := bson.D{{"$gte", bson.Raw(by)}}
		doc = append(doc, bson.E{"created", rVal})
	}

	if len(query.Readers) != 0 {
		by, err := bson.Marshal(&query.Readers)
		if nil != err {
			return nil, err
		}

		rVal := bson.D{{"$all", bson.Raw(by)}}
		doc = append(doc, bson.E{"readers", rVal})
	}

	if len(query.Labels) != 0 {
		items := bson.A{}
		for _, s := range query.Labels {
			items = append(items, s)
		}

		rVal := bson.D{{"$all", items}}
		doc = append(doc, bson.E{"labels", rVal})
	}

	if len(query.Notes) != 0 {
		by, err := bson.Marshal(&query.Notes)
		if nil != err {
			return nil, err
		}

		rVal := bson.D{{"$all", bson.Raw(by)}}
		doc = append(doc, bson.E{"notes", rVal})
	}

	return m.query(doc, bson.D{{"expires", -1}})
}

func (m MongoDataStore) AddTask(ownerId int, task model.Task) (string, error) {
	if task.Owner != ownerId {
		return "", fmt.Errorf("Owner %d does not own the given task to add", ownerId)
	}

	task.Created = time.Now()
	task.ID = nil

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	result, err := m.collection().InsertOne(ctx, &task)
	if nil != err {
		return "", err
	}
	oid, ok := result.InsertedID.(primitive.ObjectID);
	if !ok {
		return "", fmt.Errorf("failed to read new id of inserted item")
	}
	return oid.Hex(), nil
}

func (m MongoDataStore) UpdateTask(ownerId int, task model.Task) error {
	existing := m.GetTask(task.Id())
	if nil == existing { // doesn't exist, treat as an Add
		_, err := m.AddTask(ownerId, task)
		return err
	}
	if existing.Owner != ownerId {
		return fmt.Errorf("Owner %d does not own the given task to update", ownerId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	by, err := bson.Marshal(&task)

	filter := bson.D{{"_id", existing.ID}}
	update := bson.D{{"$set", bson.Raw(by)}}
	_, err = m.collection().UpdateOne(ctx, filter, update)
	if nil != err {
		return err
	}
	return nil

}

func (m MongoDataStore) DeleteTask(ownerId int, taskId string) bool {
	existing := m.GetTask(taskId)
	if nil == existing {
		return false
	}
	if ownerId != existing.Owner {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	filter := bson.D{{"_id", existing.ID}}

	count, err := m.collection().DeleteOne(ctx, filter)
	return nil == err && count.DeletedCount > 0
}

func (m MongoDataStore) GetTask(taskId string) *model.Task {
	docId, err := primitive.ObjectIDFromHex(taskId)
	if nil != err {
		return nil
	}
	tasks, err := m.query(bson.D{{"_id", docId}}, nil)
	if nil != err || len(tasks) == 0 {
		return nil
	}
	return tasks[0]
}

func (m MongoDataStore) CountTasks(ownerId int) int {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	query := bson.D{{"owner", ownerId}}

	c, err := m.collection().CountDocuments(ctx, query, nil)
	if nil != err {
		return -1
	}
	return int(c)
}

func (m MongoDataStore) Exists(taskId string) bool {
	return m.GetTask(taskId) != nil
}

func (m MongoDataStore) Users() ([]int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	vals, err := m.collection().Distinct(ctx, "owner", bson.D{}, options.Distinct())
	if nil != err {
		return nil, err
	}

	owners := make([]int, len(vals))
	for i, val := range vals {
		ival := val.(int32)
		owners[i] = int(ival)
	}

	return owners, nil
}


func (m MongoDataStore) collection() *mongo.Collection {
	return m.db.Collection(m.collectionName)
}

func (m MongoDataStore) query(query bson.D, sort bson.D) ([]*model.Task, error) {
	findOptions := options.Find()
	findOptions.SetLimit(maxTaskCount)
	if nil != sort {
		findOptions.SetSort(sort)
	}

	var tasks []*model.Task
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	cur, err := m.collection().Find(ctx, query, findOptions)
	if nil != err {
		return nil, err
	}

	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var task model.Task

		err := cur.Decode(&task)
		if nil != err {
			return nil, err
		}

		tasks = append(tasks, &task)
	}
	return tasks, nil
}
