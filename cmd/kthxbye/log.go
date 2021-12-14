package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupLogger() {
	log.Logger = zerolog.New(os.Stderr).With().Logger().Output(zerolog.ConsoleWriter{
		Out:     os.Stderr,
		NoColor: true,
		FormatLevel: func(level interface{}) string {
			if level == nil {
				return ""
			}
			return fmt.Sprintf("level=%s", level)
		},
		FormatMessage: func(msg interface{}) string {
			return fmt.Sprintf("msg=%q", msg)
		},
		FormatTimestamp: func(interface{}) string {
			return ""
		},
	})
}

func silenceMatchersToLogField(s Silence) (matchers []string) {
	for _, sm := range s.Matchers {
		op := "="
		if sm.IsRegex {
			op = "=~"
		}
		matchers = append(matchers, fmt.Sprintf("%s%s%s", sm.Name, op, sm.Value))
	}
	return matchers
}
