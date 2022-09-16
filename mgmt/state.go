package mgmt

import (
	"strconv"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

func state(chStComm structs.ChannelStateComm, trStComm structs.TransmissionStatusComm) {
	channels := make(map[string]*structs.Channel)
	for i := 0; i < 50; i++ {
		channels[strconv.Itoa(i)] = &structs.Channel{Id: strconv.Itoa(i), Suscribers: make(map[string]*structs.Suscriber), Files: make([]string, 0)}
	}

	utils.Logger.Info("state initiated with x channels")

	for {
		select {
		case read := <-chStComm.Read:
			aux := make(map[string]structs.Suscriber)
			for key, element := range channels[read.Alias].Suscribers {
				aux[key] = structs.Suscriber{
					Id:   element.Id,
					Comm: element.Comm,
				}
			}
			read.Response <- structs.ChannelState{
				Id:         read.Alias,
				Suscribers: aux,
				Files:      channels[read.Alias].Files,
			}

		case write := <-chStComm.Write:

			if write.AddSuscriber {
				channels[write.Alias].Suscribers[write.AddSuscriberData.Id] = &write.AddSuscriberData
			}
			if write.RemoveSuscriber != "" {
				delete(channels[write.Alias].Suscribers, write.RemoveSuscriber)
			}
			if write.AddFile != "" {
				channels[write.Alias].Files = append(channels[write.Alias].Files, write.AddFile)
			}
			if write.RemoveFile != "" {
				aux := -1
				for key, element := range channels[write.Alias].Files {
					if element == write.RemoveFile {
						aux = key
						break
					}
				}
				if aux != -1 {
					channels[write.Alias].Files = append(channels[write.Alias].Files[:aux], channels[write.Alias].Files[aux+1:]...)
				}
			}
			write.Response <- true
		default:
		}
		// we inform the transmission manager of new files to be send to each channel
		for key, elem := range channels {
			if len(elem.Suscribers) == 0 {
				continue
			}
			// we ask the transmission manager if the channel is being currently broadcasting a file
			transmissionStatus := structs.ReadTransmissionStatus{
				Alias:    key,
				Response: make(chan structs.TransmissionStatus),
			}

			trStComm.Read <- transmissionStatus
			transmissionManagerState := <-transmissionStatus.Response

			if !transmissionManagerState.IsTransfering && !transmissionManagerState.IsError && len(elem.Files) != 0 {
				utils.Logger.Info("informing transmission manager of new file to be send to the channel")
				firstFileAdded := elem.Files[0] // get element from queue
				elem.Files = elem.Files[1:]     //removes element from queue
				writeManager := structs.WriteTranmissionStatus{Alias: key, Data: structs.TransmissionStatus{File: firstFileAdded, IsTransfering: false, IsError: false}, Response: make(chan bool)}
				trStComm.Write <- writeManager
				<-writeManager.Response
			}
		}
	}
}
