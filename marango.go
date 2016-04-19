// Package Marango provides an intuitive ODM (Object Document Model) library for working
// with MongoDB documents.
// It builds on top of the awesome mgo library
package marango

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
)

//Convenient access to bson.M
type M bson.M

//Convenient access to bson.D
type D bson.D

type Marango struct {
	Db        *mgo.Database
	documents map[string]Document
	models    map[string]*Model
	modelTag  string
}

// New returns a new intance of the Marango type
func New(s *mgo.Session, db string) *Marango {
	marango := &Marango{Db: s.DB(db), modelTag: "model"}
	marango.documents = make(map[string]Document)
	marango.models = make(map[string]*Model)
	return marango
}

// SetModelTag changes the default tag key of `model` to an arbitrary key.
// This value is read to make relationships for populting based on ObjectIds
func (z *Marango) SetModelTag(key string) {
	z.modelTag = key
}

// Register registers a given schema and its corresponding collection name with Marango.
// All schemas MUST be registered using this function.
// Function will return a pointer to the Marango.Model value for this model
func (z *Marango) Register(schema interface{}, collectionName string) *Model {
	typ := reflect.TypeOf(schema)
	structName := typ.Name()
	if typ.Kind() == reflect.Ptr {
		panic("Expected value, got a pointer")
	}

	idField := reflect.ValueOf(schema).FieldByName("Id")
	if !idField.IsValid() {
		panic("Schema `" + structName + "` must have an `Id` field")
	}

	model := newModel(z.Db.C(collectionName), z)
	z.models[structName] = model

	z.documents[structName] = Document{C: z.Db.C(collectionName),
		isQueried: true, schemaStruct: schema, Model: model,
		populated: make(map[string]interface{}), Found: true}

	return model
}

// CreateDoc conditions an instance of the model to become a document. Will create an ObjectId for the document.
//
// See Model.CreateDoc. They are the same
func (z *Marango) CreateDoc(doc interface{}) {
	typ := reflect.TypeOf(doc).Elem()
	structName := typ.Name()
	document := z.documents[structName]

	document.schema = doc
	document.Model = z.models[structName]
	document.Virtual = newVirtual()

	val := reflect.ValueOf(doc).Elem()
	docVal := val.FieldByName("Document")
	docVal.Set(reflect.ValueOf(document))

	idField := reflect.ValueOf(doc).Elem().FieldByName("Id")
	id := bson.NewObjectId()
	idField.Set(reflect.ValueOf(id))
}

// C gives access to the underlying *mgo.Collection value for a model.
// The model name is case sensitive.
func (z *Marango) C(model string) (*mgo.Collection, bool) {
	m, ok := z.documents[model]
	c := m.C
	return c, ok
}

// Model returns a pointer to the Model of the registered schema
func (z *Marango) Model(name string) *Model {
	return z.models[name]
}

// See ObjectId
func (z *Marango) ObjectId(id string) bson.ObjectId {
	return ObjectId(id)
}

// ObjectId converts a string hex representation of an ObjectId into type bson.ObjectId.
func ObjectId(id string) bson.ObjectId {
	return bson.ObjectIdHex(id)
}

//Function will take types string or bson.ObjectId represented by a type interface{} and returns
//a type bson.ObjectId. Will panic if wrong type is passed. Will also panic if the string
//is not a valid representation of an ObjectId
func getObjectId(id interface{}) bson.ObjectId {
	var idActual bson.ObjectId
	switch id.(type) {
	case string:
		idActual = bson.ObjectIdHex(id.(string))
		break
	case bson.ObjectId:
		idActual = id.(bson.ObjectId)
	default:
		panic("Only accepts types `string` and `bson.ObjectId` accepted as Id")
	}
	return idActual
}
