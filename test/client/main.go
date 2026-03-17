package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"openclaw-go/internal/messages"
	"openclaw-go/internal/protocol"
)

type pingReq struct {
	ClientTime int64 `json:"client_time"`
}

type echoReq struct {
	Text string `json:"text"`
}

type createCharacterReq struct {
	PlayerID string `json:"player_id"`
	Name     string `json:"name"`
}

type getCharacterReq struct {
	PlayerID string `json:"player_id"`
}

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "gateway address")
	mode := flag.String("mode", "ping", "ping | echo | create-player | get-player | interactive")
	text := flag.String("text", "hello from test client", "echo text when mode=echo")
	playerID := flag.String("player-id", "player_001", "player id for character api")
	name := flag.String("name", "claw", "character name for create-player mode")
	count := flag.Int("count", 1, "send count for ping/echo mode")
	interval := flag.Duration("interval", 500*time.Millisecond, "send interval for ping/echo mode")
	flag.Parse()

	logger := log.New(os.Stdout, "[test-client] ", log.LstdFlags|log.Lmicroseconds)
	conn, err := net.DialTimeout("tcp", *addr, 3*time.Second)
	if err != nil {
		logger.Fatalf("connect failed: %v", err)
	}
	defer conn.Close()

	logger.Printf("connected to %s", *addr)

	done := make(chan struct{})
	go readLoop(conn, logger, done)

	switch strings.ToLower(*mode) {
	case "ping":
		sendPing(conn, logger, *count, *interval)
	case "echo":
		sendEcho(conn, logger, *text, *count, *interval)
	case "create-player":
		sendCreatePlayer(conn, logger, *playerID, *name)
	case "get-player":
		sendGetPlayer(conn, logger, *playerID)
	case "interactive":
		runInteractive(conn, logger)
	default:
		logger.Fatalf("unknown mode: %s", *mode)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}

func sendPing(conn net.Conn, logger *log.Logger, count int, interval time.Duration) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		payload, err := protocol.MarshalJSON(pingReq{ClientTime: time.Now().UnixMilli()})
		if err != nil {
			logger.Printf("marshal ping failed: %v", err)
			return
		}
		if err := sendPacket(conn, protocol.Packet{MsgID: messages.MsgPingReq, Payload: payload}); err != nil {
			logger.Printf("send ping failed: %v", err)
			return
		}
		logger.Printf("[send] msg_id=%d payload=%s", messages.MsgPingReq, prettyBytes(payload))
		if i < count-1 {
			time.Sleep(interval)
		}
	}
}

func sendEcho(conn net.Conn, logger *log.Logger, text string, count int, interval time.Duration) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		payload, err := protocol.MarshalJSON(echoReq{Text: text})
		if err != nil {
			logger.Printf("marshal echo failed: %v", err)
			return
		}
		if err := sendPacket(conn, protocol.Packet{MsgID: messages.MsgEchoReq, Payload: payload}); err != nil {
			logger.Printf("send echo failed: %v", err)
			return
		}
		logger.Printf("[send] msg_id=%d payload=%s", messages.MsgEchoReq, prettyBytes(payload))
		if i < count-1 {
			time.Sleep(interval)
		}
	}
}

func runInteractive(conn net.Conn, logger *log.Logger) {
	logger.Println("interactive mode started, type text and press Enter to send echo request")
	logger.Println("input /ping, /create <player_id> <name>, /get <player_id>, /quit")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				logger.Printf("stdin read failed: %v", err)
			}
			return
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "/quit" {
			return
		}
		if line == "/ping" {
			sendPing(conn, logger, 1, 0)
			continue
		}
		if strings.HasPrefix(line, "/create ") {
			parts := strings.Fields(line)
			if len(parts) < 3 {
				logger.Println("usage: /create <player_id> <name>")
				continue
			}
			sendCreatePlayer(conn, logger, parts[1], strings.Join(parts[2:], " "))
			continue
		}
		if strings.HasPrefix(line, "/get ") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				logger.Println("usage: /get <player_id>")
				continue
			}
			sendGetPlayer(conn, logger, parts[1])
			continue
		}
		sendEcho(conn, logger, line, 1, 0)
	}
}

func sendCreatePlayer(conn net.Conn, logger *log.Logger, playerID string, name string) {
	payload, err := protocol.MarshalJSON(createCharacterReq{
		PlayerID: playerID,
		Name:     name,
	})
	if err != nil {
		logger.Printf("marshal create-player failed: %v", err)
		return
	}
	if err := sendPacket(conn, protocol.Packet{MsgID: messages.MsgCreateCharacterReq, Payload: payload}); err != nil {
		logger.Printf("send create-player failed: %v", err)
		return
	}
	logger.Printf("[send] msg_id=%d payload=%s", messages.MsgCreateCharacterReq, prettyBytes(payload))
}

func sendGetPlayer(conn net.Conn, logger *log.Logger, playerID string) {
	payload, err := protocol.MarshalJSON(getCharacterReq{
		PlayerID: playerID,
	})
	if err != nil {
		logger.Printf("marshal get-player failed: %v", err)
		return
	}
	if err := sendPacket(conn, protocol.Packet{MsgID: messages.MsgGetCharacterReq, Payload: payload}); err != nil {
		logger.Printf("send get-player failed: %v", err)
		return
	}
	logger.Printf("[send] msg_id=%d payload=%s", messages.MsgGetCharacterReq, prettyBytes(payload))
}

func readLoop(conn net.Conn, logger *log.Logger, done chan<- struct{}) {
	defer close(done)

	reader := bufio.NewReader(conn)
	for {
		packet, err := protocol.DecodePacket(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Println("connection closed by server")
				return
			}
			logger.Printf("decode packet failed: %v", err)
			return
		}
		logger.Printf("[recv] msg_id=%d payload=%s", packet.MsgID, prettyBytes(packet.Payload))
	}
}

func sendPacket(conn net.Conn, packet protocol.Packet) error {
	data, err := protocol.EncodePacket(packet)
	if err != nil {
		return err
	}

	written := 0
	for written < len(data) {
		n, werr := conn.Write(data[written:])
		if werr != nil {
			return werr
		}
		written += n
	}
	return nil
}

func prettyBytes(payload []byte) string {
	if len(payload) == 0 {
		return "<empty>"
	}

	var out bytes.Buffer
	if err := json.Indent(&out, payload, "", "  "); err == nil {
		return out.String()
	}
	if isPrintable(payload) {
		return string(payload)
	}
	return "0x" + hex.EncodeToString(payload)
}

func isPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return true
}
