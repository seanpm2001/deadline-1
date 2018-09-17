package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var good = Config{
	DAO: "file",
	FileConfig: FileConfig{
		Directory: "../server/",
	},
	DBConfig: DBConfig{
		ConnectionString: "N/A",
	},
	Server: ServerConfig{
		Port: "8081",
	},
	Logconfig: make(map[string]string),
}

func TestGoodConfig(test *testing.T) {
	cfg, err := LoadConfig("testdata/goodfile.yml")
	assert.Nil(test, err, "Could not load a good file")
	assert.Equal(test, &good, cfg)
	assert.NotNil(test, cfg.Logconfig, "")
}

func TestCantFindConfig(test *testing.T) {
	cfg, err := LoadConfig("testdata/cantfindfile.yml")
	assert.NotNil(test, err, "Loaded a bad file")
	assert.Nil(test, cfg, "")
}
