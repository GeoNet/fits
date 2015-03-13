// Package cfg is for loading application config with override capability for different environments.
package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// the directory to look for config override files in.
var over = "/etc/sysconfig"

type Config struct {
	DataBase   *DataBase
	WebServer  *WebServer
	SQS        *SQS
	SNS        *SNS
	SC3        *SC3
	Env        *Env
	Librato    *Librato
	Logentries *Logentries
	HeartBeat  *HeartBeat
}

// DataBase for database config.  Elements with an env tag can be overidden via env var.  See Load.
type DataBase struct {
	Host              string `doc:"database host name." env:"${PREFIX}_DATABASE_HOST"`
	Name              string `doc:"database name on Host."`
	User              string `doc:"database user."`
	Password          string `doc:"database User password (unencrypted)." env:"${PREFIX}_DATABASE_PASSWORD"`
	SSLMode           string `doc:"usually disable or require." env:"${PREFIX}_DATABASE_SSL_MODE"`
	MaxOpenConns      int    `doc:"database connection pool." env:"${PREFIX}_DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns      int    `doc:"database connection pool." env:"${PREFIX}_DATABASE_MAX_IDLE_CONNS"`
	ConnectionTimeOut int    `doc:"database connect timeout."`
}

type Env struct {
	Prefix string `doc:"prefix for env vars.  Provides env var name space."`
}

// WebServer for web server config.  Elements with an env tag can be overidden via env var.  See Load.
type WebServer struct {
	Port       string `doc:"web server port." env:"${PREFIX}_WEB_SERVER_PORT"`
	CNAME      string `doc:"public CNAME for the service." env:"${PREFIX}_WEB_SERVER_CNAME"`
	Production bool   `doc:"true if the app is production." env:"${PREFIX}_WEB_SERVER_PRODUCTION"`
}

// SQS for AWS SQS config.  Elements with an env tag can be overidden via env var.  See Load.
type SQS struct {
	AWSRegion         string `doc:"SQS region e.g., ap-southeast-2." env:"${PREFIX}_SQS_AWS_REGION"`
	QueueName         string `doc:"SQS queue name." env:"${PREFIX}_SQS_QUEUE_NAME"`
	AccessKey         string `doc:"SQS queue user access key." env:"${PREFIX}_SQS_ACCESS_KEY"`
	SecretKey         string `doc:"SQS queue user secret." env:"${PREFIX}_SQS_SECRET_KEY"`
	NumberOfListeners int    `doc:"number of SQS listeners." env:"${PREFIX}_SQS_NUMBER_OF_LISTENERS"`
}

type SNS struct {
	AWSRegion string `doc:"SNS region e.g., ap-southeast-2." env:"${PREFIX}_SNS_AWS_REGION"`
	AccessKey string `doc:"SNS queue user access key." env:"${PREFIX}_SNS_ACCESS_KEY"`
	SecretKey string `doc:"SNS queue user secret." env:"${PREFIX}_SNS_SECRET_KEY"`
	TopicArn  string `doc:"SNS Topic Arn." env:"${PREFIX}_SNS_TOPIC_ARN"`
}

type HeartBeat struct {
	ServiceID string `doc:"A service id for heartbeat messages." env:"${PREFIX}_SERVICE_ID"`
}

type SC3 struct {
	SpoolDir string `doc:"Spool directory for SeisComPML files." env:"${PREFIX}_SC3_SPOOL_DIR"`
	Site     string `doc:"The SC3 site - primary or backup." env:${PREFIX}_SC3_SITE"`
}

type Librato struct {
	User   string `doc:"username for Librato." env:"${PREFIX}_LIBRATO_USER"`
	Key    string `doc:"key for Librato." env:"${PREFIX}_LIBRATO_KEY"`
	Source string `doc:"source for metrics.  Appended to host if not empty." env:"${PREFIX}_LIBRATO_SOURCE"`
}

type Logentries struct {
	Token string `doc:"token for Logentries." env:"${PREFIX}_LOGENTRIES_TOKEN"`
}

func (c *Config) env() {
	if c.Env != nil {
		env(c.Env.Prefix, c.DataBase)
		env(c.Env.Prefix, c.WebServer)
		env(c.Env.Prefix, c.SQS)
		env(c.Env.Prefix, c.Librato)
		env(c.Env.Prefix, c.Logentries)
		env(c.Env.Prefix, c.SNS)
		env(c.Env.Prefix, c.HeartBeat)
		env(c.Env.Prefix, c.SC3)
	}
}

