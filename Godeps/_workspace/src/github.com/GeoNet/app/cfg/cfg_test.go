package cfg

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Load with no env var overrides set.
	unset()

	c := Load()

	if c.DataBase.Host != "a" {
		t.Errorf("got %s for DataBase.Host expected %s", c.DataBase.Host, "a")
	}
	if c.DataBase.MaxOpenConns != 1 {
		t.Errorf("got %d for DataBase.MaxOpenConns expected %d", c.DataBase.MaxOpenConns, 1)
	}
	if c.WebServer != nil {
		t.Errorf("Expected nil c.WebServer got %v", c.WebServer)
	}
	if c.Librato.User != "XXX" {
		t.Errorf("got %s for Librato.User expected %s", c.Librato.User, "XXX")
	}
	if c.Logentries.Token != "ZZZ" {
		t.Errorf("got %s for Logentries.Token expected %s", c.Logentries.Token, "ZZZ")
	}

	// change where Load will look for override files instead of /etc/sysconfig
	over = "etc"

	c = Load()

	if c.DataBase.Host != "aa" {
		t.Errorf("got %s for DataBase.Host expected %s", c.DataBase.Host, "aa")
	}
	if c.DataBase.MaxOpenConns != 10 {
		t.Errorf("got %d for DataBase.MaxOpenConns expected %d", c.DataBase.MaxOpenConns, 10)
	}
	if !c.WebServer.Production {
		t.Error("Expected c.WebServer.Production to be true")
	}

	// load again (from override in etc) but with an env var set
	set()

	c = Load()

	if c.DataBase.Host != "aaa" {
		t.Errorf("got %s for DataBase.Host expected %s", c.DataBase.Host, "aaa")
	}
	if c.DataBase.MaxOpenConns != 100 {
		t.Errorf("got %d for DataBase.MaxOpenConns expected %d", c.DataBase.MaxOpenConns, 100)
	}
	if c.WebServer.Production {
		t.Error("Expected c.WebServer.Production to be false")
	}

	// unset env var and load from local default config.
	unset()
	over = "/etc/sysconfig"

	c = Load()

	if c.DataBase.Host != "a" {
		t.Errorf("got %s for DataBase.Host expected %s", c.DataBase.Host, "a")
	}
	if c.DataBase.MaxOpenConns != 1 {
		t.Errorf("got %d for DataBase.MaxOpenConns expected %d", c.DataBase.MaxOpenConns, 1)
	}

	// set env and check override of local default config.
	set()

	c = Load()
	if c.DataBase.Host != "aaa" {
		t.Errorf("got %s for DataBase.Host expected %s", c.DataBase.Host, "aaa")
	}
	if c.DataBase.MaxOpenConns != 100 {
		t.Errorf("got %d for DataBase.MaxOpenConns expected %d", c.DataBase.MaxOpenConns, 100)
	}
	if c.WebServer != nil {
		t.Errorf("Expected nil c.WebServer got %v", c.WebServer)
	}
	if c.Librato.User != "XXXX" {
		t.Errorf("got %s for Librato.User expected %s", c.Librato.User, "XXXX")
	}
	if c.Logentries.Token != "ZZZZ" {
		t.Errorf("got %s for Logentries.Token expected %s", c.Logentries.Token, "ZZZZ")
	}

	d, err := c.EnvDoc()
	if err != nil {
		t.Errorf("Non nil error %s", err)
	}
	if len(d) == 0 {
		t.Error("Zero length for d.")
	}
}

func unset() {
	// Set the test vars empty.
	// os.Unsetenv is in Go 1.4 but this approach is fine for this use.
	os.Setenv("GO_CFG_TEST_DATABASE_HOST", "")
	os.Setenv("GO_CFG_TEST_DATABASE_MAX_OPEN_CONNS", "")
	os.Setenv("GO_CFG_TEST_WEB_SERVER_PRODUCTION", "")
	os.Setenv("LIBRATO_USER", "")
	os.Setenv("LOGENTRIES_TOKEN", "")
}

func set() {
	os.Setenv("GO_CFG_TEST_DATABASE_HOST", "aaa")
	os.Setenv("GO_CFG_TEST_DATABASE_MAX_OPEN_CONNS", "100")
	os.Setenv("GO_CFG_TEST_WEB_SERVER_PRODUCTION", "false")
	os.Setenv("LIBRATO_USER", "XXXX")
	os.Setenv("LOGENTRIES_TOKEN", "ZZZZ")
}
