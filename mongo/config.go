package mongo

// Config for mongo client
type Config struct {
	URI string `json:"uri"`
	Db  string `json:"db"`
}
