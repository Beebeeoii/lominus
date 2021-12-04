package api

type Request struct {
	Url       string
	JwtToken  string
	UserAgent string
}

const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0"
