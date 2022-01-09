package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	gorm.Model
}

func (user *User) TableName() string {
	return "user"
}

func init() {
	dbConected()
}

func main() {

	e := echo.New()

	e.GET("/", httpHome)
	e.GET("/users", httpAllUsers)
	e.POST("/users", httpCreateUsers)
	e.DELETE("/users/:id", httpDeleteUser)

	if err := e.Start(":2000"); err != nil {
		log.Println("Not Running Server A...", err.Error())
	}

}

var users = User{}

func httpHome(c echo.Context) error {
	return c.JSON(http.StatusOK, "Welcome")
}

func httpAllUsers(c echo.Context) error {

	data := map[string]interface{}{}

	userRedis := getkeyCache("users")

	result, _ := Clientredis().Exists(context.Background(), "users").Result()

	if result != 0 {
		users.UnmarshalBinary(userRedis)
		data["redis"] = users
		return c.JSON(http.StatusOK, data)
	}

	repository := dbConected()
	err := repository.Find(&users).Error

	if err != nil {
		data["error"] = "not found users"
		return c.JSON(http.StatusNotFound, data)
	}

	data["databse"] = &users
	return c.JSON(http.StatusOK, data)
}

func httpCreateUsers(c echo.Context) error {

	if err := c.Bind(&users); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	repository := dbConected()
	err := repository.Create(&users).Error

	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	errRedis := setValueCache("users", &users)
	if errRedis != nil {
		fmt.Println(errRedis)
	}

	return c.JSON(http.StatusCreated, users)
}

func httpDeleteUser(c echo.Context) error {

	iduser := c.Param("id")

	repository := dbConected()
	err := repository.Delete(&users, iduser).Error
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	Clientredis().Del(context.Background(), "users")

	return c.JSON(http.StatusOK, "deleted")
}

func (users *User) UnmarshalBinary(data []byte) error {
	fmt.Println("UnmarshalBinary")
	return json.Unmarshal(data, users)
}

func (users *User) MarshalBinary() ([]byte, error) {
	fmt.Println("MarshalBinary")
	return json.Marshal(users)
}

func getkeyCache(key string) []byte {

	val, err := Clientredis().Get(context.Background(), key).Bytes()

	if err != nil {
		fmt.Println("key not exists", err)
	}

	return val
}

func setValueCache(key string, value interface{}) error {

	err := Clientredis().Set(context.Background(), key, value, 0).Err()

	if err != nil {
		return err
	}

	return nil
}

func dbConected() *gorm.DB {
	db, err := gorm.Open("sqlite3", "ombre.db")

	if err != nil {
		defer db.Close()
		panic(err)
	}

	db.AutoMigrate(&User{})

	return db

}

func Clientredis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return rdb
}