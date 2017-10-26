package main

import (
	"fmt"
	"time"

	"github.com/metakeule/fmtdate"
)

func main() {

	stringOffset := "+06.00h"

	offSet, err := time.ParseDuration(stringOffset)
	if err != nil {
		panic(err)
	}
	date := fmtdate.Format("MM/DD/YYYY hh:mm:ss", time.Now().Add(offSet))
	fmt.Println(date)

}
