package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
)

var (
	MongoSession    mgo.Session
	MongoCollection *mgo.Collection
)

type syslogEntry struct {
	Id      bson.ObjectId `bson:"_id" json:"id"`
	Message string        `bson:"MESSAGE" json:"message"`
	Host    string        `bson:"HOST" json:"host_from"`
	Date    string        `bson:"DATE" json:"date"`
}

func initCollection() {
	MongoSession, err := mgo.Dial("mongodb://" + os.Getenv("MONGO_HOST") + ":" + os.Getenv("MONGO_PORT") + "/syslog")

	if err != nil {
		log.Fatal(err)
	}

	if MongoSession.Ping != nil {
		MongoSession.Refresh()
	}

	MongoCollection = MongoSession.DB("").C("messages")
}

func getDevices(c echo.Context) error {
	initCollection()

	var registros []string

	MongoCollection.Find(nil).Distinct("HOST", &registros)

	defer MongoSession.Close()

	if len(registros) != 0 {
		return c.JSON(http.StatusOK, registros)
	} else {
		message := []byte("{ \"message\": \"Device not found!\" }")
		return c.JSON(http.StatusNotFound, json.RawMessage(message))
	}

}

func getDeviceLogsAll(c echo.Context) error {

	device := c.Param("device")

	initCollection()

	var registros []syslogEntry

	MongoCollection.Find(bson.M{"HOST": device}).All(&registros)

	if len(registros) != 0 {
		return c.JSON(http.StatusOK, registros)
	} else {
		message := []byte("{ \"message\": \"Device or log not found!\" }")
		return c.JSON(http.StatusNotFound, json.RawMessage(message))
	}

}

func getDeviceLogsByDate(c echo.Context) error {
	device := c.Param("device")
	fecha := c.Param("date")

	initCollection()

	dateFormat := "2006-01-02"
	t, _ := time.Parse(dateFormat, fecha)

	var registros []syslogEntry

	// Syslog es tan hijo de puta que el campo DATE puede tener más de un espacio entre mes y día
	and_lookup := []bson.M{bson.M{"HOST": device}, bson.M{"DATE": bson.M{"$regex": "^" + t.Month().String()[:3] + "\\s+" + strconv.Itoa(t.Day()) + ".*$"}}}

	MongoCollection.Find(bson.M{"$and": and_lookup}).All(&registros)

	if len(registros) != 0 {
		return c.JSON(http.StatusOK, registros)
	} else {
		message := []byte("{ \"message\": \"Device or log not found!\" }")
		return c.JSON(http.StatusNotFound, json.RawMessage(message))
	}

}

func main() {
	e := echo.New()
	// http://$host/devices
	e.GET("/devices", getDevices)
	// http://$host/$device
	e.GET("/logs/:device", getDeviceLogsAll)
	// http://$host/$device/yyyy-mm-dd
	e.GET("/logs/:device/:date", getDeviceLogsByDate)
	e.Logger.Fatal(e.Start(":8000"))
}
