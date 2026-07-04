// Package watchdog implements the computer opponent for watchdog-chess
package watchdog

import "github.com/OGBlackDiamond/watchdog-chess/engine"

type WatchdogResult struct {
	Move engine.Move
	Found bool
	Err error
}
