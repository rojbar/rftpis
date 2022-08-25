package rftpis

import (
	"bufio"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

const BUFFERSIZE = 4096

// OK
func Server() {
	reads := make(chan structs.ReadFromChannel)
	writes := make(chan structs.WriteToChannel)
	go state(reads, writes)

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
		go recieve(conn, reads, writes)
	}
}

func state(reads chan structs.ReadFromChannel, writes chan structs.WriteToChannel) {
	channels := make(map[string]*structs.Channel)

	for i := 0; i < 50; i++ {
		channels[strconv.Itoa(i)] = &structs.Channel{Suscribers: 0, Files: make([]string, 0)}
	}

	for {
		select {
		case read := <-reads:
			read.Response <- structs.ChannelState{Suscribers: channels[read.Alias].Suscribers, LastFile: channels[read.Alias].Files[len(channels[read.Alias].Files)-1]}
		case write := <-writes:
			if write.File != "" {
				channels[write.Alias].Files = append(channels[write.Alias].Files, write.File)
			}
			if write.Suscriber != 0 {
				channels[write.Alias].Suscribers += write.Suscriber
			}
			write.Response <- true
		}
	}
}

//OK
func recieve(conn net.Conn, reads chan structs.ReadFromChannel, writes chan structs.WriteToChannel) {
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

	//redirect according to client action
	switch action {
	case "SEND":
		go handleRecieveFile(conn, message, reads, writes)

	case "SUBSCRIBE":
		go handleSubscription(conn, message)

	default:
		conn.Close()
	}
}

//NOT OK
func handleSubscription(conn net.Conn, message string) {
	defer conn.Close()
	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
}

//OK
func handleRecieveFile(conn net.Conn, message string, reads chan structs.ReadFromChannel, writes chan structs.WriteToChannel) {
	defer conn.Close()

	//here we check message data
	channelName, errCN := utils.GetKey(message, "CHANNEL")
	if errCN != nil {
		print(errCN)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	value, errSz := utils.GetKey(message, "SIZE")
	if errSz != nil {
		print(errSz)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		return
	}

	fileSize, errAtoi := strconv.Atoi(value)
	if errAtoi != nil || fileSize <= 0 {
		print(errAtoi)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
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

	file, errC := os.Create(uuid.NewString() + "." + extension)
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
		errRnWf := readNetWriteFile(conn, buffer, *writer)
		if errRnWf != nil {
			print(errRnWf)
			utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
			return
		}
	}
	errRnWf := readNetWriteFile(conn, lessBuffer, *writer)
	if errRnWf != nil {
		print(errRnWf)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}

	//we inform the client that we recieve the client successfully
	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")

	//here we add the file to the queue of files to be send by the server to the channel
	writeToChannel := structs.WriteToChannel{Alias: channelName, Suscriber: 0, File: file.Name(), Response: make(chan bool)}
	writes <- writeToChannel
	<-writeToChannel.Response
}

func readNetWriteFile(conn net.Conn, buffer []byte, writer bufio.Writer) error {
	_, errR := io.ReadFull(conn, buffer)
	if errR != nil {
		if errR == io.EOF {
			print(errR)
		}
		return errR
	}
	_, errW := writer.Write(buffer)
	if errW != nil {
		return errW
	}
	errF := writer.Flush()
	if errF != nil {
		return errF
	}

	return nil
}
