package match

import (
	"github.com/google/uuid"
	"plane_war/internal/model"
	"sync"
)

// MatchQueue 匹配队列
type MatchQueue struct {
	queue []*model.Player
	lock  sync.Mutex
}

var MatchQueueInstance = &MatchQueue{
	queue: make([]*model.Player, 0),
}

func (mq *MatchQueue) AddPlayer(p *model.Player) *model.Room {
	mq.lock.Lock()
	defer mq.lock.Unlock()

	mq.queue = append(mq.queue, p)

	//如果队列 >=2 就创建房间
	if len(mq.queue) >= 2 {
		p1 := mq.queue[0]
		p2 := mq.queue[1]
		mq.queue = mq.queue[2:]
		room := &model.Room{
			ID:      uuid.New().String(),
			Players: []*model.Player{p1, p2},
		}
		return room
	}
	return nil
}
