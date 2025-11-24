// Package syncs
package syncs

import "sync"

func GroupTasks(tasks ...func()) *sync.WaitGroup {
	var group sync.WaitGroup
	group.Add(len(tasks))
	for _, task := range tasks {
		go func() {
			defer group.Done()
			task()
		}()
	}
	return &group
}

func Tasks(tasks ...func()) {
	groups := GroupTasks(tasks...)
	groups.Wait()
}
