package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type CouponRequest struct {
	Coupons string `json:"coupons"`
}

type Coupon struct {
	CouponCode  string `gorm:"column:coupon_code" json:"coupon_code"`
	CouponValue int    `gorm:"column:coupon_value" json:"coupon_value"`
}

type User struct {
	Id      int    `gorm:"column:id"`
	Name    string `gorm:"column:name" json:"name"`
	City    string `gorm:"column:city" json:"city"`
	PhoneNo int    `gorm:"column:phnno" json:"phnno"`
}

type UserCoupon struct {
	CouponCode string `gorm:"column:coupon_code" json:"coupon_code"`
	UserId     int    `gorm:"column:user_id"`
	IsUsed     bool   `gorm:"column:is_used" json:"is_used"`
}

func VerifyCoupons(w http.ResponseWriter, r *http.Request) {
	var couponRequest CouponRequest
	if err := json.NewDecoder(r.Body).Decode(&couponRequest); err != nil {
		http.Error(w, "Invalid JSON Format", http.StatusBadRequest)
		return
	}

	fmt.Println(couponRequest)

	phnNoStr := r.Header.Get("phnno")
	fmt.Println("phnNoStr:", phnNoStr)

	if phnNoStr == "" {
		http.Error(w, "Missing phnno in query parameters", http.StatusBadRequest)
		return
	}

	phnNo := parsePhoneNumber(phnNoStr)
	if phnNo == 0 {
		http.Error(w, "Invalid phnno in query parameters", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Table("users2").Where("phnno = ?", phnNo).First(&user).Error; err != nil {
		http.Error(w, "PhnNo doesn't exist in our database", http.StatusBadRequest)
		return
	}

	couponCodes := strings.Fields(couponRequest.Coupons)

	var successCount int
	var validCoupons []string

	for _, code := range couponCodes {
		coupon := Coupon{CouponCode: code}

		err := db.Table("coupons").Where("coupon_code = ?", coupon.CouponCode).First(&coupon).Error

		// Failure
		// if err == gorm.ErrRecordNotFound {
		// 	continue // Move to the next coupon if the current one is invalid
		// } else if err != nil {
		// 	jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
		// 	return
		// }

		if err == gorm.ErrRecordNotFound {
			continue
		}
		// Success
		status := AddUserCoupon(coupon, user)
		if !status {
			successCount++
			validCoupons = append(validCoupons, code)
		}
	}

	// validCouponsString := strings.Join(validCoupons, ",")

	// Respond with the count of successfully verified coupons, total number of coupons, and the list of valid coupons
	responseData := map[string]interface{}{
		"success_count": successCount,
		"total_coupons": len(couponCodes),
		"valid_coupons": validCoupons,
	}

	// Send the response with the success count, total number of coupons, and the list of valid coupons
	jsonResponse(w, http.StatusOK, responseData)
}

func parsePhoneNumber(phnNoStr string) int {
	phnNo, err := strconv.Atoi(phnNoStr)
	if err != nil {
		return 0
	}
	return phnNo
}

func AddUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	fmt.Println(user)

	err := db.Table("users2").Where("phnno = ?", user.PhoneNo).First(&user).Error

	if err == nil {
		str := "User is already present"
		jsonResponse(w, http.StatusOK, str)
		return
	}

	if err = db.Table("users2").Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, http.StatusOK, "User Created Successfully")
}

func AddUserCoupon(coupon Coupon, user User) bool {
	usercoupon := UserCoupon{CouponCode: coupon.CouponCode, UserId: user.Id}
	fmt.Println(usercoupon)

	if err := db.Table("user_coupon").Where("coupon_code = ? AND user_id = ?", usercoupon.CouponCode, usercoupon.UserId).First(&usercoupon).Error; err == nil {
		status := usercoupon.IsUsed
		fmt.Println("Status is: ", status)
		return status
	}
	if err := db.Table("user_coupon").Create(&usercoupon).Error; err != nil {
		fmt.Println("User Coupon Created Successfully")
	}

	return false
}

func GetMoney(w http.ResponseWriter, r *http.Request) {
	var couponRequest struct {
		Coupons []string `json:"coupons"`
	}
	phnNoStr := r.Header.Get("phnno")
	if phnNoStr == "" {
		http.Error(w, "Missing phnno in query parameters", http.StatusBadRequest)
		return
	}

	phnNo, err := strconv.Atoi(phnNoStr)
	if err != nil {
		http.Error(w, "Invalid phnno in query parameters", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Table("users2").Where("phnno = ?", phnNo).First(&user).Error; err != nil {
		http.Error(w, "PhnNo doesnt exist in our database", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&couponRequest); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON: %s", err), http.StatusBadRequest)
		return
	}
	fmt.Println(couponRequest)

	sum, flag := TotalMoney(couponRequest.Coupons)
	if SetIsUsedToTrue(user.Id, couponRequest.Coupons) && flag {
		jsonResponse4(w, http.StatusOK, sum)
		return
	}
	http.Error(w, "Failed to fetch the coupons", http.StatusInternalServerError)
}

func SetIsUsedToTrue(userId int, couponCodes []string) bool {
	for _, code := range couponCodes {
		err := db.Table("user_coupon").Where("coupon_code = ? AND user_id = ?", code, userId).Update("is_used", true).Error
		if err != nil {
			return false
		}
	}
	return true
}

func TotalMoney(couponCodes []string) (int, bool) {
	sum := 0
	for _, code := range couponCodes {
		var coupon Coupon
		err := db.Table("coupons").Where("coupon_code = ?", code).First(&coupon).Error
		if err != nil {
			return 0, false
		}
		sum += coupon.CouponValue
	}
	return sum, true
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// func jsonResponse2(w http.ResponseWriter, statusCode int, data interface{}, status bool) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	response := map[string]interface{}{
// 		"data":    data,
// 		"is_used": status,
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

// func jsonResponse3(w http.ResponseWriter, statusCode int, data interface{}, sum int) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)
// 	response := map[string]interface{}{
// 		"data": data,
// 		"sum":  sum,
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

func jsonResponse4(w http.ResponseWriter, statusCode int, sum int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"sum": sum,
	}

	json.NewEncoder(w).Encode(response)
}
