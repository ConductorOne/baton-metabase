package main

import (
	cfg "github.com/conductorone/baton-metabase/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("metabase", cfg.Config)
}
