package main

import (
	"chat_application/api/auth"
	"chat_application/api/dal"
	"chat_application/api/errors"
	"chat_application/graph"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	db, err := dal.Connect()
	errors.CheckErr(err)
	defer db.Close()
	fmt.Println("Server started")
	router := chi.NewRouter()
	
	router.Use(auth.Middleware)
	
	// srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{},Directives: graph.DirectiveRoot{}}))
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.NewRootResolvers()))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}