package imageserver

type ReqQueue struct {
	ch     chan []byte
	length int32
}

func NewReqQueue() *ReqQueue {
	return &ReqQueue{
		ch:     make(chan []byte),
		length: 0,
	}
}
