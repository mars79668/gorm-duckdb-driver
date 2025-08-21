package duckdb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ===== STRUCT TYPES =====

// StructType represents a DuckDB STRUCT type - complex nested data with named fields
type StructType map[string]interface{}

// Value implements driver.Valuer interface for StructType
func (s StructType) Value() (driver.Value, error) {
	if s == nil {
		return "NULL", nil
	}

	if len(s) == 0 {
		return "{}", nil
	}

	var parts []string
	for key, value := range s {
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int, int64, float64, float32:
			valueStr = fmt.Sprintf("%v", v)
		case bool:
			valueStr = strconv.FormatBool(v)
		case nil:
			valueStr = "NULL"
		default:
			// Fallback to JSON encoding for complex types
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal struct field %s: %w", key, err)
			}
			valueStr = fmt.Sprintf("'%s'", strings.ReplaceAll(string(jsonBytes), "'", "''"))
		}
		parts = append(parts, fmt.Sprintf("'%s': %s", key, valueStr))
	}

	return "{" + strings.Join(parts, ", ") + "}", nil
}

// Scan implements sql.Scanner interface for StructType
func (s *StructType) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return s.scanFromString(v)
	case []byte:
		return s.scanFromString(string(v))
	case map[string]interface{}:
		*s = StructType(v)
		return nil
	default:
		// Try JSON unmarshaling as fallback
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot scan %T into StructType", value)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal JSON into StructType: %w", err)
		}
		*s = StructType(result)
		return nil
	}
}

func (s *StructType) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "NULL" || str == "" {
		*s = nil
		return nil
	}

	// Simple struct parsing - could be enhanced for complex nested cases
	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	if strings.TrimSpace(str) == "" {
		*s = make(StructType)
		return nil
	}

	// Try JSON unmarshaling first
	var result map[string]interface{}
	if err := json.Unmarshal([]byte("{"+str+"}"), &result); err == nil {
		*s = StructType(result)
		return nil
	}

	// Fallback to simple parsing
	result = make(map[string]interface{})
	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.Trim(strings.TrimSpace(parts[0]), "'\"")
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				value = value[1 : len(value)-1]
			}
			result[key] = value
		}
	}

	*s = StructType(result)
	return nil
}

// GormDataType implements the GormDataTypeInterface for StructType
func (StructType) GormDataType() string {
	return "STRUCT"
}

// ===== MAP TYPES =====

// MapType represents a DuckDB MAP type - key-value pairs with typed keys and values
type MapType map[string]interface{}

// Value implements driver.Valuer interface for MapType
func (m MapType) Value() (driver.Value, error) {
	if m == nil {
		return "MAP {}", nil
	}

	if len(m) == 0 {
		return "MAP {}", nil
	}

	var pairs []string
	for key, value := range m {
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case int, int64, float64, float32:
			valueStr = fmt.Sprintf("%v", v)
		case bool:
			valueStr = strconv.FormatBool(v)
		case nil:
			valueStr = "NULL"
		default:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal map value for key %s: %w", key, err)
			}
			valueStr = fmt.Sprintf("'%s'", strings.ReplaceAll(string(jsonBytes), "'", "''"))
		}
		pairs = append(pairs, fmt.Sprintf("'%s': %s", key, valueStr))
	}

	return "MAP {" + strings.Join(pairs, ", ") + "}", nil
}

// Scan implements sql.Scanner interface for MapType
func (m *MapType) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return m.scanFromString(v)
	case []byte:
		return m.scanFromString(string(v))
	case map[string]interface{}:
		*m = MapType(v)
		return nil
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot scan %T into MapType", value)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal JSON into MapType: %w", err)
		}
		*m = MapType(result)
		return nil
	}
}

