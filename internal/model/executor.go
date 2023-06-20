package model

// StartCommandRequest is used to specify command to run on executor.
type StartCommandRequest struct {
	ID      string            `path:"id"`
	Dir     string            `json:"dir"`
	Command []string          `json:"command" vd:"len($)>0; msg:'command cannot be empty'"`
	Env     map[string]string `json:"env"`
	// TODO: implement it
	// User    string            `json:"user"`
}
