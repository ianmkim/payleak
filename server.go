package main

import(
    "log"
    "os"
    "strings"

    "github.com/Kamva/mgm/v2"
    "github.com/gofiber/fiber/v2"
    "github.com/parvusvox/wagetracker/controllers"

    "github.com/gofiber/template/django"

    "go.mongodb.org/mongo-driver/mongo/options"
)

func init(){
    connectionString := os.Getenv("CONNSTRING")
    err := mgm.SetDefaultConfig(nil, "Job", options.Client().ApplyURI(connectionString))
    if err != nil{
        log.Fatal(err)
    }
}

func main(){
    engine := django.New("./views", ".html")

    engine.AddFunc("replaceSpaces", func(name string) string {
        return strings.ReplaceAll(name, " ", "_")
    })

    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Static("/static", "./static")

    app.Get("/", controllers.GetIndex)
    app.Get("/get_jobs", controllers.GetJobs)
    app.Get("/get_coords", controllers.GetCoords)
    app.Get("/report", controllers.PostReport)

    app.Post("/add_job", controllers.PostAddJob)

    port := os.Getenv("PORT")
    if port == ""{
        port = ":3000"
    } else {
        port = ":" + port
    }

    log.Fatal(app.Listen(port))

}

