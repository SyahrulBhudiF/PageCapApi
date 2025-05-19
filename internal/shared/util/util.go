package util

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/bcrypt"
	"math/big"
	"sync"
)

var PasswordPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&#])[A-Za-z\d@$!%*?&#]{8,}$`

var (
	json             = jsoniter.ConfigCompatibleWithStandardLibrary
	UploadParamsPool = sync.Pool{
		New: func() interface{} {
			return &uploader.UploadParams{
				ResourceType: "image",
			}
		},
	}
)

func HashPassword(password string, salt string) string {
	salted := password + salt
	hashedByte, err := bcrypt.GenerateFromPassword([]byte(salted), bcrypt.DefaultCost)
	hashed := string(hashedByte)

	if err != nil {
		panic(err)
	}

	return hashed
}

func ComparePassword(hashedPassword, password, salt string) bool {
	salted := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(salted))
	return err == nil
}

func GenerateOTP() string {
	mx := big.NewInt(1000000)

	n, _ := rand.Int(rand.Reader, mx)

	return fmt.Sprintf("%06d", n)
}

func GetBody[T any](c *gin.Context, key string) (T, error) {
	val, exists := c.Get(key)
	if !exists {
		var zero T
		return zero, fmt.Errorf("request %s not found", key)
	}

	body, ok := val.(*T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("invalid request %s type", key)
	}

	return *body, nil
}

func ErrorInList(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

const OptimalBufferSize = 64 * 1024

var BufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, OptimalBufferSize))
	},
}

func GetBuffer() *bytes.Buffer {
	return BufferPool.Get().(*bytes.Buffer)
}

func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	BufferPool.Put(buf)
}
