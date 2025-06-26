package main

import (
	"fmt"
	"log"
	"time"

	duckdb "gorm.io/driver/duckdb"
	"gorm.io/gorm"
)

// User model demonstrating basic GORM features
type User struct {
	ID        uint               `gorm:"primaryKey" json:"id"`
	Name      string             `gorm:"size:100;not null" json:"name"`
	Email     string             `gorm:"size:255;uniqueIndex" json:"email"`
	Age       uint8              `json:"age"`
	Birthday  time.Time          `json:"birthday"`
	CreatedAt time.Time          `gorm:"autoCreateTime:false" json:"created_at"`
	UpdatedAt time.Time          `gorm:"autoUpdateTime:false" json:"updated_at"`
	Posts     []Post             `gorm:"foreignKey:UserID" json:"posts"`
	Tags      duckdb.StringArray `json:"tags"` // Now using proper array type!
}

// Post model demonstrating relationships
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"` // Remove autoIncrement
	Title     string    `gorm:"size:200;not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
	Tags      []Tag     `gorm:"many2many:post_tags;" json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tag model demonstrating many-to-many relationships
type Tag struct {
	ID    uint   `gorm:"primaryKey" json:"id"` // Remove autoIncrement
	Name  string `gorm:"size:50;uniqueIndex" json:"name"`
	Posts []Post `gorm:"many2many:post_tags;" json:"posts"`
}

