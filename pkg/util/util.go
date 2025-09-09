package util

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/types/errno"
)

func SetAuthorization(c *gin.Context, access string, refresh string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.Header("x-access-token", access)
	c.SetCookie("llmops_refresh", refresh, int(time.Hour*24), "/", "", false, true)
}

func GetCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, errno.ErrUnauthorized.AppendBizMessage(errors.New("未登录"))
	}

	return userID, nil
}

func GenerateHash(text string) string {
	text = text + "None"
	hash := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", hash[:])
}

func ConvertViaJSON(dest interface{}, src interface{}) error {
	jsonData, err := sonic.Marshal(src)
	if err != nil {
		return err
	}

	return sonic.Unmarshal(jsonData, dest)
}

func GetValueType(value any) string {
	valueType := reflect.TypeOf(value).Kind().String()

	switch valueType {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	default:
		return valueType
	}
}

func Contains(slice any, value any) bool {
	sliceVal := reflect.ValueOf(slice)
	if sliceVal.Kind() != reflect.Slice {
		return false // 不是切片，直接返回 false
	}

	targetVal := reflect.ValueOf(value)
	for i := 0; i < sliceVal.Len(); i++ {
		elem := sliceVal.Index(i)
		if reflect.DeepEqual(elem.Interface(), targetVal.Interface()) {
			return true
		}
	}
	return false
}

func LessThan(a, b interface{}) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return aVal.Int() < bVal.Int()
	case reflect.Float32, reflect.Float64:
		return aVal.Float() < bVal.Float()
	default:
		return false
	}
}

func GreaterThan(a, b interface{}) bool {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	switch aVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return aVal.Int() > bVal.Int()
	case reflect.Float32, reflect.Float64:
		return aVal.Float() > bVal.Float()
	default:
		return false
	}
}

// UniqueStrings 返回去重后的字符串切片
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func UniqueUUID(slice []uuid.UUID) []uuid.UUID {
	keys := make(map[uuid.UUID]bool)
	list := []uuid.UUID{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
