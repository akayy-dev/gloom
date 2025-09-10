package utils

import "time"

/*
Channel watches for when the event loop is running,
ensures binding functions don't run before the event loop does.
*/
var ProgramRunning = make(chan struct{})

/*
Wrapper function that waits for the BubbleTea event loop to start before running
bindingFunc(), times out after 5 seconds.
*/
func waitForEventLoop(bindingFunc func(), funcName string) {
	go func() {
		select {
		case <-ProgramRunning:
			bindingFunc()
		case <-time.After(5 * time.Second):
			UserLog.Warnf("%s() timed out waiting for event loop to start", funcName)
		}
	}()
}

func Notify(prompt string, displayMs int) {
	waitForEventLoop(func() {
		Program.Send(SendNotificationMsg{
			Message:     prompt,
			DisplayTime: displayMs,
		})
	}, "Notify")
}
