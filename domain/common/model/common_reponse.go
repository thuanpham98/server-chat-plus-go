package domain_common_model

type CommonReponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data any `json:"data"`
}