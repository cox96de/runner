package util

// MakeEnvPairs converts string maps to key=value pairs.
// exec.Command requires envs to be in key=value format.
func MakeEnvPairs(envMaps ...map[string]string) []string {
	var envs []string
	for _, envMap := range envMaps {
		for k, v := range envMap {
			if k != "" {
				envs = append(envs, k+"="+v)
			}
		}
	}
	return envs
}
