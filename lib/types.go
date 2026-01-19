package lib

import (
	"time"
)

type RedmineIssue struct {
	Issue Issue `json:"issue"`
}

type RedmineNews struct {
	News News `json:"news"`
}

type RedmineDate struct {
	time.Time
}

type RedmineAPIObject struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	IsClosed *bool  `json:"is_closed"`
}

var IssuePriority = struct {
	Default int
	Low     int
	Normal  int
	High    int
	Urgent  int
}{
	Default: 5, // 5 is the lowest priority
	Low:     4, // 4 is low priority
	Normal:  3, // 3 is normal priority
	High:    2, // 2 is high priority
	Urgent:  1, // 1 is the highest priority
}

// Working = in progress
//
// InBreak = in break time
//
// Feedback = waiting for employee
//
// Feedback2 = waiting for customer
var IssueStatus = struct {
	Working   int // in progress
	InBreak   int // in break time
	Feedback  int // waiting for employee
	Feedback2 int // waiting for customer
	Resolved  int
	Closed    int
}{
	Working:   2, // in progress
	InBreak:   7, // in break time
	Feedback:  8, // waiting for employee
	Feedback2: 4, // waiting for customer
	Resolved:  3,
	Closed:    5,
}
