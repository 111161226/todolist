package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.TaskList)
	engine.GET("/task/:id", service.ShowTask) // ":id" is a parameter
	//add new task
	engine.GET("/task/new", service.NewTaskForm)
	engine.POST("/task/new", service.RegisterTask)
	//edit the task
	engine.GET("/task/edit/:id", service.EditTaskForm)
	engine.POST("/task/edit/:id", service.UpdateTask)
	//delete the task
	engine.GET("/task/delete/:id", service.DeleteTask)

	//register user
    engine.GET("/user/new", service.NewUserForm)
    engine.POST("/user/new", service.RegisterUser)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
