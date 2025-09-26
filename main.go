package main

import (
	"fmt"
	"time"

	"github.com/Snakdy/terrarium/cmd"
)

var (
	version = "0.0.0"
	commit  = "develop"
	date    = time.Time{}.String()
)

func main() {
	cmd.Execute(fmt.Sprintf("%s-%s (%s)", version, commit, date))
}
