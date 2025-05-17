package route

import (
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
)

func CustomRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			fmt.Printf("Panic recovered: %v\n", err)
			response.InternalServerError(c, err)
		}
	})
}
