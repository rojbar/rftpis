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
			distributeMessage(channelState, write.Data, stChComm, alias)
			write.Response <- true
		}
	}
}

func distributeMessage(channelState structs.ChannelState, data structs.ChannelMemory, stChComm structs.ChannelStateComm, alias string) {
	results := make(chan bool, len(channelState.Suscribers))

	for key := range channelState.Suscribers {
		go func(element structs.Suscriber) {
			responseWriteToClient := make(chan bool)
			timer := time.NewTimer(200 * time.Millisecond)
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
				//we remove the handle subscription from the state
				updateState := structs.WriteChannelState{
					Alias:           alias,
					AddSuscriber:    false,
					RemoveSuscriber: element.Id,
					AddFile:         "",
					RemoveFile:      "",
					Response:        make(chan bool),
				}
				stChComm.Write <- updateState
				<-updateState.Response
				results <- false
			}
		}(channelState.Suscribers[key])
	}

	for i := 0; i < len(channelState.Suscribers); i++ {
		<-results
	}
}
