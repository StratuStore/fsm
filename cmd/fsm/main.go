package main

import (
	"github.com/StratuStore/fsm/internal/fsm/app"
	"github.com/StratuStore/fsm/internal/libs/config"
	"go.uber.org/fx"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	fx.New(app.CreateApp(cfg)).Run()
}
