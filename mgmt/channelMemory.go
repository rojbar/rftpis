package mgmt

import (
	"time"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

func channelMemory(alias string, chMeComm structs.ChannelMemoryComm, stChComm structs.ChannelStateComm) {
	utils.Logger.Info("initiated channel memory", zap.String("channel", alias))
	for {
		select {
		case write := <-chMeComm.Write:
			readState := structs.ReadChannelState{
				Alias:    alias,
				Response: make(chan structs.ChannelState),
			}
			stChComm.Read <- readState
			channelState := <-readState.Response
			distributeMessage(channelState, write.Data)
			write.Response <- true
		}
	}
}

func distributeMessage(channelState structs.ChannelState, data structs.ChannelMemory) {
	results := make(chan bool, len(channelState.Suscribers))

	for key := range channelState.Suscribers {
		go func(element structs.Suscriber) {
			responseWriteToClient := make(chan bool)
			timer := time.NewTimer(10000 * time.Second)
			go func() {
				element.Comm.Write <- structs.WriteChannelMemory{
					Data:     data,
					Response: responseWriteToClient,
				}
			}()
			select {
			case <-responseWriteToClient:
				results <- true
			case <-timer.C:
				utils.Logger.Info("coudlnt write to handle subscription on time")
				results <- false
			}
		}(channelState.Suscribers[key])
	}

	for i := 0; i < len(channelState.Suscribers); i++ {
		<-results
	}
}
