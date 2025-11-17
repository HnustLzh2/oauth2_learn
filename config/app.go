package config

type App struct {
	Session struct {
		Name      string `yaml:"name"`
		SecretKey string `yaml:"secret_key"`
		MaxAge    int    `yaml:"max_age"`
	} `yaml:"session"`

	AuthMode string `yaml:"auth_mode"`

	DB struct {
		Default DB `yaml:"default"`
	} `yaml:"db"`

	LDAP LDAP `yaml:"ldap"`

	Redis struct {
		Default Redis `yaml:"redis"`
	} `yaml:"redis"`

	OAuth2 struct {
		AccessTokenExp string         `yaml:"access_token_exp"`
		JWTSignedKey   string         `yaml:"jwt_signed_key"`
		Client         []OAuth2Client `yaml:"client"`
	} `yaml:"oauth2"`
}

type DB struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type OAuth2Client struct {
	ID     string  `yaml:"client_id"`
	Secret string  `yaml:"client_secret"`
	Name   string  `yaml:"client_name"`
	Domain string  `yaml:"client_domain"`
	Scope  []Scope `yaml:"client_scope"`
}

type Scope struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

type LDAP struct {
	URL            string `yaml:"url"`
	SearchDN       string `yaml:"search_dn"`
	SearchPassword string `yaml:"search_password"`
	BaseDN         string `yaml:"base_dn"`
	Filter         string `yaml:"filter"`
}
