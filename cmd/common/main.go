package common

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"slices"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
type BookStore struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	BookName   string
	BookAuthor string
	BookISBN   string
	BookPages  int
	BookYear   int
}

// Wraps the "Template" struct to associate a necessary method
// to determine the rendering procedure
type Template struct {
	tmpl *template.Template
}

// Preload the available templates for the view folder.
// This builds a local "database" of all available "blocks"
// to render upon request, i.e., replace the respective
// variable or expression.
// For more on templating, visit https://jinja.palletsprojects.com/en/3.0.x/templates/
// to get to know more about templating
// You can also read Golang's documentation on their templating
// https://pkg.go.dev/text/template
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Method definition of the required "Render" to be passed for the Rendering
// engine.
// Contraire to method declaration, such syntax defines methods for a given
// struct. "Interfaces" and "structs" can have methods associated with it.
// The difference lies that interfaces declare methods whether struct only
// implement them, i.e., only define them. Such differentiation is important
// for a compiler to ensure types provide implementations of such methods.
func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// Here we make sure the connection to the database is correct and initial
// configurations exists. Otherwise, we create the proper database and collection
// we will store the data.
// To ensure correct management of the collection, we create a return a
// reference to the collection to always be used. Make sure if you create other
// files, that you pass the proper value to ensure communication with the
// database
// More on what bson means: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
func prepareDatabase(client *mongo.Client, dbName string, collecName string) (*mongo.Collection, error) {
	db := client.Database(dbName)

	names, err := db.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, collecName) {
		cmd := bson.D{{"create", collecName}}
		var result bson.M
		if err = db.RunCommand(context.TODO(), cmd).Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	coll := db.Collection(collecName)
	return coll, nil
}

// Here we prepare some fictional data and we insert it into the database
// the first time we connect to it. Otherwise, we check if it already exists.
func prepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{
			BookName:   "The Vortex",
			BookAuthor: "José Eustasio Rivera",
			BookISBN:   "958-30-0804-4",
			BookPages:  292,
			BookYear:   1924,
		},
		{
			BookName:   "Frankenstein",
			BookAuthor: "Mary Shelley",
			BookISBN:   "978-3-649-64609-9",
			BookPages:  280,
			BookYear:   1818,
		},
		{
			BookName:   "The Black Cat",
			BookAuthor: "Edgar Allan Poe",
			BookISBN:   "978-3-99168-238-7",
			BookPages:  280,
			BookYear:   1843,
		},
	}

	// This syntax helps us iterate over arrays. It behaves similar to Python
	// However, range always returns a tuple: (idx, elem). You can ignore the idx
	// by using _.
	// In the topic of function returns: sadly, there is no standard on return types from function. Most functions
	// return a tuple with (res, err), but this is not granted. Some functions
	// might return a ret value that includes res and the err, others might have
	// an out parameter.
	for _, book := range startData {
		cursor, err := coll.Find(context.TODO(), book)
		var results []BookStore
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		if len(results) > 1 {
			log.Fatal("more records were found")
		} else if len(results) == 0 {
			result, err := coll.InsertOne(context.TODO(), book)
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", result)
			}

		} else {
			for _, res := range results {
				cursor.Decode(&res)
				fmt.Printf("%+v\n", res)
			}
		}
	}
}

// Generic method to perform "SELECT * FROM BOOKS" (if this was SQL, which
// it is not :D ), and then we convert it into an array of map. In Golang, you
// define a map by writing map[<key type>]<value type>{<key>:<value>}.
// interface{} is a special type in Golang, basically a wildcard...
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"id":     res.ID.Hex(),
			"name":   res.BookName,
			"author": res.BookAuthor,
			"isbn":   res.BookISBN,
			"pages":  res.BookPages,
			"year":   res.BookYear,
		})
	}

	return ret
}
