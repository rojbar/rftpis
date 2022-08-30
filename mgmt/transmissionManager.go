package mgmt

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/structs"
)

func transmissionManager(trStComm structs.TransmissionStatusComm, chMeComm map[string]structs.ChannelMemoryComm) {
	channels := make(map[string]*structs.TransmissionStatus)
	for i := 0; i < 50; i++ {
		channels[strconv.Itoa(i)] = &structs.TransmissionStatus{File: "", IsTransfering: false, IsError: false}
	}

	for {
		select {
		case read := <-trStComm.Read:
			read.Response <- structs.TransmissionStatus{
				File:          channels[read.Alias].File,
				IsTransfering: channels[read.Alias].IsTransfering,
				IsError:       channels[read.Alias].IsError,
			}
		case write := <-trStComm.Write:
			channels[write.Alias] = &write.Data
			write.Response <- true
		default:
			// we start a channel manager, the channel manager should
			// inform when it has finished, in case it fails whe should retry and after some retrys we should dropped it or inform state
			for key, channel := range channels {
				if !channel.IsTransfering && !channel.IsError && channel.File != "" {
					channel.IsTransfering = true
					fmt.Println("iniciando channel Manager en status A", channel, uuid.NewString())
					go channelManagerIO(channel.File, key, trStComm, chMeComm[key])

				}
				if !channel.IsTransfering && channel.IsError {
					channel.IsError = false
					channel.IsTransfering = true
					fmt.Println("iniciando channel Manager en status B", channel, uuid.NewString())
					go channelManagerIO(channel.File, key, trStComm, chMeComm[key])
				}
			}
		}
	}
}
