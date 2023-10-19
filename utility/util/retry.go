package util

import "time"

func Retry(num int, sleep time.Duration, fn func() (any, error)) (any, error) {

	var err error
	for i := 0; i < num; i++ {
		if res, err := fn(); err == nil {
			return res, nil
		}
		time.Sleep(sleep)
	}

	return nil, err
}
