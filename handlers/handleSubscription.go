package handlers

import (
	"bufio"
	"fmt"
	"net"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

//OK
func HandleSubscription(conn net.Conn, message string, chStComm structs.ChannelStateComm, chanMem structs.ChannelMemoryComm) {
	defer conn.Close()
	fmt.Println("RECIEVED SUBSCRIPTION", message)
	channelName, _ := utils.GetKey(message, "CHANNEL")
	updateState := structs.WriteChannelState{
		Alias:    channelName,
		Data:     structs.ChannelState{Suscribers: 1, LastFile: ""},
		Response: make(chan bool),
	}
	chStComm.Write <- updateState
	<-updateState.Response
	fmt.Println("UPDATED STATE", message)

	// here we inform he is subscribed
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		print(errI)
		return
	}

	//we read the channel

	for {
		readMessage := structs.ReadChannelMemory{Response: make(chan structs.ChannelMemory)}
		chanMem.Read <- readMessage
		memoryMessage := <-readMessage.Response

		if memoryMessage.IsMessage {
			fmt.Println("Sending message")
			utils.SendMessage(conn, string(memoryMessage.Data))
			continue

		}
		writer := bufio.NewWriter(conn)

		buffer := make([]byte, BUFFERSIZE)
		for {
			_, errW := writer.Write(buffer)
			if errW != nil {
				print(errW)
			}
			errF := writer.Flush()
			if errF != nil {
				print(errF)
			}
		}
	}
}
