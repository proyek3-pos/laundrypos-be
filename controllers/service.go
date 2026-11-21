package controllers

import (
	"context"
	"encoding/json"
	"laundry-pos/config"
	"laundry-pos/models"
	"laundry-pos/utils"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateService(s models.Service) error {
	if err := utils.ValidateName(s.ServiceName); err != nil {
		return err
	}
	if err := utils.ValidatePrice(s.UnitPrice); err != nil {
		return err
	}
	return utils.ValidateUnit(s.Unit)
}

func CreateService(w http.ResponseWriter, r *http.Request) {
	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if err := validateService(service); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	service.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := config.ServiceCollection.InsertOne(ctx, service); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menambah layanan")
		return
	}

	utils.JSON(w, http.StatusCreated, service)
}

// GetAllServices tanpa pagination: daftar ini juga dipakai dropdown kasir.
func GetAllServices(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.ServiceCollection.Find(ctx, bson.M{}, nil)
	if err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal mendapatkan layanan")
		return
	}
	defer cursor.Close(ctx)

	var services []models.Service
	if err := cursor.All(ctx, &services); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal membaca layanan")
		return
	}

	utils.JSON(w, http.StatusOK, services)
}

func GetServiceByID(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var service models.Service
	if err := config.ServiceCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&service); err != nil {
		utils.Err(w, http.StatusNotFound, "Layanan tidak ditemukan")
		return
	}

	utils.JSON(w, http.StatusOK, service)
}

func UpdateService(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	var service models.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		utils.Err(w, http.StatusBadRequest, "Input tidak valid")
		return
	}
	if err := validateService(service); err != nil {
		utils.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{
		"serviceName": service.ServiceName,
		"description": service.Description,
		"unitPrice":   service.UnitPrice,
		"unit":        service.Unit,
	}}
	if _, err := config.ServiceCollection.UpdateOne(ctx, bson.M{"_id": id}, update); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal memperbarui layanan")
		return
	}

	service.ID = id
	utils.JSON(w, http.StatusOK, service)
}

func DeleteService(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		utils.Err(w, http.StatusBadRequest, "ID tidak valid")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := config.ServiceCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		utils.Err(w, http.StatusInternalServerError, "Gagal menghapus layanan")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]string{"message": "Layanan dihapus"})
}
