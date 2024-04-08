package domain_common_modell

type CommonReponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data any `json:"data"`
}