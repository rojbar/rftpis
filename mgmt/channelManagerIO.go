package mgmt

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

const BUFFERSIZE = 4096

func channelManagerIO(fileName string, channelName string, trStComm structs.TransmissionStatusComm, chMeComm structs.ChannelMemoryComm) {
	utils.Logger.Info("initiated channel manager with", zap.String("channel", channelName), zap.String("file", fileName))

	file, errO := os.Open(fileName)
	if errO != nil {
		utils.Logger.Error(errO.Error())
		//inform of error the tranmission manager
		return
	}
	defer file.Close()
	fileInfo, errFS := file.Stat()
	if errFS != nil {
		utils.Logger.Error(errFS.Error())
		//inform of error the tranmission manager
		return
	}

	sizeInt := fileInfo.Size()
	size := strconv.Itoa(int(sizeInt))
	extension := fileInfo.Name()
	ext := "EXTENSION: "
	_, after, found := strings.Cut(extension, ".")
	if found {
		ext = "EXTENSION: " + after
	}

	// we first inform the user we are gonna send a file
	messageNewFile := []byte("SFTP > 1.0 ACTION: SEND SIZE: " + size + " " + ext + " CHANNEL: " + channelName + ";")
	utils.Logger.Info("channel manager sending message new file to memory", zap.String("channel", channelName), zap.String("file", fileName), zap.String("message", string(messageNewFile)))
	messageGonnaSend := structs.WriteChannelMemory{
		Data: structs.ChannelMemory{
			Data:      messageNewFile,
			IsMessage: true,
			Count:     -1,
			Id:        fileName,
			IsEOF:     false,
		},
		Response: make(chan bool),
	}
	writeInMemoryUntilOk(chMeComm, messageGonnaSend)

	//after informing the user we are gonna send a new file we start transfering the data

	reader := bufio.NewReader(file)
	buffer := make([]byte, BUFFERSIZE)
	chunks, sizeLastChunk := utils.CalculateChunksToSendExactly(int(sizeInt), BUFFERSIZE)
	remainderBuffer := make([]byte, sizeLastChunk)

	loops := chunks
	if sizeLastChunk != 0 {
		loops += 1
	}

	utils.Logger.Info("channel manager ready for sending file to memory", zap.String("channel", channelName), zap.String("file", fileName), zap.Int("loops", loops))
	for i := 0; i < loops; i++ {
		auxBuffer := buffer
		isEof := false
		if i == chunks {
			auxBuffer = remainderBuffer
			isEof = true
		}
		_, errP := reader.Read(auxBuffer)
		if errP != nil {
			if errP == io.EOF {
				utils.Logger.Info("channel manager read EOF prematurly", zap.String("channel", channelName), zap.String("file", fileName))
				//infor of error the tranmission manager
				break
			}
			if errP != nil {
				utils.Logger.Error(errP.Error())
				//inform of error the tranmission manager
				return
			}
		}

		//here we send the buffered data we read
		utils.Logger.Info(
			"channel manager sending chunk of data to memory",
			zap.String("channel", channelName),
			zap.String("file", fileName),
			zap.Int("chunk", i),
			zap.Int("chunk size", len(auxBuffer)),
			zap.Bool("isEof", isEof),
		)
		messageGonnaSend = structs.WriteChannelMemory{
			Data: structs.ChannelMemory{
				Data:      auxBuffer,
				IsMessage: false,
				Count:     i,
				Id:        fileName,
				IsEOF:     isEof,
			},
			Response: make(chan bool),
		}
		writeInMemoryUntilOk(chMeComm, messageGonnaSend)
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

// this is here because the queue in chanel memory can become full so we should wait
// buffered channels are not recommended as implementations of queue so i strive for this solution
func writeInMemoryUntilOk(chMeComm structs.ChannelMemoryComm, messageGonnaSend structs.WriteChannelMemory) {
	for {
		chMeComm.Write <- messageGonnaSend
		ok := <-messageGonnaSend.Response
		if ok {
			break
		}
	}
}
