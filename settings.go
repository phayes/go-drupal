package drupal

type Settings map[string]interface{}

func (s Settings) HasValue(key string) bool {
	_, ok := s[key]
	return ok
}

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
