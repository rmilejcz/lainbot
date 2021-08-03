package framework

import "sync"

var wg sync.WaitGroup

type SongQueue struct {
	list    []Song
	current *Song
	Running bool
}

func (queue SongQueue) Get() []Song {
	return queue.list
}

func (queue *SongQueue) Set(list []Song) {
	queue.list = list
}

func (queue *SongQueue) Add(song Song) {
	queue.list = append(queue.list, song)
}

func (queue SongQueue) HasNext() bool {
	return len(queue.list) > 0
}

func (queue *SongQueue) Next() Song {
	song := queue.list[0]
	queue.list = queue.list[1:]
	queue.current = &song
	return song
}

func (queue *SongQueue) Clear() {
	queue.list = make([]Song, 0)
	//queue.Running = false
	queue.current = nil
}

func (queue *SongQueue) Start(sess *Session, callback func(string)) {
	wg.Add(1)
	queue.Running = true
	for queue.HasNext() && queue.Running {
		println(queue.HasNext(), "in a loop of sorts?")
		song := queue.Next()
		callback("Now playing `" + song.Title + "`.")
		sess.Play(song)
		println("done playing song")
	}

	if !queue.Running {
		callback("Stopped playing.")
	} else {
		callback("Finished queue.")
	}

}

func (q *SongQueue) Play(sess *Session, callback func(string)) {
	if q.HasNext() {
		song := q.Next()
		callback("Now playing `" + song.Title + "`.")
		sess.Play(song)
	}
}

func (queue *SongQueue) Current() *Song {
	return queue.current
}

func (queue *SongQueue) Pause() {
	queue.Running = false
}

func newSongQueue() *SongQueue {
	queue := new(SongQueue)
	queue.list = make([]Song, 0)
	return queue
}
