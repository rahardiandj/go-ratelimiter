package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

var conn redis.Conn

func main() {
	fmt.Println("Bonjour!!")

	router := mux.NewRouter()

	router.HandleFunc("/user/{id}", getUser)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
	}

	pool := newPool()
	conn = pool.Get()
	defer conn.Close()

	err := ping(conn)

	// err = set(conn, "mykey", "1")
	// _, err = setEx(conn, "mykeyEx", "1", 60)
	// counter, err := incr(conn, "counter")

	// fmt.Printf("counter : %v\n", counter)

	if err != nil {
		fmt.Printf("Got error : %v\n", err)
	}

	log.Fatal(srv.ListenAndServe())

}

const MAX_LIMIT int = 5
const MAX_DURATION int = 60

func getUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	strUserId := params["id"]

	userId, err := strconv.Atoi(strUserId)

	if err != nil {
		fmt.Printf("Param error\n")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if isExceedLimit(conn, userId) {
		fmt.Printf("%v is over limit\n", userId)
		w.Write([]byte("Reuest is exceed limit"))
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write([]byte("OK"))
		w.WriteHeader(http.StatusOK)
	}

}

func isExceedLimit(c redis.Conn, userId int) bool {
	strUserId := strconv.Itoa(userId)
	strCounter, err := get(c, strUserId)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	counter, err := strconv.Atoi(strCounter)

	if err != nil {
		fmt.Println(err)
	}

	counter, err = incr(c, strUserId)

	if strCounter == "" {
		err := expire(c, strUserId, MAX_DURATION)
		if err != nil {
			fmt.Println(err)
		}
	}

	if err != nil || counter > MAX_LIMIT {

		return true
	}

	return false
}

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 120000,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

func ping(c redis.Conn) error {
	pong, err := c.Do("PING")

	if err != nil {
		return err
	}

	str, err := redis.String(pong, err)

	if err != nil {
		return err
	}

	fmt.Printf("Got message: %v\n", str)

	return nil
}

func set(c redis.Conn, key, value string) error {
	pong, err := c.Do("SET", key, value)

	if err != nil {
		return err
	}

	str, err := redis.String(pong, err)

	if err != nil {
		return err
	}

	fmt.Printf("Got message: %v\n", str)

	return nil
}

func setEx(c redis.Conn, key, value string, expired int) (string, error) {
	str, err := redis.String(c.Do("SETEX", key, expired, value))

	if err != nil {
		return "", err
	}

	fmt.Printf("Got message: %v\n", str)

	return str, nil
}

func get(c redis.Conn, key string) (string, error) {
	str, err := redis.String(c.Do("GET", key))

	if err != nil {
		return "", err
	}

	fmt.Printf("Got value: %v\n", str)

	return str, nil
}

func expire(c redis.Conn, key string, expired int) error {
	_, err := c.Do("EXPIRE", key, expired)

	if err != nil {
		return err
	}

	return nil
}

func incr(c redis.Conn, key string) (int, error) {
	count, err := redis.Int(c.Do("INCR", key))

	if err != nil {
		return 0, err
	}

	return count, nil
}

// const TCP_TRANSPORT = "tcp"

// func initRedis() {
// 	New(":6379")
// }

// type Redis struct {
// 	Pool  *redis.Pool
// 	mutex sync.Mutex
// }

// func New(address string) *Redis {
// 	return &Redis{
// 		Pool: &redis.Pool{
// 			MaxIdle:     100,
// 			MaxActive:   100,
// 			IdleTimeout: 100 * time.Second,
// 			Dial: func() (redis.Conn, error) {
// 				return redis.Dial(TCP_TRANSPORT, address)
// 			},
// 		},
// 	}
// }
