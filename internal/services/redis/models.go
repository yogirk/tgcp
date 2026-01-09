package redis

type Instance struct {
	Name              string // Short ID
	DisplayName       string
	ProjectID         string
	Location          string
	Tier              string // BASIC, STANDARD_HA
	MemorySizeGb      int
	RedisVersion      string // REDIS_6_X
	Host              string
	Port              int
	State             string // READY, CREATING
	AuthorizedNetwork string
}
