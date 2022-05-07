package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	listenAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:23306")
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	destAddr, err := net.ResolveTCPAddr("tcp", "192.168.31.231:3306")
	if err != nil {
		panic(err)
	}

	for {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		fmt.Println("new incoming connection:", tcpConn)

		go func() {
			mysqlConn := NewMysSQLConn()

			defer tcpConn.Close()

			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(5 * time.Minute)

			destTcpConn, err := net.DialTCP("tcp", nil, destAddr)
			if err != nil {
				panic(err)
			}
			destTcpConn.SetKeepAlive(true)
			destTcpConn.SetKeepAlivePeriod(5 * time.Minute)

			go func() {
				for {
					rawPacket := new(RawPacket)
					err := rawPacket.ReadRawPacket(destTcpConn)
					if err != nil {
						tcpConn.Close()
						destTcpConn.Close()
						fmt.Println("dest -> source: read error:", err)
						return
					}

					err = mysqlConn.HandleServerPacket(rawPacket)
					if err != nil {
						tcpConn.Close()
						destTcpConn.Close()
						fmt.Println("dest -> source: parse error:", err)
						continue // read buf until empty
					}

					err = rawPacket.WriteRawPacket(tcpConn)
					if err != nil {
						tcpConn.Close()
						destTcpConn.Close()
						fmt.Println("dest -> source: write error:", err)
						continue // read buf until empty
					}
				}
			}()

			for {
				rawPacket := new(RawPacket)
				err := rawPacket.ReadRawPacket(tcpConn)
				if err != nil {
					tcpConn.Close()
					destTcpConn.Close()
					fmt.Println("source -> dest: read error:", err)
					return
				}

				if mysqlConn.GetCurrentState() == StateCommand {
					err = mysqlConn.HandleClientPacket(rawPacket)
					if err != nil {
						tcpConn.Close()
						destTcpConn.Close()
						fmt.Println("source -> dest: write error:", err)
						continue // read buf until empty
					}
				}

				err = rawPacket.WriteRawPacket(destTcpConn)
				if err != nil {
					tcpConn.Close()
					destTcpConn.Close()
					fmt.Println("source -> dest: write error:", err)
					continue // read buf until empty
				}
			}
		}()
	}
}
