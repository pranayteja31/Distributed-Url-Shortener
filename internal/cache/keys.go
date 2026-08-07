package cache

import "fmt"

func URLKey(shortCode string) string {
	return fmt.Sprintf("url:%s",shortCode)
}