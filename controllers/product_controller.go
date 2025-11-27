package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"backend-mobile-skill-test/config"
	"backend-mobile-skill-test/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var productCollection *mongo.Collection

func InitProductCollection() {
	dbName := os.Getenv("MONGO_DBNAME")

	if config.MongoCLient == nil {
		log.Fatal("MongoDB Client is nil, cannot initialize product collection.")
	}

	productCollection = config.MongoCLient.Database(dbName).Collection("products")
	log.Println("Product Collection initialized.")
}

func validateProduct(product models.Product) (bool, string) {
	if product.Name == "" || product.Price == 0 {
		return false, "Semua input (Name dan Price) wajib diisi."
	}

	if len(product.Name) < 5 || len(product.Name) > 50 {
		return false, "Panjang Nama Produk harus antara 5 dan 50 karakter."
	}

	if product.Price <= 0 {
		return false, "Harga produk harus lebih besar dari nol."
	}
	return true, ""
}

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body"})
		return
	}

	if ok, msg := validateProduct(product); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := productCollection.InsertOne(ctx, product)
	if err != nil {
		log.Printf("MongoDB Insert Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "product": product})
}

func GetProductList(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := productCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}
	defer cursor.Close(ctx)

	var products []models.Product
	if err = cursor.All(ctx, &products); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func GetProductByID(c *gin.Context) {
	var input struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body or missing ID"})
		return
	}

	if input.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID wajib diisi"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID format"})
		return
	}

	var product models.Product
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = productCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	if err != nil {
		log.Printf("MongoDB FindOne Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body"})
		return
	}

	if product.ID.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Product wajib di body"})
		return
	}

	if ok, msg := validateProduct(product); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: product.Name}, {Key: "price", Value: product.Price}}}}

	result, err := productCollection.UpdateOne(ctx, bson.M{"_id": product.ID}, update)
	if err != nil {
		log.Printf("MongoDB Update Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found or no changes made"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
}

func DeleteProduct(c *gin.Context) {
	var input struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input body or missing ID"})
		return
	}

	if input.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID wajib diisi"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(input.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Product ID format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := productCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		log.Printf("MongoDB Delete Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}