func (m *MapType) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "NULL" || str == "" || str == "MAP {}" {
		*m = make(MapType)
		return nil
	}

	// Remove MAP prefix if present
	if strings.HasPrefix(str, "MAP") {
		str = strings.TrimSpace(str[3:])
	}

	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	if strings.TrimSpace(str) == "" {
		*m = make(MapType)
		return nil
	}

	// Try JSON parsing
	var result map[string]interface{}
	if err := json.Unmarshal([]byte("{"+str+"}"), &result); err == nil {
		*m = MapType(result)
		return nil
	}

	// Fallback to simple parsing
	result = make(map[string]interface{})
	pairs := strings.Split(str, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.Trim(strings.TrimSpace(parts[0]), "'\"")
			value := strings.TrimSpace(parts[1])
			if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
				value = value[1 : len(value)-1]
			}
			result[key] = value
		}
	}

	*m = MapType(result)
	return nil
}

// GormDataType implements the GormDataTypeInterface for MapType
func (MapType) GormDataType() string {
	return "MAP(VARCHAR, VARCHAR)"
}

// ===== LIST TYPES (Dynamic Arrays) =====

// ListType represents a DuckDB LIST type - dynamic arrays with variable element types
type ListType []interface{}

// Value implements driver.Valuer interface for ListType
func (l ListType) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}

	if len(l) == 0 {
		return "[]", nil
	}

	elements := make([]string, 0, len(l))
	for _, item := range l {
		switch v := item.(type) {
		case string:
			elements = append(elements, fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''")))
		case int, int64, float64, float32:
			elements = append(elements, fmt.Sprintf("%v", v))
		case bool:
			elements = append(elements, strconv.FormatBool(v))
		case nil:
			elements = append(elements, "NULL")
		default:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal list element: %w", err)
			}
			elements = append(elements, fmt.Sprintf("'%s'", strings.ReplaceAll(string(jsonBytes), "'", "''")))
		}
	}

	return "[" + strings.Join(elements, ", ") + "]", nil
}

// Scan implements sql.Scanner interface for ListType
func (l *ListType) Scan(value interface{}) error {
	if value == nil {
		*l = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		return l.scanFromString(v)
	case []byte:
		return l.scanFromString(string(v))
	case []interface{}:
		*l = ListType(v)
		return nil
	default:
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot scan %T into ListType", value)
		}
		var result []interface{}
		if err := json.Unmarshal(jsonBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal JSON into ListType: %w", err)
		}
		*l = ListType(result)
		return nil
	}
}

func (l *ListType) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "[]" || str == "" {
		*l = ListType{}
		return nil
	}

	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		str = str[1 : len(str)-1]
	}

	if strings.TrimSpace(str) == "" {
		*l = ListType{}
		return nil
	}

	// Try JSON parsing first
	var result []interface{}
	if err := json.Unmarshal([]byte("["+str+"]"), &result); err == nil {
		*l = ListType(result)
		return nil
	}

	// Fallback to simple parsing
	parts := strings.Split(str, ",")
	result = make([]interface{}, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "'") && strings.HasSuffix(part, "'") {
			part = part[1 : len(part)-1]
			part = strings.ReplaceAll(part, "''", "'")
		}
		result = append(result, part)
	}

	*l = ListType(result)
	return nil
}

// GormDataType implements the GormDataTypeInterface for ListType
func (ListType) GormDataType() string {
	return "LIST"
}

// ===== DECIMAL TYPES =====

// DecimalType represents a DuckDB DECIMAL type with precise numeric operations
type DecimalType struct {
	Data      string // Store as string to preserve precision
	Precision int    // Total digits
	Scale     int    // Digits after decimal point
}

// NewDecimal creates a new DecimalType from a string representation
func NewDecimal(value string, precision, scale int) DecimalType {
	return DecimalType{
		Data:      value,
		Precision: precision,
		Scale:     scale,
	}
}

// Value implements driver.Valuer interface for DecimalType
func (d DecimalType) Value() (driver.Value, error) {
	if d.Data == "" {
		return "0", nil
	}
	return d.Data, nil
}

