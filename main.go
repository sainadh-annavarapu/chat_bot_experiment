package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	initDB()
	router := mux.NewRouter()

	router.HandleFunc("/verify_coupons/", VerifyCoupons).Methods("POST")
	router.HandleFunc("/add_user/", AddUser).Methods("POST")
	router.HandleFunc("/get_money/", GetMoney).Methods("POST")

	c := cors.AllowAll()

	handler := c.Handler(router)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
