package api

import (
	"cf"
	"cf/configuration"
	"code.google.com/p/go.net/websocket"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"sort"
	"time"
)

type LogsRepository interface {
	RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message)) (err error)
	TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), printInterval time.Duration) (err error)
}

type LoggregatorLogsRepository struct {
	config       *configuration.Configuration
	endpointRepo EndpointRepository
}

func NewLoggregatorLogsRepository(config *configuration.Configuration, endpointRepo EndpointRepository) (repo LoggregatorLogsRepository) {
	repo.config = config
	repo.endpointRepo = endpointRepo
	return
}

func (repo LoggregatorLogsRepository) RecentLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message)) (err error) {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		err = errors.New(apiResponse.Message)
		return
	}
	location := host + fmt.Sprintf("/dump/?app=%s", app.Guid)
	return repo.connectToWebsocket(location, app, onConnect, onMessage, nil)
}

func (repo LoggregatorLogsRepository) TailLogsFor(app cf.Application, onConnect func(), onMessage func(*logmessage.Message), printInterval time.Duration) error {
	host, apiResponse := repo.endpointRepo.GetEndpoint(cf.LoggregatorEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return errors.New(apiResponse.Message)
	}
	location := host + fmt.Sprintf("/tail/?app=%s", app.Guid)
	return repo.connectToWebsocket(location, app, onConnect, onMessage, time.Tick(printInterval*time.Second))
}

func (repo LoggregatorLogsRepository) connectToWebsocket(location string, app cf.Application, onConnect func(), onMessage func(*logmessage.Message), tickerChan <-chan time.Time) (err error) {
	const EOF_ERROR = "EOF"

	config, err := websocket.NewConfig(location, "http://localhost")
	if err != nil {
		return
	}

	config.Header.Add("Authorization", repo.config.AccessToken)
	config.TlsConfig = &tls.Config{InsecureSkipVerify: true}

	ws, err := websocket.DialConfig(config)
	if err != nil {
		return
	}

	onConnect()

	msgChan := make(chan *logmessage.Message, 1000)
	errChan := make(chan error, 0)

	go repo.listenForMessages(ws, msgChan, errChan)
	go repo.sendKeepAlive(ws)

	sortableMsg := &sortableLogMessages{}

Loop:
	for {
		select {
		case err = <-errChan:
			break Loop
		case msg, ok := <-msgChan:
			if !ok {
				break Loop
			}
			sortableMsg.Messages = append(sortableMsg.Messages, msg)
		case <-tickerChan:
			invokeCallbackWithSortedMessages(sortableMsg, onMessage)
			sortableMsg.Messages = []*logmessage.Message{}
		}
		if err != nil {
			break
		}
	}

	if tickerChan == nil {
		invokeCallbackWithSortedMessages(sortableMsg, onMessage)
	}

	if err.Error() == EOF_ERROR {
		err = nil
	}

	return
}

func invokeCallbackWithSortedMessages(messages *sortableLogMessages, callback func(*logmessage.Message)) {
	sort.Sort(messages)
	for _, msg := range messages.Messages {
		callback(msg)
	}
}

func (repo LoggregatorLogsRepository) sendKeepAlive(ws *websocket.Conn) {
	for {
		websocket.Message.Send(ws, "I'm alive!")
		time.Sleep(25 * time.Second)
	}
}

func (repo LoggregatorLogsRepository) listenForMessages(ws *websocket.Conn, msgChan chan<- *logmessage.Message, errChan chan<- error) {
	var err error
	defer close(msgChan)
	for {
		var data []byte
		err = websocket.Message.Receive(ws, &data)
		if err != nil {
			errChan <- err
			break
		}

		msg, msgErr := logmessage.ParseMessage(data)
		if msgErr != nil {
			continue
		}
		msgChan <- msg
	}
}

type sortableLogMessages struct {
	Messages []*logmessage.Message
}

func (sort *sortableLogMessages) Len() int {
	return len(sort.Messages)
}

func (sort *sortableLogMessages) Less(i, j int) bool {
	msgI := sort.Messages[i]
	msgJ := sort.Messages[j]
	return *msgI.GetLogMessage().Timestamp < *msgJ.GetLogMessage().Timestamp
}

func (sort *sortableLogMessages) Swap(i, j int) {
	sort.Messages[i], sort.Messages[j] = sort.Messages[j], sort.Messages[i]
}
