package rftpis

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/mgmt"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

const BUFFERSIZE = 4096

// OK
func Server() {
	channelComm, channelMemoryComm := mgmt.Init()
	for i := 0; i < 50; i++ {
		err := os.MkdirAll("recieve/channels/"+strconv.Itoa(i), 0750)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}
	}

	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, errA := ln.Accept()
		if errA != nil {
			print(errA)
			continue
		}
		go recieve(conn, channelComm, channelMemoryComm)
	}
}

//OK
func recieve(conn net.Conn, chComm structs.ChannelStateComm, chMeComm map[string]structs.ChannelMemoryComm) {
	// here we recieve the request
	message, errR := utils.ReadMessage(conn)
	if errR != nil {
		print(errR)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		conn.Close()
		return
	}

	// here we parse
	action, errA := utils.GetKey(message, "ACTION")
	if errA != nil {
		print(errA)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	channel, errCh := utils.GetKey(message, "CHANNEL")
	if errCh != nil {
		print(errCh)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	//redirect according to client action
	switch action {
	case "SEND":
		go handleRecieveFile(conn, message, chComm)

	case "SUBSCRIBE":
		go handleSubscription(conn, message, chComm, chMeComm[channel])

	default:
		conn.Close()
	}
}

//OK
func handleRecieveFile(conn net.Conn, message string, channelComm structs.ChannelStateComm) {
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

	//we inform the client that we recieve the client successfully
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

//NOT OK
func handleSubscription(conn net.Conn, message string, chStComm structs.ChannelStateComm, channelMemory structs.ChannelMemoryComm) {
	defer conn.Close()
	// here we inform he is subscribed
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		print(errI)
		return
	}
	//notify state that a client has subscribe and wait for the next file tobe send
}
