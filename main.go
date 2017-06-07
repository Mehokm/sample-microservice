package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/Mehokm/go-tfts"
	"github.com/spf13/viper"
)

// User struct holds basic data about a user
type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var users = []User{
	User{"John", "Smith"},
	User{"Jane", "Doe"},
	User{"Bruce", "Wayne"},
}

// This is a stub db struct
type stubDB struct{}

func (s stubDB) execute(query string) {
	fmt.Println("I executed: " + query)
}

// UserController is a basic struct to encapsulate all user actions
type UserController struct {
	db stubDB
}

// GetUser action maps to route /users/{id:i}
func (uc UserController) GetUser(c rest.Context) rest.ResponseSender {
	return rest.NewOKJSONResponse(User{"John", "Smith"})
}

// GetAllUsers action maps to route /users (GET)
func (uc UserController) GetAllUsers(c rest.Context) rest.ResponseSender {
	return rest.NewOKJSONResponse(users)
}

// CreateUser action maps to route /users (POST)
func (uc UserController) CreateUser(c rest.Context) rest.ResponseSender {
	var user User

	c.BindJSONEntity(&user)

	users = append(users, user)

	uc.db.execute(fmt.Sprintf("INSERT INTO User (`firstName`, `lastName`) VALUES ('%v', '%v')", user.FirstName, user.LastName))

	return rest.NewCreatedJSONResponse(user)
}

func main() {
	// example viper config with ENV vars
	viper.SetConfigName("sample-conf")
	viper.SetEnvPrefix("spl")

	viper.AutomaticEnv()

	fmt.Println(fmt.Sprintf("Viper: %v", viper.Get("MYSQL_USER")))

	port := viper.GetString("PORT")
	if port == "" {
		port = "5000"
	}

	root := "/api/v1"

	userController := UserController{stubDB{}}

	router := rest.DefaultRouter().Prefix(root).RouteMap(
		rest.NewRoute().For("/users/{id:i}").
			With(rest.MethodGET, userController.GetUser),
		rest.NewRoute().For("/users").
			With(rest.MethodGET, userController.GetAllUsers).
			And(rest.MethodPOST, userController.CreateUser),
	)

	health := rest.Router("admin").RouteMap(
		rest.NewRoute().For("/health").With(rest.MethodGET, func(c rest.Context) rest.ResponseSender {

			return rest.NewOKJSONResponse("status ok")
		}),
		rest.NewRoute().For("/fun").With(rest.MethodGET, func(c rest.Context) rest.ResponseSender {
			cmd := exec.Command("sh", "-c", "basename \"$(cat /proc/1/cpuset)\"")

			var out bytes.Buffer
			cmd.Stdout = &out

			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}

			return rest.NewOKJSONResponse(fmt.Sprintf("Hello from container id: %v", out.String()))
		}),
		rest.NewRoute().For("/env").With(rest.MethodGET, func(c rest.Context) rest.ResponseSender {

			return rest.NewOKJSONResponse(os.Environ())
		}),
	)

	mainHandler := rest.NewHandler(router)

	healthHandler := rest.NewHandler(health)

	http.Handle("/", healthHandler)
	http.Handle(root+"/", mainHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
