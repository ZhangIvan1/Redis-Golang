package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Pair[L, R any] func() (L, R)

func NewPair[L, R any](left L, right R) Pair[L, R] {
	return func() (L, R) {
		return left, right
	}
}

const (
	DEFAULT_TYPE string = "tcp"
	DEFAULT_HOST string = "0.0.0.0"
	DEFAULT_PORT string = "6379"

	DEFAULT_MASTER_HOST string = ""
	DEFAULT_MASTER_PORT string = ""

	DEFAULT_POOL_SIZE int = 30
)

type Config struct {
	netType string
	host    string
	port    string

	masterHost string
	masterPort string

	poolSize int
}

func (config Config) String() string {
	return fmt.Sprintf("{netType: %s, host: %s, port: %s, masterHost: %s, masterPort: %s}",
		config.netType, config.host, config.port, config.masterHost, config.masterPort)
}

const (
	MASTER string = "master"
	SLAVE  string = "slave"
)

type Redis struct {
	config Config

	listener net.Listener

	store          map[string]string
	timestamp      map[string]time.Time
	timeExpiration map[string]time.Duration

	role             string
	replicationSet   []Slave
	masterReplId     string
	masterReplOffset int
	masterHost       string
	masterPort       string
	masterConn       net.Conn

	connectionPool ConnectionPool

	commandChan chan Pair[Command, net.Conn]
	writeChan   chan Pair[Command, net.Conn]
	readChan    chan Pair[Command, net.Conn]
	sendChan    chan Pair[string, net.Conn]
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	args := os.Args

	config := Config{
		netType:    DEFAULT_TYPE,
		host:       DEFAULT_HOST,
		port:       DEFAULT_PORT,
		masterHost: DEFAULT_MASTER_HOST,
		masterPort: DEFAULT_MASTER_PORT,
		poolSize:   DEFAULT_POOL_SIZE,
	}
	for i, arg := range args {
		if arg == "--port" {
			config.port = args[i+1]
		}
		if arg == "--replicaof" {
			config.masterHost = args[i+1]
			config.masterPort = args[i+2]
		}
	}

	fmt.Println("config:", config)

	rd := Make(config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	fmt.Println("Shutting down the server...")

	if err := rd.listener.Close(); err != nil {
		fmt.Println("Error closing redis server:", err)
	}
}

func Make(config Config) *Redis {
	rd := &Redis{}
	rd.config = config
	rd.store = make(map[string]string)
	rd.timestamp = make(map[string]time.Time)
	rd.timeExpiration = make(map[string]time.Duration)
	rd.connectionPool = ConnectionPool{conns: make([]net.Conn, 0), capacity: config.poolSize}

	if config.masterHost == "" {
		rd.role = MASTER
		hash := sha1.New()
		hash.Write([]byte(config.String()))
		rd.masterReplId = fmt.Sprintf("%x", hash.Sum(nil))
		rd.replicationSet = make([]Slave, 0)

	} else {
		rd.role = SLAVE
		rd.masterHost = config.masterHost
		rd.masterPort = config.masterPort
	}

	rd.masterReplOffset = 0

	if listener, err := net.Listen(config.netType, config.host+":"+config.port); err != nil {
		log.Fatalln("Failed to bind to port", config.port, err)
	} else {
		rd.listener = listener
	}

	go rd.sendTicker(rd.sendChan)
	go rd.writeTicker(rd.writeChan)
	go rd.readTicker(rd.readChan)
	go rd.commandClassifyTicker(rd.commandChan)
	go rd.handleConnectionTicker(rd.commandChan)
	go rd.listenConnectionTicker()

	if rd.role == SLAVE {
		go rd.handshakeTicker()
	}

	return rd
}
