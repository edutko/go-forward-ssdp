package config

import (
	"io"
	"io/ioutil"
	"os"
)

type Config struct {
	ListenIfNames []string
	SendIfNames   []string
}

func LoadFromFile(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	return LoadFromReader(f)
}

func LoadFromReader(r io.Reader) (Config, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return Config{}, err
	}
	return LoadFromBytes(data)
}

func LoadFromBytes(data []byte) (Config, error) {
	return Config{}, nil
}
