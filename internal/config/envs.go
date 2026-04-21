package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Environment struct {
	Name string   `json:"name"`
	Vars []EnvVar `json:"vars"`
}

type Environments struct {
	Active int           `json:"active"` // -1 = none selected
	Envs   []Environment `json:"envs"`
}

func (e *Environments) ActiveName() string {
	if e.Active < 0 || e.Active >= len(e.Envs) {
		return ""
	}
	return e.Envs[e.Active].Name
}

func (e *Environments) VarMap() map[string]string {
	if e.Active < 0 || e.Active >= len(e.Envs) {
		return nil
	}
	env := e.Envs[e.Active]
	m := make(map[string]string, len(env.Vars))
	for _, v := range env.Vars {
		if v.Key != "" {
			m[v.Key] = v.Value
		}
	}
	return m
}

func LoadEnvs() (Environments, error) {
	path, err := envsPath()
	if err != nil {
		return Environments{Active: -1}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Environments{Active: -1}, nil
	}
	if err != nil {
		return Environments{Active: -1}, fmt.Errorf("envs lesen: %w", err)
	}
	var e Environments
	if err := json.Unmarshal(data, &e); err != nil {
		return Environments{Active: -1}, fmt.Errorf("envs parsen: %w", err)
	}
	return e, nil
}

func SaveEnvs(e Environments) error {
	path, err := envsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("config-verzeichnis anlegen: %w", err)
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func envsPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "apish", "envs.json"), nil
}
