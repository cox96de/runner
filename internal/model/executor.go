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

// GetCommandLogRequest is used to get command log from executor.
type GetCommandLogRequest struct {
	ID string `path:"id"`
}

// GetCommandStatusRequest is used to get command status from executor.
type GetCommandStatusRequest = GetCommandLogRequest

// GetCommandStatusResponse presents command status.
type GetCommandStatusResponse struct {
	// ExitCode is the exit code of the command.
	// When Exit is false, this value is meaningless.
	ExitCode int `json:"exit_code"`
	// Exit is true when the command is finished.
	// Get exit code from ExitCode.
	Exit bool `json:"exit"`
	// Error is the error message from command.Wait().
	Error string `json:"error"`
}
