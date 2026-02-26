package thoth

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
)

type Task func() error

func GoAndWait(tasks ...Task) error {
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	for _, t := range tasks {
		f := t
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					errChan <- fmt.Errorf("msg: %v\nstack: %v\n", r, string(debug.Stack()))
				}
			}()

			if err := f(); err != nil {
				errChan <- err
			}

			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
