package main
import (
  "fmt"
  "ccms.com/api/internal/utils"
  "github.com/google/uuid"
)
func main(){
  u := uuid.MustParse("a07486fa-0967-418a-9a98-971975da70ea")
  token, err := utils.GenerateTokens(u, "your-super-secret-key-min-32-chars-change-in-production")
  fmt.Println("generate err:", err)
  fmt.Println(token)
  uid, err := utils.ValidateToken(token, "your-super-secret-key-min-32-chars-change-in-production")
  fmt.Printf("uid=%v err=%v\n", uid, err)
}
