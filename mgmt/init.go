package mgmt

import (
	"strconv"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

func Init() (structs.ChannelStateComm, map[string]structs.ChannelMemoryComm) {
	utils.Logger.Info("initiating server state")
	chComm := structs.ChannelStateComm{
		Read:  make(chan structs.ReadChannelState),
		Write: make(chan structs.WriteChannelState),
	}

	trStComm := structs.TransmissionStatusComm{
		Read:  make(chan structs.ReadTransmissionStatus),
		Write: make(chan structs.WriteTranmissionStatus),
	}

	chMeComm := make(map[string]structs.ChannelMemoryComm)
	for i := 0; i < 50; i++ {
		chMeComm[strconv.Itoa(i)] = structs.ChannelMemoryComm{
			Read:  make(chan structs.ReadChannelMemory),
			Write: make(chan structs.WriteChannelMemory),
		}
		go channelMemory(strconv.Itoa(i), chMeComm[strconv.Itoa(i)], chComm)
	}

	go state(chComm, trStComm)
	go transmissionManager(trStComm, chMeComm)

	utils.Logger.Info("server state initiated")
	return chComm, chMeComm
}
