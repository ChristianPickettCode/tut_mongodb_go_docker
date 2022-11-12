package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

type Movie struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Title string `bson:"title,omitempty" json:"title,omitempty"`
	Plot string `bson:"plot,omitempty" json:"plot,omitempty"`
	Awards struct {
		Wins int64 `bson:"wins,omitempty" json:"wins,omitempty"`
		Nominations int64 `bson:"nominations,omitempty" json:"nominations,omitempty"`
		Text string `bson:"text,omitempty" json:"text,omitempty"`
	} `bson:"awards,omitempty" json:"awards,omitempty"`
}

func GetMoviesEndpoint(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "application/json")

	var movies []Movie

	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	
	cursor, err := collection.Find(ctx, bson.M{})

	if err != nil {
		
		response.WriteHeader(http.StatusInternalServerError)

		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		return
	}

	if err = cursor.All(ctx, &movies); err != nil {
		
		response.WriteHeader(http.StatusInternalServerError)

		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		return
	}

	json.NewEncoder(response).Encode(movies)

}

func SearchMoviesEndpoint(response http.ResponseWriter, request *http.Request) {

	response.Header().Set("content-type", "application/json")

	var movies []Movie

	queryParams := request.URL.Query()

	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	
	searchStage := bson.D{
		{ "$search", bson.D {
			{ "index", "movsearch" },
			{ "text", bson.D {
				{ "query", queryParams.Get("q")},
				{ "path", "plot" },
			}},
		}},
	}

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{searchStage})

	if err != nil {
		
		response.WriteHeader(http.StatusInternalServerError)

		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		return
	}

	if err = cursor.All(ctx, &movies); err != nil {
		
		response.WriteHeader(http.StatusInternalServerError)

		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		return
	}

	json.NewEncoder(response).Encode(movies)

}

func main() {
	
	fmt.Println("Starting the application...")

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)

	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://example:password1234@cluster0.k3nhskf.mongodb.net/?retryWrites=true&w=majority"))

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection = client.Database("sample_mflix").Collection("movies")

	router := mux.NewRouter()

	router.HandleFunc("/movies", GetMoviesEndpoint).Methods("GET")

	router.HandleFunc("/search", SearchMoviesEndpoint).Methods("GET")

	http.ListenAndServe(":12345", router)

}