// Product model demonstrating basic features
type Product struct {
	ID          uint               `gorm:"primaryKey" json:"id"` // Remove autoIncrement
	Name        string             `gorm:"size:100;not null" json:"name"`
	Price       float64            `json:"price"`
	Description string             `json:"description"`
	Categories  duckdb.StringArray `json:"categories"`  // Array support
	Scores      duckdb.FloatArray  `json:"scores"`      // Float array support
	ViewCounts  duckdb.IntArray    `json:"view_counts"` // Int array support
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func main() {
	fmt.Println("🦆 GORM DuckDB Driver Example with Array Support")
	fmt.Println("=================================================")

	// Initialize database
	db, err := gorm.Open(duckdb.Open("example.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to DuckDB")

	// Migrate the schema
	fmt.Println("🔧 Auto-migrating database schema...")
	err = db.AutoMigrate(&User{}, &Post{}, &Tag{}, &Product{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("✅ Schema migration completed")

	// Demonstrate basic CRUD operations
	demonstrateBasicCRUD(db)

	// Demonstrate array features
	demonstrateArrayFeatures(db)

	// Demonstrate relationships
	demonstrateRelationships(db)

	// Demonstrate DuckDB-specific features
	demonstrateDuckDBFeatures(db)

	// Demonstrate advanced queries
	demonstrateAdvancedQueries(db)

	fmt.Println("\n🎉 Example completed successfully!")
}

// Add helper function to get next ID
func getNextID(db *gorm.DB, tableName string) uint {
	var maxID uint
	db.Raw(fmt.Sprintf("SELECT COALESCE(MAX(id), 0) FROM %s", tableName)).Scan(&maxID)
	return maxID + 1
}

func demonstrateBasicCRUD(db *gorm.DB) {
	fmt.Println("\n📝 Basic CRUD Operations")
	fmt.Println("------------------------")

	// Get the starting ID for users
	nextUserID := getNextID(db, "users")

	// Create sample users with manual timestamps
	now := time.Now()
	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	users := []User{
		{
			ID:        nextUserID,
			Name:      "Alice Johnson",
			Email:     "alice@example.com",
			Age:       25,
			Birthday:  birthday,
			CreatedAt: now,
			UpdatedAt: now,
			Tags:      duckdb.StringArray{"developer", "go-enthusiast"}, // Now working!
		},
		{
			ID:        nextUserID + 1,
			Name:      "Bob Smith",
			Email:     "bob@example.com",
			Age:       30,
			Birthday:  time.Time{}, // Zero time for no birthday
			CreatedAt: now,
			UpdatedAt: now,
			Tags:      duckdb.StringArray{"manager", "tech-lead"}, // Now working!
		},
		{
			ID:        nextUserID + 2,
			Name:      "Charlie Brown",
			Email:     "charlie@example.com",
			Age:       35,
			Birthday:  time.Time{}, // Zero time for no birthday
			CreatedAt: now,
			UpdatedAt: now,
			Tags:      duckdb.StringArray{"analyst", "data-science"}, // Now working!
		},
	}

	// Create all users
	result := db.Create(&users)
	if result.Error != nil {
		log.Printf("Error creating users: %v", result.Error)
		return
	}
	fmt.Printf("✅ Created %d users\n", result.RowsAffected)

	// Read operations
	var allUsers []User
	db.Find(&allUsers)
	fmt.Printf("👥 Found %d users in database\n", len(allUsers))

	// Show users with their tags
	for _, user := range allUsers {
		if len(user.Tags) > 0 {
			fmt.Printf("🏷️  %s has tags: %v\n", user.Name, []string(user.Tags))
		}
	}

	// Array querying example (basic substring search)
	var developersWithArrays []User
	// Note: DuckDB array syntax might vary, this is a basic example
	result = db.Where("array_to_string(tags, ',') LIKE ?", "%developer%").Find(&developersWithArrays)
	if result.Error == nil && len(developersWithArrays) > 0 {
		fmt.Printf("🔍 Found %d users with 'developer' in tags\n", len(developersWithArrays))
	}

	// Update operation
	db.Model(&users[0]).Update("age", 26)
	fmt.Printf("✏️ Updated user: %s\n", users[0].Name)

	// Delete operation
	db.Delete(&users[2])
	fmt.Printf("🗑️ Deleted user: %s\n", users[2].Name)
}

// Add this new function to demonstrate array features:
func demonstrateArrayFeatures(db *gorm.DB) {
	fmt.Println("\n🎨 Array Features Demonstration")
	fmt.Println("-------------------------------")

	// Get the starting ID for products
	nextProductID := getNextID(db, "products")

	// Create products with arrays
	now := time.Now()
	products := []Product{
		{
			ID:          nextProductID,
			Name:        "Analytics Software",
			Price:       299.99,
			Description: "Advanced data analytics platform",
			Categories:  duckdb.StringArray{"software", "analytics", "business"},
			Scores:      duckdb.FloatArray{4.5, 4.8, 4.2, 4.9},
			ViewCounts:  duckdb.IntArray{1250, 890, 2340, 567},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          nextProductID + 1,
			Name:        "Gaming Laptop",
			Price:       1299.99,
			Description: "High-performance gaming laptop",
			Categories:  duckdb.StringArray{"electronics", "computers", "gaming"},
			Scores:      duckdb.FloatArray{4.7, 4.9, 4.6},
			ViewCounts:  duckdb.IntArray{3200, 2100, 4500},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	result := db.Create(&products)
	if result.Error != nil {
		log.Printf("Error creating products with arrays: %v", result.Error)
		return
	}
	fmt.Printf("✅ Created %d products with arrays\n", result.RowsAffected)

	// Retrieve and display arrays
	var retrievedProducts []Product
	db.Find(&retrievedProducts)

	for _, product := range retrievedProducts {
		fmt.Printf("📦 Product: %s\n", product.Name)
		fmt.Printf("   Categories: %v\n", []string(product.Categories))
		fmt.Printf("   Scores: %v\n", []float64(product.Scores))
		fmt.Printf("   View Counts: %v\n", []int64(product.ViewCounts))
	}

	// Update arrays
	if len(retrievedProducts) > 0 {
		product := &retrievedProducts[0]
		product.Categories = append(product.Categories, "premium")
		product.Scores = append(product.Scores, 5.0)
		product.ViewCounts = append(product.ViewCounts, 1000)

		result = db.Save(product)
		if result.Error != nil {
			log.Printf("Error updating product arrays: %v", result.Error)
		} else {
			fmt.Printf("✅ Updated arrays for product: %s\n", product.Name)
			fmt.Printf("   New categories: %v\n", []string(product.Categories))
		}
	}
}

func demonstrateRelationships(db *gorm.DB) {
	fmt.Println("\n🔗 Relationships and Associations")
	fmt.Println("----------------------------------")

	// Get the starting IDs
	nextTagID := getNextID(db, "tags")
	nextPostID := getNextID(db, "posts")

	// Create a test tag first
	testTag := Tag{
		ID:   nextTagID,
		Name: "test-single",
	}
	result := db.Create(&testTag)
	if result.Error != nil {
		log.Printf("Error creating test tag: %v", result.Error)
		return
	}
	fmt.Printf("✅ Created test tag: %s\n", testTag.Name)

	// Create tags with manual ID assignment
	tags := []Tag{
		{ID: nextTagID + 1, Name: "go"},
		{ID: nextTagID + 2, Name: "database"},
		{ID: nextTagID + 3, Name: "tutorial"},
	}

	// Create tags individually to handle unique constraints
	for i := range tags {
		result := db.Create(&tags[i])
		if result.Error != nil {
			log.Printf("Error creating tag %s: %v", tags[i].Name, result.Error)
			continue
		}
		fmt.Printf("✅ Created tag: %s\n", tags[i].Name)
	}

	// Get the first user for posts
	var firstUser User
	if err := db.First(&firstUser).Error; err != nil {
		log.Printf("No users found for creating posts: %v", err)
		return
	}

	// Create posts with relationships
	posts := []Post{
		{
			ID:      nextPostID,
			Title:   "Getting Started with GORM",
			Content: "This is a comprehensive guide to GORM basics...",
			UserID:  firstUser.ID,
		},
		{
			ID:      nextPostID + 1,
			Title:   "Advanced DuckDB Features",
			Content: "Exploring advanced features of DuckDB database...",
			UserID:  firstUser.ID,
		},
	}

	// Create posts individually
	for i := range posts {
		result := db.Create(&posts[i])
		if result.Error != nil {
			log.Printf("Error creating post %s: %v", posts[i].Title, result.Error)
			continue
		}
		fmt.Printf("✅ Created post: %s\n", posts[i].Title)

		// Associate with tags (only with successfully created tags)
		var availableTags []Tag
		db.Where("name IN ?", []string{"go", "database"}).Find(&availableTags)
		if len(availableTags) > 0 {
			err := db.Model(&posts[i]).Association("Tags").Append(availableTags)
			if err != nil {
				log.Printf("Error associating tags with post: %v", err)
			} else {
				fmt.Printf("🏷️ Associated %d tags with post: %s\n", len(availableTags), posts[i].Title)
			}
		}
	}

	// Demonstrate preloading relationships
	var userWithPosts User
	db.Preload("Posts.Tags").First(&userWithPosts)
	fmt.Printf("📄 User %s has %d posts\n", userWithPosts.Name, len(userWithPosts.Posts))
}

func demonstrateDuckDBFeatures(db *gorm.DB) {
	fmt.Println("\n🦆 DuckDB-Specific Features")
	fmt.Println("----------------------------")

	// Get the starting ID for products
	nextProductID := getNextID(db, "products")

	// Create sample products
	products := []Product{
		{
			ID:          nextProductID,
			Name:        "Laptop",
			Price:       999.99,
			Description: "High-performance laptop for developers",
		},
		{
			ID:          nextProductID + 1,
			Name:        "Coffee Maker",
			Price:       149.99,
			Description: "Premium coffee maker with programmable features",
		},
	}

	result := db.Create(&products)
	if result.Error != nil {
		log.Printf("Error creating products: %v", result.Error)
	}
	fmt.Printf("✅ Created %d products\n", result.RowsAffected)

	// Demonstrate analytical queries
	var expensiveProducts []Product
	db.Where("price > ?", 500.0).Find(&expensiveProducts)
	fmt.Printf("🔍 Found %d expensive products\n", len(expensiveProducts))

	// Calculate average price
	var avgPrice float64
	err := db.Model(&Product{}).Select("AVG(price)").Row().Scan(&avgPrice)
	if err != nil {
		log.Printf("Error calculating average price: %v", err)
		avgPrice = 0
	}
	fmt.Printf("💰 Average product price: $%.2f\n", avgPrice)
}

func demonstrateAdvancedQueries(db *gorm.DB) {
	fmt.Println("\n🔍 Advanced Queries")
	fmt.Println("-------------------")

	// Count users by age groups
	type UserStat struct {
		AgeGroup string
		Count    int64
	}

	var userStats []UserStat
	db.Model(&User{}).
		Select("CASE WHEN age < 30 THEN 'Young' ELSE 'Mature' END as age_group, COUNT(*) as count").
		Group("age_group").
		Scan(&userStats)

	fmt.Println("📊 User statistics:")
	for _, stat := range userStats {
		fmt.Printf("   %s: %d users\n", stat.AgeGroup, stat.Count)
	}

	// Demonstrate transaction
	fmt.Println("\n💳 Transaction Example")

	err := db.Transaction(func(tx *gorm.DB) error {
		// Get the next post ID
		nextPostID := getNextID(tx, "posts")

		// Get the first user
		var user User
		if err := tx.First(&user).Error; err != nil {
			return err
		}

		// Create a post within transaction
		post := Post{
			ID:      nextPostID,
			Title:   "Transaction Post",
			Content: "Created in transaction",
			UserID:  user.ID,
		}

		if err := tx.Create(&post).Error; err != nil {
			return err // This will rollback the transaction
		}

		fmt.Printf("✅ Created post in transaction: %s\n", post.Title)
		return nil
	})

	if err != nil {
		fmt.Println("❌ Transaction failed and rolled back")
	} else {
		fmt.Println("✅ Transaction completed successfully")
	}

	// Final count
	var userCount, postCount, tagCount, productCount int64
	db.Model(&User{}).Count(&userCount)
	db.Model(&Post{}).Count(&postCount)
	db.Model(&Tag{}).Count(&tagCount)
	db.Model(&Product{}).Count(&productCount)

	fmt.Printf("\n📈 Final Database State:\n")
	fmt.Printf("   👥 Users: %d\n", userCount)
	fmt.Printf("   📄 Posts: %d\n", postCount)
	fmt.Printf("   🏷️  Tags: %d\n", tagCount)
	fmt.Printf("   📦 Products: %d\n", productCount)
}
