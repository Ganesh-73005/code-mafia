package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// MongoDB connection details
	mongoURI := ""
	dbName := "code_mafia"

	// Admin credentials
	adminUsername := "admin"
	adminPassword := "admin123" // Change this to a secure password
	adminTeamName := "Admin Team"

	if len(os.Args) > 1 {
		adminUsername = os.Args[1]
	}
	if len(os.Args) > 2 {
		adminPassword = os.Args[2]
	}

	// Connect to MongoDB
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Ping the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB successfully")

	db := client.Database(dbName)

	// Check if admin user already exists
	usersCollection := db.Collection("users")
	var existingUser bson.M
	err = usersCollection.FindOne(ctx, bson.M{"username": adminUsername}).Decode(&existingUser)
	if err == nil {
		log.Fatalf("User '%s' already exists", adminUsername)
	}

	// Create admin team
	teamsCollection := db.Collection("teams")
	var team bson.M
	err = teamsCollection.FindOne(ctx, bson.M{"name": adminTeamName}).Decode(&team)
	
	var teamID bson.ObjectID
	if err == mongo.ErrNoDocuments {
		// Create new team
		teamDoc := bson.M{
			"name":       adminTeamName,
			"points":     0,
			"coins":      1000,
			"created_at": time.Now(),
		}
		result, err := teamsCollection.InsertOne(ctx, teamDoc)
		if err != nil {
			log.Fatalf("Failed to create team: %v", err)
		}
		teamID = result.InsertedID.(bson.ObjectID)
		fmt.Printf("Created team: %s (ID: %s)\n", adminTeamName, teamID.Hex())
	} else if err != nil {
		log.Fatalf("Error checking team: %v", err)
	} else {
		teamID = team["_id"].(bson.ObjectID)
		fmt.Printf("Using existing team: %s (ID: %s)\n", adminTeamName, teamID.Hex())
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create admin user
	userDoc := bson.M{
		"username":   adminUsername,
		"password":   string(hashedPassword),
		"team_id":    teamID.Hex(),
		"role":       "admin",
		"created_at": time.Now(),
	}

	result, err := usersCollection.InsertOne(ctx, userDoc)
	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("\n✅ Admin user created successfully!\n")
	fmt.Printf("   Username: %s\n", adminUsername)
	fmt.Printf("   Password: %s\n", adminPassword)
	fmt.Printf("   User ID: %s\n", result.InsertedID.(bson.ObjectID).Hex())
	fmt.Printf("   Team: %s\n", adminTeamName)
	fmt.Printf("   Role: admin\n\n")
	fmt.Printf("You can now login at: http://localhost:3000/admin/login\n")
}