// Scan implements sql.Scanner interface for DecimalType
func (d *DecimalType) Scan(value interface{}) error {
	if value == nil {
		*d = DecimalType{}
		return nil
	}

	switch v := value.(type) {
	case string:
		d.Data = v
		return nil
	case []byte:
		d.Data = string(v)
		return nil
	case int64:
		d.Data = fmt.Sprintf("%d", v)
		return nil
	case float64:
		d.Data = fmt.Sprintf("%.10f", v)
		return nil
	default:
		d.Data = fmt.Sprintf("%v", value)
		return nil
	}
}

// Float64 returns the decimal value as a float64 (may lose precision)
func (d DecimalType) Float64() (float64, error) {
	return strconv.ParseFloat(d.Data, 64)
}

// String returns the string representation of the decimal
func (d DecimalType) String() string {
	return d.Data
}

// GormDataType implements the GormDataTypeInterface for DecimalType
func (d DecimalType) GormDataType() string {
	if d.Precision > 0 && d.Scale > 0 {
		return fmt.Sprintf("DECIMAL(%d,%d)", d.Precision, d.Scale)
	}
	return "DECIMAL"
}

// ===== INTERVAL TYPES =====

// IntervalType represents a DuckDB INTERVAL type for time calculations
type IntervalType struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
	Micros  int
}

// NewInterval creates a new IntervalType
func NewInterval(years, months, days, hours, minutes, seconds, micros int) IntervalType {
	return IntervalType{
		Years:   years,
		Months:  months,
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Micros:  micros,
	}
}

// Value implements driver.Valuer interface for IntervalType
func (i IntervalType) Value() (driver.Value, error) {
	var parts []string

	if i.Years != 0 {
		parts = append(parts, fmt.Sprintf("%d YEAR", i.Years))
	}
	if i.Months != 0 {
		parts = append(parts, fmt.Sprintf("%d MONTH", i.Months))
	}
	if i.Days != 0 {
		parts = append(parts, fmt.Sprintf("%d DAY", i.Days))
	}
	if i.Hours != 0 {
		parts = append(parts, fmt.Sprintf("%d HOUR", i.Hours))
	}
	if i.Minutes != 0 {
		parts = append(parts, fmt.Sprintf("%d MINUTE", i.Minutes))
	}
	if i.Seconds != 0 {
		parts = append(parts, fmt.Sprintf("%d SECOND", i.Seconds))
	}
	if i.Micros != 0 {
		parts = append(parts, fmt.Sprintf("%d MICROSECOND", i.Micros))
	}

	if len(parts) == 0 {
		return "INTERVAL '0 SECOND'", nil
	}

	return "INTERVAL '" + strings.Join(parts, " ") + "'", nil
}

// Scan implements sql.Scanner interface for IntervalType
func (i *IntervalType) Scan(value interface{}) error {
	if value == nil {
		*i = IntervalType{}
		return nil
	}

	switch v := value.(type) {
	case string:
		return i.parseInterval(v)
	case []byte:
		return i.parseInterval(string(v))
	case time.Duration:
		return i.fromDuration(v)
	default:
		return fmt.Errorf("cannot scan %T into IntervalType", value)
	}
}

func (i *IntervalType) parseInterval(str string) error {
	str = strings.TrimSpace(str)

	// Remove INTERVAL prefix if present
	if strings.HasPrefix(str, "INTERVAL") {
		str = strings.TrimSpace(str[8:])
	}

	// Remove quotes
	str = strings.Trim(str, "'\"")

	// Reset all fields
	*i = IntervalType{}

	// Simple parsing - could be enhanced for complex formats
	parts := strings.Fields(str)
	for j := 0; j < len(parts)-1; j += 2 {
		if j+1 >= len(parts) {
			break
		}

		value, err := strconv.Atoi(parts[j])
		if err != nil {
			continue
		}

		unit := strings.ToUpper(parts[j+1])
		switch unit {
		case "YEAR", "YEARS":
			i.Years = value
		case "MONTH", "MONTHS":
			i.Months = value
		case "DAY", "DAYS":
			i.Days = value
		case "HOUR", "HOURS":
			i.Hours = value
		case "MINUTE", "MINUTES":
			i.Minutes = value
		case "SECOND", "SECONDS":
			i.Seconds = value
		case "MICROSECOND", "MICROSECONDS":
			i.Micros = value
		}
	}

	return nil
}

