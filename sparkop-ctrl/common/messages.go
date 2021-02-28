package common

import(
)

type SparkService struct {
	ServiceName string  `json:"servicename"`
	ServiceType string  `json:"servicetype"`  // tcp, webm udp
	EntryPoint string   `json:"entrypoint"`   // web, mysql, pg, hive2, sebsecuret
	ServicePort int32   `json:"serviceport"`
	TargetPort  int32   `json:"targetport"`
	ServiceIP   string    `json:"serviceip"`
	RoutePath   string  `json:"routepath"`
}

type StateMessage struct {
	State string    `json:"state"`
	Message string   `json:"message"`
}

type SparkAppState struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	State     StateMessage   `json:"state"`
	//DriverState StateMessage   `json:"driverstate"`
	//ExecutorState  StateMessage   `json:"executorstate"`
}

type SparkAppCreatedMsg struct {
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	UID string `json:"uid"`
	ExtraServices []*SparkService `json: extraservices`
}

type DBConnConfig struct {
	Type string `form:"type" json:"type" xml:"type"`
	Host string `form:"host" json:"host" xml:"host"`
	Port string `form:"port" json:"port" xml:"port"`
	User string `form:"user" json:"user" xml:"user"`
	Password string `form:"password" json:"password" xml:"password"`
	Dbname string `form:"dbname" json:"dbname" xml:"dbname"`
}

type RedisConfig struct {
    Addr       string `form:"addr" json:"addr" xml:"addr" yaml:"addr"`
    Db         int    `form:"db" json:"db" xml:"db" yaml:"db"`
    Password   string `form:"password" json:"password" xml:"password" yaml:"password"`
	Poolsize int    `form:"poolsize" json:"poolsize" xml:"poolsize" yaml:"poolsize"`
    Cache struct {
        Tokenexpired int  `form:"tokenexpired" json:"tokenexpired" xml:"tokenexpired" yaml:"tokenexpired"`
        Userexpired  int  `form:"userexpired" json:"userexpired" xml:"userexpired" yaml:"userexpired"`
    }
}

