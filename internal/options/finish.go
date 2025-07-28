package options

import (
	"errors"
	"os"
	"fmt"
)
type FinishOptions struct {
    Certificate 	string
	Key        		string
	PfxFile			string
	Password		string
	Chain			bool
	IncludeRoot		bool
}

func (opts* FinishOptions) Validate() error {
	if opts.Certificate == "" {
		return errors.New("Certificate file (-c) is required")
	}
	if opts.Key == "" {
		return errors.New("Key file (-k) is required")
	}
	if opts.Password == "" {
		fmt.Printf("Found environment variable of %s\n", os.Getenv("sslpass"))
		if os.Getenv("sslpass") == "" {
			return errors.New("Password is not set")
		}
		opts.Password = os.Getenv("sslpass")
	}
	return nil
}
