package main

import (
	"encoding/base32"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {

	buf := make([]byte, 10)

	rand.Seed(time.Now().Unix())

	_, err := rand.Read(buf)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read data")
	}

	fmt.Print(strings.ToLower(base32.HexEncoding.EncodeToString(buf)))
}
