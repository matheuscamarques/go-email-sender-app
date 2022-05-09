package jsonsender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/jordan-wright/email"
)

type Text string

type JsonSender struct {
	smtpHost string
	smtpPort int
	username string
	password string
	dialer   *email.Email
}

type Payload struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Template string `json:"template"`
	Message  Text   `json:"message"`
}

// Cria um novo JsonSender
// Retorna um erro em caso dos campos serem vazios ou caso o alcance da porta esteja inválido
// Essa função não verifica a validade dos dados inseridos junto ao servidor de email,
// sendo assim em caso de erro nessa parte, ele será informado pela biblioteca nativa log
// ao tentar enviar o email e falhar
func New(smtpHost string, smtpPort int, username string, password string) (*JsonSender, error) {
	if smtpHost == "" {
		return nil, fmt.Errorf("smtpHost can't be empty")
	} else if smtpPort <= 0 || smtpPort > 65535 {
		return nil, fmt.Errorf("smtpPort must be greater than 0 and less than 65535")
	} else if username == "" {
		return nil, fmt.Errorf("username can't be empty")
	} else if password == "" {
		return nil, fmt.Errorf("password can't be empty")
	}

	e := email.NewEmail()

	return &JsonSender{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		dialer:   e,
	}, nil
}

// Inicia o JsonSender
// O update é o tempo periódico entre baixar o json do jsonHost e envia-lo pro sendTo
// ErrorRetry é o delay de cada tentativa que deve ser feita em caso de erro na etapa de envio,
// atualmente ele tenta infinitamente até conseguir enviar o json
// Essa função retorna um erro apenas em caso do JsonSender já estiver previamente inciado,
// qualquer outro tipo de erro é reportado pela biblioteca nativa log
// func (s *JsonSender) Start(jsonPath string) error {
// 	mensagens, err := s.getJsonFile(jsonPath)
// 	if err != nil {
// 		return fmt.Errorf("error getting json: %s", err)
// 	}

// 	for i := range mensagens {
// 		err = s.sendData(mensagens[i])
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	}

// 	return nil
// }

// Para o JsonSender
// Retorna um erro caso o JsonSender já esteja parado
// func (s *JsonSender) Stop() error {
// 	if !s.started {
// 		return fmt.Errorf("JsonSender is not started")
// 	}

// 	s.stop <- true

// 	return nil
// }

func (tx *Text) ParseHTML(templ string) error {
	t, err := template.ParseFiles(templ)
	if err != nil {
		return fmt.Errorf("error parsing template: %s", err)
	}

	buf := bytes.NewBufferString("")

	text := string(*tx)

	data := struct {
		Message string
		Ano     string
	}{
		Message: text,
		Ano:     fmt.Sprint(time.Now().Year()),
	}

	err = t.Execute(buf, data)
	if err != nil {
		return fmt.Errorf("error executing template: %s", err)
	}

	*tx = Text(buf.String())

	return nil
}

func (s *JsonSender) GetJsonFile(jsonPath string) ([]Payload, error) {
	var messages []Payload

	data, err := os.ReadFile(jsonPath)
	if os.IsExist(err) {
		return nil, fmt.Errorf("file not found")
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("invalid json")
	}

	bytes := bytes.NewReader(data)
	err = json.NewDecoder(bytes).Decode(&messages)
	if err != nil {
		return nil, fmt.Errorf("error decoding json")
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages")
	}

	return messages, nil
}

// func (s *JsonSender) getJsonWeb(jsonHost string) (data io.Reader, err error) {
// 	resp, err := http.Get(jsonHost)
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting json: %s", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("error getting json: %s", resp.Status)
// 	}

// 	buffer := new(bytes.Buffer)

// 	_, err = io.Copy(buffer, resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error copying buffer json: %s", err)
// 	}

// 	if !json.Valid(buffer.Bytes()) {
// 		return nil, fmt.Errorf("error validating json: %s", err)
// 	}

// 	return buffer, nil
// }

func (s *JsonSender) Send(messages ...Payload) error {
	if len(messages) == 0 {
		return fmt.Errorf("no messages")
	}

	for i := range messages {
		err := s.sendData(messages[i])
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (s *JsonSender) sendData(payload Payload) error {
	err := payload.Message.ParseHTML(fmt.Sprintf("./template/%s.go.html", payload.Template))
	if err != nil {
		return fmt.Errorf("error parsing template: %s", err)
	}

	s.dialer.From = fmt.Sprintf("Web-Engenharia <%s>", s.username)
	s.dialer.To = []string{payload.To}
	s.dialer.Subject = payload.Subject
	// html, err := insert_varables(message.HTML, messagem)
	s.dialer.HTML = []byte(payload.Message)

	// _, err := s.dialer.Attach(r, filename, "application/json")
	// if err != nil {
	// 	return fmt.Errorf("error attaching file: %s", err)
	// }

	return s.dialer.Send(fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort), smtp.PlainAuth("eMail Sender", s.username, s.password, s.smtpHost))
}