func (i *IntervalType) fromDuration(d time.Duration) error {
	// Convert duration to interval components
	total := int64(d)

	i.Micros = int(total % 1000000)
	total /= 1000000

	i.Seconds = int(total % 60)
	total /= 60

	i.Minutes = int(total % 60)
	total /= 60

	i.Hours = int(total % 24)
	total /= 24

	i.Days = int(total)

	return nil
}

// ToDuration converts the interval to a Go time.Duration (approximate for days/months/years)
func (i IntervalType) ToDuration() time.Duration {
	duration := time.Duration(i.Micros) * time.Microsecond
	duration += time.Duration(i.Seconds) * time.Second
	duration += time.Duration(i.Minutes) * time.Minute
	duration += time.Duration(i.Hours) * time.Hour
	duration += time.Duration(i.Days) * 24 * time.Hour
	// Note: Months and years are approximate
	duration += time.Duration(i.Months) * 30 * 24 * time.Hour
	duration += time.Duration(i.Years) * 365 * 24 * time.Hour

	return duration
}

// GormDataType implements the GormDataTypeInterface for IntervalType
func (IntervalType) GormDataType() string {
	return "INTERVAL"
}

// ===== UUID TYPE =====

// UUIDType represents a DuckDB UUID type
type UUIDType struct {
	Data string // Store UUID as string
}

// NewUUID creates a new UUIDType from a string
func NewUUID(uuid string) UUIDType {
	return UUIDType{Data: uuid}
}

// Value implements driver.Valuer interface for UUIDType
func (u UUIDType) Value() (driver.Value, error) {
	if u.Data == "" {
		return nil, nil
	}
	return u.Data, nil
}

// Scan implements sql.Scanner interface for UUIDType
func (u *UUIDType) Scan(value interface{}) error {
	if value == nil {
		u.Data = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		u.Data = v
		return nil
	case []byte:
		u.Data = string(v)
		return nil
	default:
		u.Data = fmt.Sprintf("%v", value)
		return nil
	}
}

// String returns the UUID as a string
func (u UUIDType) String() string {
	return u.Data
}

// GormDataType implements the GormDataTypeInterface for UUIDType
func (UUIDType) GormDataType() string {
	return "UUID"
}

// ===== JSON TYPE =====

// JSONType represents a DuckDB JSON type with native JSON operations
type JSONType struct {
	Data interface{} // Can hold any JSON-serializable data
}

// NewJSON creates a new JSONType from any JSON-serializable data
func NewJSON(data interface{}) JSONType {
	return JSONType{Data: data}
}

// Value implements driver.Valuer interface for JSONType
func (j JSONType) Value() (driver.Value, error) {
	if j.Data == nil {
		return "NULL", nil
	}

	jsonBytes, err := json.Marshal(j.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON data: %w", err)
	}

	return string(jsonBytes), nil
}

// Scan implements sql.Scanner interface for JSONType
func (j *JSONType) Scan(value interface{}) error {
	if value == nil {
		j.Data = nil
		return nil
	}

	var jsonStr string
	switch v := value.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONType", value)
	}

	if jsonStr == "NULL" || jsonStr == "" {
		j.Data = nil
		return nil
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	j.Data = result
	return nil
}

// String returns the JSON as a formatted string
func (j JSONType) String() string {
	if j.Data == nil {
		return "null"
	}

	jsonBytes, err := json.Marshal(j.Data)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonBytes)
}

// GormDataType implements the GormDataTypeInterface for JSONType
func (JSONType) GormDataType() string {
	return "JSON"
}
