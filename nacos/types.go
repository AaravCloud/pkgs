package nacos

type ConfigNaCos struct {
	NaCosServer      string `yaml:"NaCosServer"`
	NaCosPort        int    `yaml:"NaCosPort"`
	NaCosNameSpaceId string `yaml:"NaCosNameSpaceId"`
	NaCosDataId      string `yaml:"NaCosDataId"`
	NaCosGroup       string `yaml:"NaCosGroup"`
}

type ConfigPostGreSQL struct {
	Datasource Datasource `yaml:"datasource"`
	Hikari     Hikari     `yaml:"hikari"`
}

type Datasource struct {
	URL             string `yaml:"url"`               //数据库连接地址（格式：jdbc:postgresql://主机:端口/数据库名）
	Username        string `yaml:"username"`          //用户名
	Password        string `yaml:"password"`          //密码
	DriverClassName string `yaml:"driver-class-name"` //驱动
}

type Hikari struct {
	ConnectionTimeout int `yaml:"connection-timeout"` //链接超时时间（单位:毫秒）
	MaximumPoolSize   int `yaml:"maximum-pool-size"`  //最大链接池大小
	MinimumIdle       int `yaml:"minimum-idle"`       //最小空闲链接数
	IdleTimeout       int `yaml:"idle-timeout"`       //空闲链接超时时间
	MaxLifetime       int `yaml:"max-lifetime"`       //链接最大存活时间
}

type APPConfig struct {
	PostGreSQL ConfigPostGreSQL `yaml:"spring"`
}
