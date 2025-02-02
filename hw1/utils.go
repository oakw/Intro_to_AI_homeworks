package main

// These Queue utilities were created using chatgpt.
// Queue struct to manage BFS traversal
type Queue struct {
	data [][2]int
}

// Enqueue adds an element to the queue
func (q *Queue) Enqueue(val [2]int) {
	q.data = append(q.data, val)
}

// Dequeue removes and returns the first element from the queue
func (q *Queue) Dequeue() ([2]int, bool) {
	if len(q.data) == 0 {
		return [2]int{}, false
	}
	val := q.data[0]
	q.data = q.data[1:]
	return val, true
}

// IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(q.data) == 0
}

// TODO: remove this since it turned out to be unnecessary
func directionArrayToAction(directions [2]int) func(*Agent) {
	return func(agent *Agent) {
		directionsToActions := map[[2]int]func(){
			{-1, 0}: agent.moveUp,
			{1, 0}:  agent.moveDown,
			{0, -1}: agent.moveLeft,
			{0, 1}:  agent.moveRight,
		}

		directionsToActions[directions]()
	}
}
