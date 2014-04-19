package gameserver

import (
	"fmt"
	"os"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true
)

var LOGV = util.NewLogger(DEBUG_LOG, "DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "ERROR", os.Stderr)
