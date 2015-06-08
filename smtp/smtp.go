package smtp

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/textproto"
	"strings"

	"github.com/gchaincl/antimail/models"
)

type Session struct {
	msg *models.Message

	end     bool
	reader  *bufio.Reader
	scanner *bufio.Scanner
	writer  *bufio.Writer
}

type command struct {
	action    string
	arguments []string
}

func parseLine(line string) *command {
	fields := strings.Fields(line)
	return &command{
		action:    strings.ToUpper(fields[0]),
		arguments: fields[1:],
	}
}

func parseAddress(line string) string {
	addr := strings.Split(line, ":")[1]
	return addr[1 : len(addr)-1]
}

func (s *Session) handle(line string) (err error) {
	cmd := parseLine(line)

	switch cmd.action {
	case "EHLO", "HELO":
		err = s.handleEHLO(cmd)
	case "MAIL":
		err = s.handleMAIL(cmd)
	case "RCPT":
		err = s.handleRCPT(cmd)
	case "DATA":
		err = s.handleDATA(cmd)
	case "QUIT":
		err = s.handleQUIT(cmd)
	default:
		err = fmt.Errorf("Unknown command: %s %q\n", cmd.action, cmd.arguments)
	}

	return err
}

func (s *Session) close() {
	s.end = true
}

func (s *Session) reply(status int, message string) (err error) {
	line := fmt.Sprintf("%d %s\r\n", status, message)
	_, err = fmt.Fprintf(s.writer, line)
	if err != nil {
		return
	}
	err = s.writer.Flush()
	if err != nil {
		return
	}

	log.Printf("-> %s", line)
	return
}

func (s *Session) handleEHLO(cmd *command) error {
	return s.reply(250, "Hola Mundo")
}

func (s *Session) handleMAIL(cmd *command) error {
	s.msg.From = parseAddress(cmd.arguments[0])
	return s.reply(250, "Go ahead")
}

func (s *Session) handleRCPT(cmd *command) error {
	s.msg.Rcpt = parseAddress(cmd.arguments[0])
	return s.reply(250, "Go ahead")
}

func (s *Session) handleDATA(cmd *command) error {
	s.reply(354, "Go ahead. End your data with <CR><LF>.<CR><LF>")
	s.msg.Data = &bytes.Buffer{}
	reader := textproto.NewReader(s.reader).DotReader()
	_, err := io.CopyN(s.msg.Data, reader, int64(10240000))

	if err == io.EOF {
		log.Printf("->\n %s\n", s.msg.Data)
		return s.reply(250, "Thank you.")
	}

	return err
}

func (s *Session) handleQUIT(cmd *command) error {
	err := s.reply(221, "OK, bye")
	s.close()
	return err
}

func (s *Session) Run() (*models.Message, error) {
	s.reply(220, "localhost ESMTPD ready.")
	for s.scanner.Scan() {
		line := s.scanner.Text()
		log.Printf("<- %s", line)
		err := s.handle(line)
		if err != nil {
			return nil, err
		}

		if s.end == true {
			return s.msg, nil
		}
	}

	return nil, nil
}

func NewSession(reader io.Reader, writer io.Writer) *Session {
	return &Session{
		msg:     &models.Message{},
		reader:  bufio.NewReader(reader),
		scanner: bufio.NewScanner(reader),
		writer:  bufio.NewWriter(writer),
	}
}