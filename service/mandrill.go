package service

type MandrillTo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type MandrilMergeVarContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}
type MandrillVar struct {
	Rcpt string                   `json:"rcpt"`
	Vars []MandrilMergeVarContent `json:"vars"`
}
type MandrillAttachment struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
}
type MandrillMessage struct {
	To            []MandrillTo         `json:"to"`
	SendingDomain string               `json:"sending_domain"`
	MergeVars     []MandrillVar        `json:"merge_vars"`
	Attachments   []MandrillAttachment `json:"attachments"`
}

type MandrillPayload struct {
	Key             string                   `json:"key"`
	TemplateName    string                   `json:"template_name"`
	TemplateContent []MandrilMergeVarContent `json:"template_content"`
	Message         MandrillMessage          `json:"message"`
}
