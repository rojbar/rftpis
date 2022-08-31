package mgmt

import (
	"fmt"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

func channelMemory(alias string, chMeComm structs.ChannelMemoryComm) {
	messages := utils.NewQueue[structs.ChannelMemory](3)
	fmt.Println("channelMemory", alias, "current memory", messages)
	for {
		select {
		case read := <-chMeComm.Read:
			mes, errMes := messages.Retrieve()
			if errMes != nil {
				print(errMes)
				mes = structs.ChannelMemory{Data: make([]byte, 0), IsMessage: false}
			}
			read.Response <- mes
		case write := <-chMeComm.Write:
			errMes := messages.Enqueue(write.Data)
			if errMes != nil {
				messages.Dequeue()
				messages.Enqueue(write.Data)
			}
			write.Response <- true
		}
	}
}
