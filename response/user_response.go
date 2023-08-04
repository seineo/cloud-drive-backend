package response

type UserSignResponse struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	RootHash string `json:"rootHash" binding:"required"`
}
