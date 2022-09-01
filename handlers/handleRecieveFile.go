package handlers

import (
	"bufio"
	"net"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

const BUFFERSIZE = 4096

//OK
func HandleRecieveFile(conn net.Conn, message string, channelComm structs.ChannelStateComm) {
	defer conn.Close()

	//here we check message data
	channelName, errCN := utils.GetKey(message, "CHANNEL")
	value, errSz := utils.GetKey(message, "SIZE")
	fileSize, errAtoi := strconv.Atoi(value)
	if errCN != nil || errSz != nil || errAtoi != nil || fileSize <= 0 {
		utils.Logger.Error("handle recieve file error in message attributes", zap.String("channelAtrr", channelName), zap.String("sizeAttr", errSz.Error()+errAtoi.Error()))
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	extension, errExt := utils.GetKey(message, "EXTENSION")
	if errExt != nil {
		extension = " "
		utils.Logger.Info("handle recieve file warsning extension not provided", zap.String("extensionAttr", errExt.Error()))
	}

	file, errC := os.Create("recieve/channels/" + channelName + "/" + uuid.NewString() + "." + extension)
	if errC != nil {
		utils.Logger.Error(errC.Error())
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}
	defer file.Close()

	// here we inform the client all ready for recieving file
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		utils.Logger.Error(errI.Error())
		return
	}

	buffer := make([]byte, BUFFERSIZE)
	chunks, sizeLastChunk := utils.CalculateChunksToSendExactly(fileSize, BUFFERSIZE)
	remainderBuffer := make([]byte, sizeLastChunk)

	loops := chunks
	if sizeLastChunk != 0 {
		loops += 1
	}

	writer := bufio.NewWriter(file)

	for i := 0; i < loops; i++ {
		auxBuffer := buffer
		if i == chunks {
			auxBuffer = remainderBuffer
		}

		errRnWf := utils.ReadThenWrite(conn, *writer, auxBuffer)
		if errRnWf != nil {
			utils.Logger.Error(errRnWf.Error())
			utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
			return
		}
	}
	//we inform the client that we recieve the file successfully
	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	utils.Logger.Info("handl recieve file recivied file correctly informing state", zap.String("channel", channelName), zap.String("file", file.Name()))
	//here we add the file to the queue of files to be send by the server to the channel
	writeChannelState := structs.WriteChannelState{
		Alias:    channelName,
		Data:     structs.ChannelState{Suscribers: 0, LastFile: file.Name()},
		Response: make(chan bool)}
	channelComm.Write <- writeChannelState
	<-writeChannelState.Response
}
