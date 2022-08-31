package mgmt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/rojbar/rftpis/structs"
)

func channelManagerIO(fileName string, channelName string, trStComm structs.TransmissionStatusComm, chMeComm structs.ChannelMemoryComm) {
	fmt.Println("channelManagerIO", channelName, "sending", fileName)

	file, errO := os.Open("recieved/channels/" + channelName + "/" + fileName)
	if errO != nil {
		print(errO)
		return
	}
	defer file.Close()
	fileInfo, errFS := file.Stat()
	if errFS != nil {
		print(errFS)
		return
	}
	sizeInt := fileInfo.Size()
	size := strconv.Itoa(int(sizeInt))
	extension := fileInfo.Name()

	// we first inform the user we are gonna send a file
	messageGonnaSend := structs.WriteChannelMemory{
		Data: structs.ChannelMemory{
			Data:      []byte("SFTP > 1.0 ACTION: SEND SIZE: " + size + " " + extension + " CHANNEL: " + channelName + ";"),
			IsMessage: true,
		},
		Response: make(chan bool),
	}

	chMeComm.Write <- messageGonnaSend
	<-messageGonnaSend.Response

	//after informing the user we are gonna send a new file we start transfering the data

	reader := bufio.NewReader(file)
	buffer := make([]byte, 4096)
	for {
		_, errP := reader.Read(buffer)
		if errP != nil {
			if errP == io.EOF {
				break
			}
			if errP != nil {
				print(errP)
				//return errP
			}
		}

		//here we send the buffere data we read
		messageGonnaSend = structs.WriteChannelMemory{
			Data: structs.ChannelMemory{
				Data:      buffer,
				IsMessage: false,
			},
			Response: make(chan bool),
		}
		chMeComm.Write <- messageGonnaSend
		<-messageGonnaSend.Response
	}

	// if all alright here we inform the transmissionStatus we are done
	messageTest := structs.WriteTranmissionStatus{
		Alias:    channelName,
		Data:     structs.TransmissionStatus{File: "", IsTransfering: false, IsError: false},
		Response: make(chan bool),
	}

	trStComm.Write <- messageTest
	<-messageTest.Response
}
