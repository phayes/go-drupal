package drupal

// Settings represents drupal settings defined in $settings of settings.php
type Settings map[string]interface{}

// HasValue checks if the settings has the specific key defined
func (s Settings) HasValue(key string) bool {
	_, ok := s[key]
	return ok
}

// GetString gets a settings value as a string
func (s Settings) GetString(key string) string {
	val, ok := s[key]
	if !ok {
		return ""
	}
	strval, ok := val.(string)
	if !ok {
		return ""
	}
	return strval
}

// GetInt gets a settings value as an int
func (s Settings) GetInt(key string) int {
	val, ok := s[key]
	if !ok {
		return 0
	}
	intval, ok := val.(int)
	if !ok {
		return 0
	}
	return intval
}

// GetBool gets a settings value as a bool
// To differentiate between a false value and an empty value, use HasValue()
func (s Settings) GetBool(key string) bool {
	val, ok := s[key]
	if !ok {
		return false
	}
	boolval, ok := val.(bool)
	if !ok {
		return false
	}
	return boolval
}

// GetFloat gets a settings value as a float
func (s Settings) GetFloat(key string) float64 {
	val, ok := s[key]
	if !ok {
		return 0
	}
	floatval, ok := val.(float64)
	if !ok {
		return 0
	}
	return floatval
}

// GetAssocArray gets an associate array settings value and returns it as a Settings struct
func (s Settings) GetAssocArray(key string) Settings {
	val, ok := s[key]
	if !ok {
		return nil
	}
	settingsval, ok := val.(Settings)
	if !ok {
		return nil
	}
	return settingsval
}

// GetArray gets an array of string settings values
func (s Settings) GetArray(key string) []string {
	val, ok := s[key]
	if !ok {
		return nil
	}

	array := []string{}
	arrayval := val.([]interface{})

	for _, subval := range arrayval {
		strval, ok := subval.(string)
		if !ok {
			continue
		}
		array = append(array, strval)
	}

	return array
}
