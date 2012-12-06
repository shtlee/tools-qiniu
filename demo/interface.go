package main 

type Notifier interface {
    BlockNotify(int, int)
    ChunkNotify(int, string)
}
