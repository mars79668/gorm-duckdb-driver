module example

go 1.24

toolchain go1.24.4

require (
	gorm.io/driver/duckdb v1.0.0
	gorm.io/gorm v1.25.12
)

replace gorm.io/driver/duckdb => ../

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/ramya-rao-a/go-outline v0.0.0-20210608161538-9736a4bde949 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
)
