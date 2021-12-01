package api

type module struct {
	Name         string
	ModuleCode   string
	Id           string
	CreatorName  string
	CreatorEmail string
}

const MODULE_URL_ENDPOINT = "https://luminus.nus.edu.sg/v2/api/module/?populate=Creator%2CtermDetail%2CisMandatory"

func Get() {
	// request := Request{MODULE_URL_ENDPOINT, "s", "s"}
}
