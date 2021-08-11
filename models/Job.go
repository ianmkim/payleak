package models

import (
    //"context"
    //"fmt"
    //"github.com/google/uuid"
    //"go.mongodb.org/mongo-driver/bson"
    "github.com/Kamva/mgm/v2"
)

type Job struct {
    mgm.DefaultModel `bson:",inline"`
    IntId int `json:"intId" bson:"intId"`
    JobName string `json:"jobName" bson:"jobName"`
    Wage float64 `json:"wage" bson:"wage"`
    StringLocation string `json:"stringLocation" bson:"stringLocation"`
    Location Location `json:"location" bson:"location"`
    Employer string `json:"employer" bson:"employer"`
    Time string `json:"time" bson:"time"`

    Rating int `json:"rating" bson:"rating"`
}

type Location struct {
    GeoJSONType string `json:"type" bson:"type"`
    Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

func CreateJob(intId int, jobName string, wage float64, location string, coordinates []float64, employer, time string) *Job{
    gjson := Location{
        GeoJSONType : "Point",
        Coordinates : coordinates,
    }
    return &Job{
        IntId : intId,
        JobName : jobName,
        Wage : wage,
        Location :gjson,
        StringLocation: location,
        Employer : employer,
        Time : time,
        Rating : 0,
    }
}

