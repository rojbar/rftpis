package mgmt

import (
	"time"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

func channelMemory(alias string, chMeComm structs.ChannelMemoryComm) {
	messages := utils.NewQueue[structs.ChannelMemory](3)
	toDequeque := make(chan bool)
	utils.Logger.Info("initiated channel memory", zap.String("channel", alias))

	for {
		select {
		case read := <-chMeComm.Read:
			mes, errMes := messages.Retrieve()
			if errMes != nil {
				//print("esta vacia!", errMes)
				mes = structs.ChannelMemory{Data: make([]byte, 0), IsMessage: false, Count: 0, Id: "EMPTY MEMORY", IsEOF: false}
			}
			read.Response <- mes
		case write := <-chMeComm.Write:
			utils.Logger.Debug("channel memory recieved write message", zap.String("channel", alias), zap.Any("data", write.Data))
			errMes := messages.Enqueue(write.Data)
			if errMes != nil {
				utils.Logger.Debug("channel memory recieved write message failed queue full", zap.String("channel", alias))
				write.Response <- false
				break
			}

			aux, _ := messages.Retrieve()
			if aux.Id == write.Data.Id && aux.Count == write.Data.Count {
				//the package we enqueue is at the top so we should start the timer for its removal
				go dequeueAfterTime(toDequeque)
			}
			write.Response <- true

		case <-toDequeque:
			// we dequeue
			messages.Dequeue()
			// we add a timer for the new message in head, only if the queue is not empty!
			_, errRet := messages.Retrieve()
			if errRet == nil {
				go dequeueAfterTime(toDequeque)
			}
		}
	}
}

func dequeueAfterTime(toDequeque chan bool) {
	timer := time.NewTicker(2 * time.Second)
	<-timer.C
	toDequeque <- true
}
