package mgmt

import (
	"fmt"
	"strconv"

	"github.com/rojbar/rftpis/structs"
)

func state(chStComm structs.ChannelStateComm, trStComm structs.TransmissionStatusComm) {
	channels := make(map[string]*structs.Channel)
	for i := 0; i < 50; i++ {
		channels[strconv.Itoa(i)] = &structs.Channel{Suscribers: 0, Files: make([]string, 0)}
	}
	fmt.Println("STATE INITAL VALUE", channels)
	for {
		select {
		case read := <-chStComm.Read:
			read.Response <- structs.ChannelState{
				Suscribers: channels[read.Alias].Suscribers,
				LastFile:   channels[read.Alias].Files[len(channels[read.Alias].Files)-1],
			}
		case write := <-chStComm.Write:
			if write.Data.LastFile != "" {
				channels[write.Alias].Files = append(channels[write.Alias].Files, write.Data.LastFile)
			}
			if write.Data.Suscribers != 0 {
				channels[write.Alias].Suscribers += write.Data.Suscribers
			}
			write.Response <- true
		default:
		}
		// we inform the transmission manager of new files to be send to each channel
		for key, elem := range channels {
			// we ask the transmission manager if the channel is being currently broadcasting a file
			transmissionStatus := structs.ReadTransmissionStatus{
				Alias:    key,
				Response: make(chan structs.TransmissionStatus),
			}

			trStComm.Read <- transmissionStatus
			isCurrentlyBroadcasting := <-transmissionStatus.Response

			if !isCurrentlyBroadcasting.IsTransfering && !isCurrentlyBroadcasting.IsError && len(elem.Files) != 0 {
				firstFileAdded := elem.Files[0] // get element from queue
				elem.Files = elem.Files[1:]     //removes element from queue
				writeManager := structs.WriteTranmissionStatus{Alias: key, Data: structs.TransmissionStatus{File: firstFileAdded, IsTransfering: false, IsError: false}, Response: make(chan bool)}
				trStComm.Write <- writeManager
				<-writeManager.Response
			}
		}
	}
}
