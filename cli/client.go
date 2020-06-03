package cli

type Config struct {
	Name    string `yaml:"name"`
	SvrHost string `yaml:"svr_host"`
	SvrPort int    `yaml:"svr_port"`
	Auth    string `yaml:"auth"`
	MaxConn int    `yaml:"max_conn"`
}

func Start(cf Config) {

}
