package controllers

import (
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "os"
    "fmt"
    "github.com/Kamva/mgm/v2"
    "github.com/gofiber/fiber/v2"
    "github.com/parvusvox/wagetracker/models"
    "strings"
    "strconv"
    "time"
    "context"
    "log"

    "io/ioutil"
    "net/http"
)


func Database() (*mongo.Database, *mongo.Client){
    opt := options.Client().ApplyURI(os.Getenv("CONNSTRING"))
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    client, err := mongo.Connect(ctx, opt)
    if err != nil {
        log.Fatal(err.Error())
    }

    return client.Database("Job"), client
}

func CreateIndex(collectionName string, field string, unique bool) bool {
    mod := mongo.IndexModel{
        Keys : bson.M{"location": "2dsphere"},
        Options : options.Index().SetUnique(unique),
    }

    ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
    defer cancel()
    database, client := Database()
    collection := database.Collection(collectionName)
    _, err := collection.Indexes().CreateOne(ctx, mod)

    client.Disconnect(ctx)

    if( err != nil ){
        fmt.Println(err.Error())
        return false
    }
    return true
}


func GetCoords(c *fiber.Ctx) error {
    query := c.Query("q")

    response, err := http.Get("http://open.mapquestapi.com/geocoding/v1/address?key={KEY}&location=" + query)

    if err != nil {
        fmt.Println(err.Error())
        return c.Status(200).JSON(fiber.Map{
            "ok": false,
            "error": "no coords found",
        })
    }
    data, _ := ioutil.ReadAll(response.Body)
    return c.Status(200).JSON(fiber.Map{
        "ok": true,
        "data": string(data),
    })

}

func PostReport(c *fiber.Ctx) error {
    params := new(struct {
        Id string
    })
    
    params.Id = c.Query("id")

    job := &models.Job{}
    coll := mgm.Coll(job)
    err := coll.FindByID(params.Id, job)

    if err != nil {
        return c.Status(404).JSON(fiber.Map{
            "ok": false,
            "error": "job not found",
        })
    }

    job.Rating = job.Rating - 1
    err = coll.Update(job)
    if err != nil {
        return c.Status(500).JSON(fiber.Map {
            "ok": false,
            "error": err.Error(),
        })
    }
    return c.Redirect("/?lat=" + c.Query("lat")+ "&lng=" + c.Query("lng")+ "&zoom=" +c.Query("zoom"))
}

func GetJobs(c *fiber.Ctx) error {
    params := new(struct {
        Lat float64
        Lng float64
        Distance int
    })

    params.Lat, _= strconv.ParseFloat(c.Query("lat"), 64)
    params.Lng, _= strconv.ParseFloat(c.Query("lng"), 64)
    dist ,_ := strconv.ParseFloat(c.Query("distance"), 32)

    params.Distance = int(dist)

    location := models.Location {
        GeoJSONType : "Point",
        Coordinates : []float64{params.Lng, params.Lat},
    }
    jobs, err := GetJobByDistance(location, params.Distance);
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "ok": false,
            "error": err.Error(),
        })
    }
    return c.JSON(fiber.Map{
        "ok": true,
        "jobs": jobs,
    })
}

func GetJobByDistance(location models.Location, distance int) ([]models.Job, error){
    filter := bson.D {
        {"location",
        bson.D{
            {"$near", bson.D{
                {"$geometry", location},
                {"$maxDistance", distance},
            }},
        }},
    }

    var results []models.Job
    database, client := Database()
    cur, err := database.Collection("jobs").Find(context.TODO(), filter)

    client.Disconnect(context.TODO())

    if err != nil {
        fmt.Println(err.Error())
        return []models.Job{}, err
    }

    for cur.Next(context.TODO()){
        var elem models.Job
        err := cur.Decode(&elem)
        if err != nil {
            fmt.Println("Could not decode Point")
            return []models.Job {}, err
        }

        results = append(results, elem)
    }

    return results, nil
}

func PostAddJob(c *fiber.Ctx) error{
    jobName := c.FormValue("job")
    pay := c.FormValue("pay")
    location := c.FormValue("location")
    inZoom:= c.FormValue("zoom")
    sCoords := strings.Split(location, ",")

    lat, _ := strconv.ParseFloat(sCoords[0], 64)
    lng, _ := strconv.ParseFloat(sCoords[1], 64)
    wage, _ := strconv.ParseFloat(pay, 32)
    zoom, _ := strconv.ParseFloat(inZoom, 32)

    intZoom := int(zoom)

    job := models.CreateJob(
        0,
        jobName,
        wage,
        "",
        []float64{lng,lat},
        jobName,
        time.Now().String(),
    )

    err := mgm.Coll(job).Create(job)
    CreateIndex("jobs", "geoLocation", false)

    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "ok": false,
            "error": err.Error(),
        })
    }

    cookie := new(fiber.Cookie)
    cookie.Name = "contribution"
    cookie.Value = jobName
    cookie.Expires = time.Now().Add(480* time.Hour)

    c.Cookie(cookie)

    stringLat := fmt.Sprintf("%f", lat)
    stringLng := fmt.Sprintf("%f", lng)
    stringZoom := fmt.Sprintf("%d", intZoom)
    return c.Redirect("/?lat=" + stringLat + "&lng=" + stringLng + "&zoom=" + stringZoom)
}

func GetIndex(c *fiber.Ctx) error {
    hide := true

    cookieVal := c.Cookies("contribution", "NULLERR")
    if cookieVal != "NULLERR"{
        hide = false
    }

    lat,_ := strconv.ParseFloat(c.Query("lat"), 64)
    lng,_ := strconv.ParseFloat(c.Query("lng"), 64)
    zoom,_:= strconv.ParseInt(c.Query("zoom"), 10, 32)
    return c.Render("index", fiber.Map{
        "lng": lng,
        "lat": lat,
        "zoom": zoom,
        "hide":hide,
    })
}