// Load loads config from JSON with options to override from the file system and environment variables.
// Load looks for a JSON file that matched the executable name suffixed with JSON.  If a JSON file is found
// and successfully loaded then environment variables with config.Env.Prefix are used to override the config.
// Failure to find and parse a JSON config file is a fatal error.
//
// For application foo with config.Env.Prefix FOO the load order is:
//   * Try to load /etc/sysconfig/foo.json
//   * If /etc/sysconfig/foo.json is not found then try to load ./foo.json
//   * Check the environment vars for any appropriate vars prefixed with FOO_ and override the config vals.
//     For example an env var FOO_DATABASE_PASSWORD will be read and override config.DataBase.Password
//   * Return config.
//
// Load can be used to init a var so that config will be available before init() funcs are called:
//   var config = cfg.Load()
//
// If override of config from
// environment vars would change application behaviour then it may be appropriate to override config elements after Load e.g.,
//   var config = cfg.Load()
//
//   func init() {
//	config.SQS.NumberOfListeners = 1
//     }
//
// The config file should be JSON that will parse into Config.  A complete example is shown below.  Objects can be omitted
// if not required e.g., if there is no SQS object in the JSON then Config will have a nil SQS pointer.  Elements of objects that
// are omitted in the JSON will have the corresponding type zero value in Config.
//
//   {
// 	"DataBase": {
// 		"Host": "localhost",
// 		"Name": "test",
// 		"User": "test_w",
// 		"Password": "test",
// 		"MaxOpenConns": 1,
// 		"MaxIdleConns": 2,
// 		"ConnectionTimeOut": 30,
// 		"SSLMode": "disable"
// 	},
// 	"SQS": {
// 		"AWSRegion": "ap-southeast-2",
// 		"QueueName": "XXX",
// 		"AccessKey": "XXX",
// 		"SecretKey": "XXX",
// 		"NumberOfListeners": 1
// 	},
// 	"WebServer": {
// 		"Port": "8080",
// 		"CNAME": "thing,com",
// 		"Production": false
// 	},
//         "Env": {
//                    "Prefix": "PREFIX"
//         },
//         "Librato": {
//	           "User": "XXX",
//                    "Key": "XXX"
//         },
//         "Logentries": {
//	           "Token": "XXX"
//         }
//   }
func Load() Config {
	a := strings.Split(os.Args[0], "/")
	name := a[len(a)-1]
	log.SetPrefix(name + " ")

	c := name + ".json"
	ov := over + "/" + c

	f, err := ioutil.ReadFile(ov)
	if err != nil {
		log.Printf("Could not load %s falling back to local file.", ov)
		f, err = ioutil.ReadFile("./" + c)
		if err != nil {
			log.Println("ERROR can't find any config for " + name)
			log.Fatal(err)
		}
	}

	var d Config
	err = json.Unmarshal(f, &d)
	if err != nil {
		log.Println("ERROR problem parsing config file.")
		log.Fatal(err)
	}

	d.env()

	return d
}

// Postgres returns a connection string that is suitable for use with sql for connecting to a Postgres DB.
func (d *DataBase) Postgres() string {
	return "host=" + d.Host +
		" connect_timeout=" + strconv.Itoa(d.ConnectionTimeOut) +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

// env takes v, a pointer to a struct and prefix, a string that is used for env var prefixes.
// It checks elements of the struct in v for tags "env:${PREFIX}_VAR_NAME".  If there is
// tag env then env var ${PREFIX}_VAR_NAME is loaded and parsed then element in v
// set to the env var value.
func env(prefix string, v interface{}) {
	if reflect.ValueOf(v).IsNil() {
		return
	}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return
	}

	if reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return
	}

	e := reflect.TypeOf(v).Elem().Name()

	st := reflect.TypeOf(v).Elem()
	sv := reflect.ValueOf(v).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if key := field.Tag.Get("env"); key != "" {
			key = strings.Replace(key, "${PREFIX}", prefix, 1)
			val := os.Getenv(key)
			if val != "" {
				switch field.Type.Kind() {
				case reflect.String:
					// don't log anything that might be sensitive data from the ENV
					log.Printf("overriding config %s.%s from env %s", e, field.Name, key)
					sv.Field(i).SetString(val)
				case reflect.Int:
					if n, err := strconv.ParseInt(val, 10, 64); err == nil {
						log.Printf("overriding config %s.%s from env %s", e, field.Name, key)
						sv.Field(i).SetInt(n)
					} else {
						log.Printf("ERROR parsing %s=%s from env as int: %s", key, val, err)
					}
				case reflect.Bool:
					if b, err := strconv.ParseBool(val); err == nil {
						log.Printf("overriding config %s.%s from env %s", e, field.Name, key)
						sv.Field(i).SetBool(b)
					} else {
						log.Printf("ERROR parsing %s=%s from env as bool: %s", key, val, err)

					}
				}
			}
		}
	}
}

type EnvDoc struct {
	Key, Val, Doc string
}

// EnvDoc returns information about Config values that can be overridden from env vars.
func (c *Config) EnvDoc() (d []EnvDoc, err error) {
	if c.Env != nil {
		d = append(d, envDoc(c.Env.Prefix, c.DataBase)...)
		d = append(d, envDoc(c.Env.Prefix, c.WebServer)...)
		d = append(d, envDoc(c.Env.Prefix, c.SQS)...)
		d = append(d, envDoc(c.Env.Prefix, c.Librato)...)
		d = append(d, envDoc(c.Env.Prefix, c.Logentries)...)
		d = append(d, envDoc(c.Env.Prefix, c.SNS)...)
		d = append(d, envDoc(c.Env.Prefix, c.HeartBeat)...)
		d = append(d, envDoc(c.Env.Prefix, c.SC3)...)
	} else {
		err = fmt.Errorf("Found nil Prefix in the config.  Don't know how to read env var.")
	}

	return
}

func envDoc(prefix string, v interface{}) (d []EnvDoc) {
	if reflect.ValueOf(v).IsNil() {
		return
	}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return
	}

	if reflect.ValueOf(v).Elem().Kind() != reflect.Struct {
		return
	}

	st := reflect.TypeOf(v).Elem()
	sv := reflect.ValueOf(v).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if key := field.Tag.Get("env"); key != "" {
			key = strings.Replace(key, "${PREFIX}", prefix, 1)
			doc := EnvDoc{
				Key: key,
				Val: fmt.Sprintf("%v", sv.Field(i).Interface()),
				Doc: field.Tag.Get("doc"),
			}
			d = append(d, doc)
		}
	}

	return
}
