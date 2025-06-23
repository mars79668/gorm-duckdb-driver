package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/duckdb"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User model
type User struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:255;uniqueIndex"`
	Age       uint8
	Birthday  *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	Posts     []Post `gorm:"foreignKey:UserID"`
}

// Post model with foreign key relationship
type Post struct {
	ID          uint   `gorm:"primarykey"`
	Title       string `gorm:"size:200;not null"`
	Content     string `gorm:"type:text"`
	UserID      uint   `gorm:"not null;index"`
	User        User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func main() {
	// Connect to DuckDB (in-memory for this example)
	db, err := gorm.Open(duckdb.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&User{}, &Post{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("✅ Database migration completed successfully!")

	// Create sample data
	now := time.Now()
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)

	users := []User{
		{
			Name:     "John Doe",
			Email:    "john@example.com",
			Age:      33,
			Birthday: &birthday,
		},
		{
			Name:  "Jane Smith",
			Email: "jane@example.com",
			Age:   28,
		},
		{
			Name:  "Bob Johnson",
			Email: "bob@example.com",
			Age:   45,
		},
	}

	// Create users
	result := db.Create(&users)
	if result.Error != nil {
		log.Fatal("Failed to create users:", result.Error)
	}
	fmt.Printf("✅ Created %d users\n", len(users))

	// Create posts for the first user
	posts := []Post{
		{
			Title:       "My First Post",
			Content:     "This is the content of my first post. It's quite exciting!",
			UserID:      users[0].ID,
			PublishedAt: &now,
		},
		{
			Title:   "Draft Post",
			Content: "This is a draft post that hasn't been published yet.",
			UserID:  users[0].ID,
		},
		{
			Title:       "DuckDB is Amazing",
			Content:     "I'm learning about DuckDB and it's really powerful for analytics!",
			UserID:      users[1].ID,
			PublishedAt: &now,
		},
	}

	result = db.Create(&posts)
	if result.Error != nil {
		log.Fatal("Failed to create posts:", result.Error)
	}
	fmt.Printf("✅ Created %d posts\n", len(posts))

	// Query examples
	fmt.Println("\n📊 Query Examples:")

	// 1. Find all users
	var allUsers []User
	db.Find(&allUsers)
	fmt.Printf("Total users: %d\n", len(allUsers))

	// 2. Find user by email
	var user User
	db.Where("email = ?", "john@example.com").First(&user)
	fmt.Printf("Found user: %s (ID: %d)\n", user.Name, user.ID)

	// 3. Find users older than 30
	var olderUsers []User
	db.Where("age > ?", 30).Find(&olderUsers)
	fmt.Printf("Users older than 30: %d\n", len(olderUsers))

	// 4. Find posts with user information (Join)
	var postsWithUsers []Post
	db.Preload("User").Find(&postsWithUsers)
	fmt.Printf("Posts with user info: %d\n", len(postsWithUsers))
	for _, post := range postsWithUsers {
		status := "Draft"
		if post.PublishedAt != nil {
			status = "Published"
		}
		fmt.Printf("  - '%s' by %s (%s)\n", post.Title, post.User.Name, status)
	}

	// 5. Count posts per user
	type UserPostCount struct {
		Name      string
		PostCount int64
	}
	var userPostCounts []UserPostCount
	db.Table("users").
		Select("users.name, COUNT(posts.id) as post_count").
		Joins("LEFT JOIN posts ON posts.user_id = users.id").
		Group("users.id, users.name").
		Scan(&userPostCounts)

	fmt.Println("\nPost counts per user:")
	for _, upc := range userPostCounts {
		fmt.Printf("  - %s: %d posts\n", upc.Name, upc.PostCount)
	}

	// 6. Find published posts only
	var publishedPosts []Post
	db.Where("published_at IS NOT NULL").Preload("User").Find(&publishedPosts)
	fmt.Printf("\nPublished posts: %d\n", len(publishedPosts))
	for _, post := range publishedPosts {
		fmt.Printf("  - '%s' by %s\n", post.Title, post.User.Name)
	}

	// Update example
	fmt.Println("\n🔄 Update Example:")
	db.Model(&user).Update("age", 34)
	fmt.Printf("Updated %s's age to 34\n", user.Name)

	// Transaction example
	fmt.Println("\n💳 Transaction Example:")
	err = db.Transaction(func(tx *gorm.DB) error {
		// Create a new user and post in a transaction
		newUser := User{
			Name:  "Alice Wilson",
			Email: "alice@example.com",
			Age:   25,
		}
		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		newPost := Post{
			Title:   "My Introduction Post",
			Content: "Hello everyone! I'm new here.",
			UserID:  newUser.ID,
		}
		if err := tx.Create(&newPost).Error; err != nil {
			return err
		}

		fmt.Printf("✅ Created new user '%s' and their first post in transaction\n", newUser.Name)
		return nil
	})

	if err != nil {
		log.Fatal("Transaction failed:", err)
	}

	// Raw SQL example
	fmt.Println("\n🔍 Raw SQL Example:")
	var avgAge float64
	db.Raw("SELECT AVG(age) FROM users").Scan(&avgAge)
	fmt.Printf("Average user age: %.1f\n", avgAge)

	// Migration info
	fmt.Println("\n🔧 Migration Info:")
	fmt.Printf("Has users table: %v\n", db.Migrator().HasTable(&User{}))
	fmt.Printf("Has posts table: %v\n", db.Migrator().HasTable(&Post{}))
	fmt.Printf("Has email index: %v\n", db.Migrator().HasIndex(&User{}, "idx_users_email"))

	columnTypes, _ := db.Migrator().ColumnTypes(&User{})
	fmt.Printf("User table has %d columns\n", len(columnTypes))

	fmt.Println("\n🎉 DuckDB GORM Driver Demo Complete!")
}
