package handlers

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
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
		print(errCN, errSz, errAtoi)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	extension, errExt := utils.GetKey(message, "EXTENSION")
	if errExt != nil {
		extension = " "
		print(errExt)
	}

	buffer := make([]byte, BUFFERSIZE)
	loops := fileSize / BUFFERSIZE
	sizeLastRead := fileSize % BUFFERSIZE

	lessBuffer := make([]byte, sizeLastRead)

	file, errC := os.Create("recieve/channels/" + channelName + "/" + uuid.NewString() + "." + extension)
	if errC != nil {
		print(errC)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}
	defer file.Close()

	// here we inform the client all ready for recieving file
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		print(errI)
		return
	}

	writer := bufio.NewWriter(file)

	for i := 0; i < int(loops); i++ {
		errRnWf := utils.ReadThenWrite(conn, *writer, buffer)
		if errRnWf != nil {
			print(errRnWf)
			utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
			return
		}
	}
	errRnWf := utils.ReadThenWrite(conn, *writer, lessBuffer)
	if errRnWf != nil {
		print(errRnWf)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}

	//we inform the client that we recieve the file successfully
	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")

	fmt.Println("RECIEVED FILE CORRECTLY, INFORM STATE ", file.Name(), channelName)
	//here we add the file to the queue of files to be send by the server to the channel
	writeChannelState := structs.WriteChannelState{
		Alias:    channelName,
		Data:     structs.ChannelState{Suscribers: 0, LastFile: file.Name()},
		Response: make(chan bool)}
	channelComm.Write <- writeChannelState
	fmt.Println("SENDED STATE")
	<-writeChannelState.Response
}
