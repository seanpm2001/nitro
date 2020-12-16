package config

import "fmt"

// Envs is used to map a config to a known environment variable that is used
// on the container instances.
var Envs = map[string]string{
	// PHP specific settings
	"PHP_DISPLAY_ERRORS":      "display_errors",
	"PHP_MEMORY_LIMIT":        "memory_limit",
	"PHP_MAX_EXECUTION_TIME":  "max_execution_time",
	"PHP_UPLOAD_MAX_FILESIZE": "upload_max_filesize",
	"PHP_MAX_INPUT_VARS":      "max_input_vars",
	"PHP_POST_MAX_SIZE":       "post_max_size",
}

// AsEnvs takes a configuration and turns specific options
// such as PHP settings into env vars that can be set on the
// containers environment
func (c *Config) AsEnvs() []string {
	var envs []string

	if c.PHP.DisplayErrors {
		envs = append(envs, "PHP_DISPLAY_ERRORS=on")
	} else {
		envs = append(envs, "PHP_DISPLAY_ERRORS=off")
	}

	if c.PHP.MemoryLimit == "" {
		envs = append(envs, "PHP_MEMORY_LIMIT=512M")
	} else {
		envs = append(envs, "PHP_MEMORY_LIMIT="+c.PHP.MemoryLimit)
	}

	if c.PHP.MaxExecutionTime == 0 {
		envs = append(envs, "PHP_MAX_EXECUTION_TIME=5000")
	} else {
		envs = append(envs, fmt.Sprintf("%s=%d", "PHP_MAX_EXECUTION_TIME", c.PHP.MaxExecutionTime))
	}

	if c.PHP.UploadMaxFileSize == "" {
		envs = append(envs, "PHP_UPLOAD_MAX_FILESIZE=512M")
	} else {
		envs = append(envs, "PHP_UPLOAD_MAX_FILESIZE="+c.PHP.UploadMaxFileSize)
	}

	if c.PHP.MaxInputVars == 0 {
		envs = append(envs, "PHP_MAX_INPUT_VARS=512M")
	} else {
		envs = append(envs, fmt.Sprintf("%s=%d", "PHP_MAX_INPUT_VARS", c.PHP.MaxInputVars))
	}

	if c.PHP.PostMaxSize == "" {
		envs = append(envs, "PHP_POST_MAX_SIZE=512M")
	} else {
		envs = append(envs, fmt.Sprintf("%s=%s", "PHP_POST_MAX_SIZE", c.PHP.PostMaxSize))
	}

	// handle opcache settings
	if c.PHP.OpcacheEnable == true {
		envs = append(envs, "PHP_OPCACHE_ENABLE=1")
	} else {
		envs = append(envs, "PHP_OPCACHE_ENABLE=0")
	}

	if c.PHP.OpcacheRevalidateFreq > 0 {
		envs = append(envs, fmt.Sprintf("PHP_OPCACHE_REVALIDATE_FREQ=%d", c.PHP.OpcacheRevalidateFreq))
	} else {
		envs = append(envs, "PHP_OPCACHE_REVALIDATE_FREQ=0")
	}

	// TODO(jasonmccallister) add opcache settings
	// "PHP_OPCACHE_VALIDATE_TIMESTAMPS=0",
	// "PHP_OPCACHE_MAX_ACCELERATED_FILES=10000",
	// "PHP_OPCACHE_MEMORY_CONSUMPTION=128",
	// "PHP_OPCACHE_MAX_WASTED_PERCENTAGE=10",
	// "PHP_OPCACHE_INTERNED_STRINGS_BUFFER=16",
	// "PHP_OPCACHE_FAST_SHUTDOWN=1"

	// set the blackfire envs if available
	if c.Blackfire.ServerID != "" {
		envs = append(envs, "BLACKFIRE_SERVER_ID="+c.Blackfire.ServerID)
	}
	if c.Blackfire.ServerToken != "" {
		envs = append(envs, "BLACKFIRE_SERVER_TOKEN="+c.Blackfire.ServerToken)
	}

	return envs
}