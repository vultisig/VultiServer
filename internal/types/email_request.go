package types

type EmailRequest struct {
	Email       string `json:"email"`
	FileName    string `json:"file_name"`
	FileContent string `json:"file_content"`
}
