package utils

func GetValueOrDefault(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
