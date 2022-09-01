package mgmt

import (
	"strconv"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

func transmissionManager(trStComm structs.TransmissionStatusComm, chMeComm map[string]structs.ChannelMemoryComm) {
	channels := make(map[string]*structs.TransmissionStatus)
	for i := 0; i < 50; i++ {
		channels[strconv.Itoa(i)] = &structs.TransmissionStatus{File: "", IsTransfering: false, IsError: false}
	}

	utils.Logger.Info("transmission manager initiated, managing x channels")

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
		}
		// we start a channel manager, the channel manager should
		// inform when it has finished, in case it fails whe should retry and after some retrys we should dropped it or inform state
		for key, channel := range channels {
			if !channel.IsTransfering && !channel.IsError && channel.File != "" {
				channel.IsTransfering = true
				utils.Logger.Info("initiating channel manager with", zap.String("channel:", key), zap.String("File:", channel.File))
				go channelManagerIO(channel.File, key, trStComm, chMeComm[key])

			}
			if !channel.IsTransfering && channel.IsError {
				channel.IsError = false
				channel.IsTransfering = true
				utils.Logger.Info("initiating channel manager for retry with", zap.String("channel:", key), zap.String("File:", channel.File))
				go channelManagerIO(channel.File, key, trStComm, chMeComm[key])
			}
		}
	}
}
