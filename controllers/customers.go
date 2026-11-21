package controllers

import (
	"context"
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/utils"
	"net/http"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func validateCustomer(c models.Customer) error {
	if err := utils.ValidateName(c.FullName); err != nil {
		return err
	}
	if err := utils.ValidatePhone(c.PhoneNumber); err != nil {
		return err
	}
	return utils.ValidateEmail(c.Email)
}

func AddCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if err := validateCustomer(customer); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := config.CustomerCollection.InsertOne(ctx, customer)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menambah customer")
		return
	}
	customer.ID = res.InsertedID.(primitive.ObjectID)
	utils.JSON(w, http.StatusCreated, customer)
}

func GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	page, limit := utils.PageParams(r)
	filter := bson.M{}
	if q := r.URL.Query().Get("q"); q != "" {
		filter["fullName"] = bson.M{"$regex": regexp.QuoteMeta(q), "$options": "i"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, err := config.CustomerCollection.CountDocuments(ctx, filter)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghitung customer")
		return
	}

	opts := options.Find().
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit)).
		SetSort(bson.M{"fullName": 1})
	cursor, err := config.CustomerCollection.Find(ctx, filter, opts)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mendapatkan customer")
		return
	}
	defer cursor.Close(ctx)

	var customers []models.Customer
	if err := cursor.All(ctx, &customers); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal membaca customer")
		return
	}

	utils.Page(w, http.StatusOK, customers, map[string]any{"page": page, "limit": limit, "total": total})
}

func GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var customer models.Customer
	if err := config.CustomerCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&customer); err != nil {
		utils.Err(w, http.StatusNotFound, "Customer tidak ditemukan")
		return
	}

	utils.JSON(w, http.StatusOK, customer)
}

func UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if err := validateCustomer(customer); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"fullName":    customer.FullName,
		"phoneNumber": customer.PhoneNumber,
		"email":       customer.Email,
	}}
	if _, err := config.CustomerCollection.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal memperbarui customer")
		return
	}

	customer.ID = id
	utils.JSON(w, http.StatusOK, customer)
}

func DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := config.CustomerCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghapus customer")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Customer dihapus"})
}
