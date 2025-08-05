// Package syncs
package syncs

import "sync"

func TaskPool(tasks ...func()) {
	var group sync.WaitGroup
	group.Add(len(tasks))
	for _, task := range tasks {
		go func() {
			defer group.Done()
			task()
		}()
	}
	group.Wait()
}
