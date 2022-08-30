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
		//return errO
	}
	defer file.Close()

	fileInfo, errFS := file.Stat()
	if errFS != nil {
		print(errFS)
		//return errFS
	}
	sizeInt := fileInfo.Size()
	size := strconv.Itoa(int(sizeInt))
	extension := fileInfo.Name()

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

	}

	// if all alright here we inform the transmissionStatus
	messageTest := structs.WriteTranmissionStatus{
		Alias:    channelName,
		Data:     structs.TransmissionStatus{File: "", IsTransfering: false, IsError: false},
		Response: make(chan bool),
	}

	trStComm.Write <- messageTest
	<-messageTest.Response
}
