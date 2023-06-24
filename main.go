package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/goodplayer/pgproxy/incoming"
	"net"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/goodplayer/pgproxy/api"
	"github.com/goodplayer/pgproxy/inbound_mysql"
)

var (
	paramSnifferLog  string
	unknownPacketLog string

	remoteMySQLAddr    string
	listenMySQLAddr    string
	newListenMysqlAddr string
)

func init() {
	flag.StringVar(&paramSnifferLog, "sniffer", "", "-sniffer=sniffer.log")
	flag.StringVar(&unknownPacketLog, "unknown_log", "", "-unknown_log=sniffer.log")
	flag.StringVar(&remoteMySQLAddr, "remote", "127.0.0.1:3306", "-local=127.0.0.1:3306")
	flag.StringVar(&listenMySQLAddr, "local", "0.0.0.0:23306", "-remote=0.0.0.0:23306")
	flag.StringVar(&newListenMysqlAddr, "newlocal", "0.0.0.0:33306", "-newlocal=0.0.0.0:33306")

	flag.Parse()
}

func main() {
	//TODO default params
	paramSnifferLog = "sniffer.log"
	unknownPacketLog = "unknown_packet.log"

	var handler *incoming.MysqlIncomingHandler
	if h, err := doNewListenMysql(newListenMysqlAddr); err != nil {
		panic(err)
	} else {
		handler = h
	}

	// sniffer handler
	var snifferHandler inbound_mysql.SQLSnifferHandler
	{
		f, err := os.OpenFile(paramSnifferLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("open sniffer file error.", err)
		} else {
			mux := new(sync.Mutex)
			snifferHandler = func(sql string) {
				var obj = api.SnifferElement{SQL: sql}
				j, err := json.Marshal(obj)
				if err != nil {
					panic(err)
				}
				mux.Lock()
				defer mux.Unlock()

				if _, err := f.Write(j); err != nil {
					fmt.Println("write sniffer file error.", err)
					panic(errors.New("write sniffer file error"))
				}
				if _, err := f.WriteString("\r\n"); err != nil {
					fmt.Println("write sniffer file error.", err)
					panic(errors.New("write sniffer file error"))
				}
			}
			defer f.Close()
		}
	}
	// unknown log
	var unknownPacketHandler inbound_mysql.UnknownPacketHandler
	{
		f, err := os.OpenFile(unknownPacketLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("open unknown packet log file error.", err)
		} else {
			mux := new(sync.Mutex)
			unknownPacketHandler = func(packet *inbound_mysql.RawPacket, side inbound_mysql.EndpointSide) {
				mux.Lock()
				defer mux.Unlock()

				if _, err := f.WriteString(fmt.Sprint(side, "|", packet.Data[0], "|", len(packet.Data), "\r\n")); err != nil {
					fmt.Println("write unknown packet log file error.", err)
					panic(errors.New("write unknown packet log file error"))
				}
			}
			defer f.Close()
		}
	}

	fmt.Println("start listening local mysql:", listenMySQLAddr)
	listenAddr, err := net.ResolveTCPAddr("tcp", listenMySQLAddr)
	if err != nil {
		panic(err)
	}
	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		panic(err)
	}

	fmt.Println("start connecting remote mysql:", remoteMySQLAddr)
	destAddr, err := net.ResolveTCPAddr("tcp", remoteMySQLAddr)
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
			var mysqlConn *inbound_mysql.MySQLConn
			if snifferHandler == nil || unknownPacketHandler == nil {
				mysqlConn = inbound_mysql.NewMySQLConn()
			} else {
				mysqlConn = inbound_mysql.NewMySQLConnWithSniffer(snifferHandler, unknownPacketHandler)
			}

			defer tcpConn.Close()

			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(5 * time.Minute)

			destTcpConn, err := net.DialTCP("tcp", nil, destAddr)
			if err != nil {
				panic(err)
			}
			destTcpConn.SetKeepAlive(true)
			destTcpConn.SetKeepAlivePeriod(5 * time.Minute)

			defer destTcpConn.Close()

			err = mysqlConn.HandleConnPhase(tcpConn, destTcpConn)
			if err != nil {
				fmt.Println("handle conn phase error:", err)
				return
			}

			go func() {
				for {
					rawPacket := new(inbound_mysql.RawPacket)
					err := rawPacket.ReadRawPacket(destTcpConn)
					if err != nil {
						tcpConn.Close()
						destTcpConn.Close()
						fmt.Println("dest -> source: read error:", err)
						return
					}

					err = mysqlConn.HandleCommandPhaseServerPacket(rawPacket)
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
				rawPacket := new(inbound_mysql.RawPacket)
				err := rawPacket.ReadRawPacket(tcpConn)
				if err != nil {
					tcpConn.Close()
					destTcpConn.Close()
					fmt.Println("source -> dest: read error:", err)
					return
				}

				err = mysqlConn.HandleCommandPhaseClientPacket(rawPacket)
				if err != nil {
					tcpConn.Close()
					destTcpConn.Close()
					fmt.Println("source -> dest: write error:", err)
					continue // read buf until empty
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

	runtime.KeepAlive(handler)
}

func doNewListenMysql(addr string) (*incoming.MysqlIncomingHandler, error) {
	pgbackend, err := incoming.NewPgBackend("postgres://admin:admin@localhost:5432/wordpress")
	if err != nil {
		return nil, err
	}
	in := incoming.NewMysqlIncomingHandler(addr, pgbackend)
	return in, in.Startup()
}
