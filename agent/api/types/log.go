package types

import "time"

type LogRequest struct {
	Namespace string `form:"namespace" json:"namespace,omitempty"`
	Name      string `form:"name" json:"name,omitempty"`
	// Instance is the optional pod name, that allows you to request logs from a specific instance
	Instance string `form:"instance" json:"instance,omitempty"`
	// Follow is allows the user to request a stream of logs until the timeout
	Follow bool `form:"follow" json:"follow,omitempty"`
	// Tail sets the maximum number of log messages to return, <=0 means unlimited
	Tail  int    `form:"tail" json:"tail,omitempty"`
	Since string `form:"since" json:"since,omitempty"`
	// End is the end time of the log stream
	End string `form:"end" json:"end,omitempty"`
}

// Message is a specific log message from a function container log stream
type Message struct {
	// Name is the function name
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	// instance is the name/id of the specific function instance
	Instance string `json:"instance"`
	// Timestamp is the timestamp of when the log message was recorded
	Timestamp time.Time `json:"timestamp"`
	// Text is the raw log message content
	Text string `json:"text"`
}
