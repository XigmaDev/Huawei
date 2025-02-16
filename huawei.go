package huawei

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Huawei struct {
	IP     string
	client *http.Client
	token  string
}

type ErrorResponse struct {
	XMLName xml.Name
}

type SMS struct {
	Smstat   string `xml:"Smstat"`
	Index    string `xml:"Index"`
	Phone    string `xml:"Phone"`
	Content  string `xml:"Content"`
	Date     string `xml:"Date"`
	Sca      string `xml:"Sca"`
	SaveType string `xml:"SaveType"`
	Priority string `xml:"Priority"`
	SmsType  string `xml:"SmsType"`
}

type SMSListResponse struct {
	XMLName  xml.Name `xml:"response"`
	Count    int      `xml:"Count"`
	Messages []SMS    `xml:"Messages>Message"`
}

type SMSCountResponse struct {
	XMLName      xml.Name `xml:"response"`
	LocalUnread  string   `xml:"LocalUnread"`
	LocalInbox   string   `xml:"LocalInbox"`
	LocalOutbox  string   `xml:"LocalOutbox"`
	LocalDraft   string   `xml:"LocalDraft"`
	LocalDeleted string   `xml:"LocalDeleted"`
	SimUnread    string   `xml:"SimUnread"`
	SimInbox     string   `xml:"SimInbox"`
	SimOutbox    string   `xml:"SimOutbox"`
	SimDraft     string   `xml:"SimDraft"`
	LocalMax     string   `xml:"LocalMax"`
	SimMax       string   `xml:"SimMax"`
}

type ConnectionStatusResponse struct {
	XMLName              xml.Name `xml:"response"`
	ConnectionStatus     string   `xml:"ConnectionStatus"`
	SignalStrength       string   `xml:"SignalStrength"`
	SignalIcon           string   `xml:"SignalIcon"`
	CurrentNetworkType   string   `xml:"CurrentNetworkType"`
	CurrentServiceDomain string   `xml:"CurrentServiceDomain"`
	RoamingStatus        string   `xml:"RoamingStatus"`
	BatteryStatus        string   `xml:"BatteryStatus"`
	BatteryLevel         string   `xml:"BatteryLevel"`
	SimlockStatus        string   `xml:"SimlockStatus"`
	WanIPAddress         string   `xml:"WanIPAddress"`
	PrimaryDNS           string   `xml:"PrimaryDns"`
	SecondaryDNS         string   `xml:"SecondaryDns"`
	CurrentWifiUser      string   `xml:"CurrentWifiUser"`
	TotalWifiUser        string   `xml:"TotalWifiUser"`
	ServiceStatus        string   `xml:"ServiceStatus"`
	SimStatus            string   `xml:"SimStatus"`
	WifiStatus           string   `xml:"WifiStatus"`
}

type Response struct {
	XMLName xml.Name `xml:"response"`
	Status  string   `xml:",chardata"`
}

func NewHuawei(ip string) *Huawei {
	return &Huawei{
		IP:     ip,
		client: &http.Client{},
		token:  "",
	}
}

// Login authenticates a user with the provided username and password.
// It encodes the password in base64, constructs an XML payload, and sends
// a POST request to the Huawei API endpoint for user login. The function
// sets necessary headers, including a request verification token obtained
// from the GetToken method. It reads and parses the XML response to determine
// the login status and returns an error if authentication fails.
//
// Parameters:
//   - username: The username for authentication.
//   - password: The password for authentication.
//
// Returns:
//   - error: An error if the login process fails, otherwise nil.
func (h *Huawei) Login(username, password string) error {

	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

	payload := fmt.Sprintf(`
			<request>
				<Username>%s</Username>
				<Password>%s</Password>
			</request>`,
		username, encodedPassword)

	req, err := http.NewRequest("POST", h.IP+"/api/user/login", bytes.NewReader([]byte(payload)))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	if err := h.GetToken(); err != nil {
		return err
	}
	req.Header.Set("__RequestVerificationToken", h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}
	var res Response
	if err := xml.Unmarshal(body, &res); err != nil {
		log.Fatalf("Error parsing XML: %v", err)
	}

	fmt.Printf("Login Status: %s\n", res.Status)

	if isErrorResponse(body) {
		return fmt.Errorf("authentication failed")
	}
	return nil
}

// GetToken retrieves a token from the Huawei web server.
// It sends a GET request to the /api/webserver/token endpoint and parses the XML response to extract the token.
// The token is then stored in the Huawei struct's token field.
// Returns an error if the request fails, the response cannot be read, or the XML cannot be unmarshaled.
func (h *Huawei) GetToken() error {
	req, err := http.NewRequest("GET", h.IP+"/api/webserver/token", nil)
	if err != nil {
		return err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp struct {
		XMLName xml.Name `xml:"response"`
		Token   string   `xml:"token"`
	}

	if err := xml.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	h.token = tokenResp.Token
	return nil
}

// sendRequest sends an HTTP request to the specified URL with the given method and payload.
// It sets necessary headers and includes a request verification token.
//
// Parameters:
//   - method: The HTTP method to use (e.g., "GET", "POST").
//   - url: The URL to send the request to.
//   - payload: The payload to include in the request body.
//
// Returns:
//   - []byte: The response body as a byte slice.
//   - error: An error if the request fails or if there is an issue reading the response body.
func (h *Huawei) sendRequest(method, url, payload string) ([]byte, error) {
	req, err := http.NewRequest(method, h.IP+url, strings.NewReader(xmlEscape(payload)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")

	if err := h.GetToken(); err != nil {
		return nil, err
	}
	req.Header.Set("__RequestVerificationToken", h.token)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// isErrorResponse checks if the provided XML body contains an error response.
// It unmarshals the XML body into a struct and verifies if the root element is "error".
// If the unmarshalling fails, it returns false.
// If the root element is "error", it returns true.
//
// Parameters:
//
//	body []byte - The XML body to check.
//
// Returns:
//
//	bool - True if the XML body contains an error response, false otherwise.
func isErrorResponse(body []byte) bool {
	var errResp struct {
		XMLName xml.Name `xml:"error"`
		Code    string   `xml:"code"`
	}
	if err := xml.Unmarshal(body, &errResp); err != nil {
		return false
	}
	fmt.Print(errResp)

	//TODO:error response handlig based on error code

	return errResp.XMLName.Local == "error"
}

// Connect establishes a connection to the Huawei device by sending a POST request
// to the /api/dialup/dial endpoint with a predefined payload. It returns an error
// if the request fails or if the response indicates a failure.
//
// Returns:
//   - error: An error object if the connection attempt fails, otherwise nil.
func (h *Huawei) Connect() error {
	payload := "<request><Action>1</Action></request>"
	body, err := h.sendRequest("POST", "/api/dialup/dial", payload)
	if err != nil {
		return err
	}
	if isErrorResponse(body) {
		return fmt.Errorf("connection failed")
	}
	return nil
}

// Disconnect sends a request to disconnect the Huawei device.
// It constructs an XML payload with the action to disconnect,
// sends the request to the "/api/dialup/dial" endpoint, and
// checks the response for errors.
//
// Returns an error if the request fails or if the response
// indicates a failure to disconnect.
func (h *Huawei) Disconnect() error {
	payload := "<request><Action>0</Action></request>"
	body, err := h.sendRequest("POST", "/api/dialup/dial", payload)
	if err != nil {
		return err
	}
	if isErrorResponse(body) {
		return fmt.Errorf("disconnection failed")
	}
	return nil
}

// SendSMS sends an SMS message to a specified phone number using the Huawei API.
// It constructs an XML payload with the message details and sends an HTTP POST request.
//
// Parameters:
//   - msg: The message content to be sent.
//   - phone: The recipient's phone number.
//
// Returns:
//   - error: An error if the SMS sending fails, otherwise nil.
func (h *Huawei) SendSMS(msg, phone string) error {
	url := "/api/sms/send-sms"
	date := time.Now().Format("2006-01-02 15:04:05")
	payload := fmt.Sprintf(`
			<request>
			<Index>-1</Index>
				<Phones>
					<Phone>%s</Phone>
				</Phones>
				<Sca></Sca>
				<Content>%s</Content>
				<Length>%d</Length>
				<Reserved>1</Reserved>
				<Date>%s</Date>
			</request>`,
		xmlEscape(phone), xmlEscape(msg), len(msg), xmlEscape(date))

	client := &http.Client{}
	req, err := http.NewRequest("POST", h.IP+url, strings.NewReader(payload))

	if err != nil {
		fmt.Println(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("__RequestVerificationToken", h.token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("%s OK", phone)

	// body, err := h.sendRequest("POST", "/api/sms/send-sms", payload)
	// if err != nil {
	// 	return err
	// }
	// if isErrorResponse(body) {
	// 	return fmt.Errorf("failed to sending to %s", phone)
	// }
	// fmt.Println("%s Sending OK", phone)

	return nil
}

// GetSmsCount retrieves the count of SMS messages from the Huawei device.
// It sends a GET request to the /api/sms/sms-count endpoint and parses the response.
//
// Returns a slice of strings containing the counts of various SMS categories:
// - LocalUnread: Count of unread messages in local storage.
// - LocalInbox: Count of messages in the local inbox.
// - LocalOutbox: Count of messages in the local outbox.
// - LocalDraft: Count of draft messages in local storage.
// - LocalDeleted: Count of deleted messages in local storage.
// - SimUnread: Count of unread messages on the SIM card.
// - SimInbox: Count of messages in the SIM card inbox.
// - SimOutbox: Count of messages in the SIM card outbox.
// - SimDraft: Count of draft messages on the SIM card.
// - LocalMax: Maximum number of messages that can be stored locally.
// - SimMax: Maximum number of messages that can be stored on the SIM card.
//
// Returns an error if the request fails or the response cannot be parsed.
func (h *Huawei) GetSmsCount() ([]string, error) {
	body, err := h.sendRequest("GET", "/api/sms/sms-count", "")
	if err != nil {
		return nil, err
	}
	if isErrorResponse(body) {
		return nil, fmt.Errorf("get SMS count failed")
	}

	var resp SMSCountResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return []string{
		resp.LocalUnread,
		resp.LocalInbox,
		resp.LocalOutbox,
		resp.LocalDraft,
		resp.LocalDeleted,
		resp.SimUnread,
		resp.SimInbox,
		resp.SimOutbox,
		resp.SimDraft,
		resp.LocalMax,
		resp.SimMax,
	}, nil
}

// GetSmsList retrieves a list of SMS messages from the Huawei device.
// It sends a POST request to the /api/sms/sms-list endpoint with the necessary payload.
// The function returns a slice of SMS messages and an error if any occurred during the process.
//
// Returns:
//   - []SMS: A slice of SMS messages retrieved from the device.
//   - error: An error if the request failed or the response could not be unmarshaled.
func (h *Huawei) GetSmsList() ([]SMS, error) {
	payload := `<request>
		<PageIndex>1</PageIndex>
		<ReadCount>20</ReadCount>
		<BoxType>1</BoxType>
		<SortType>0</SortType>
		<Ascending>0</Ascending>
		<UnreadPreferred>0</UnreadPreferred>
	</request>`

	body, err := h.sendRequest("POST", "/api/sms/sms-list", payload)
	if err != nil {
		return nil, err
	}
	if isErrorResponse(body) {
		return nil, fmt.Errorf("get SMS list failed")
	}

	var resp SMSListResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Messages, nil
}

// DeleteSMS deletes an SMS message from the Huawei device by its index.
// It sends a POST request to the "/api/sms/delete-sms" endpoint with the specified index.
//
// Parameters:
//   - index: The index of the SMS message to be deleted.
//
// Returns:
//   - error: An error if the request fails or if the response indicates a failure.
func (h *Huawei) DeleteSMS(index int) error {
	payload := fmt.Sprintf("<request><Index>%d</Index></request>", index)
	body, err := h.sendRequest("POST", "/api/sms/delete-sms", payload)
	if err != nil {
		return err
	}
	if isErrorResponse(body) {
		return fmt.Errorf("delete SMS failed")
	}
	return nil
}

// GetConnectionStatus retrieves the current connection status from the Huawei device.
// It sends a GET request to the /api/monitoring/status endpoint and parses the XML response.
//
// Returns a slice of strings containing various connection status details, or an error if the request fails or the response cannot be parsed.
//
// The returned slice contains the following elements in order:
// - ConnectionStatus
// - SignalStrength
// - SignalIcon
// - CurrentNetworkType
// - CurrentServiceDomain
// - RoamingStatus
// - BatteryStatus
// - BatteryLevel
// - SimlockStatus
// - WanIPAddress
// - PrimaryDNS
// - SecondaryDNS
// - CurrentWifiUser
// - TotalWifiUser
// - ServiceStatus
// - SimStatus
// - WifiStatus
//
// Returns:
// - []string: A slice of strings containing the connection status details.
// - error: An error if the request fails or the response cannot be parsed.
func (h *Huawei) GetConnectionStatus() ([]string, error) {
	body, err := h.sendRequest("GET", "/api/monitoring/status", "")
	if err != nil {
		return nil, err
	}
	if isErrorResponse(body) {
		return nil, fmt.Errorf("get connection status failed")
	}

	var resp ConnectionStatusResponse
	if err := xml.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return []string{
		resp.ConnectionStatus,
		resp.SignalStrength,
		resp.SignalIcon,
		resp.CurrentNetworkType,
		resp.CurrentServiceDomain,
		resp.RoamingStatus,
		resp.BatteryStatus,
		resp.BatteryLevel,
		resp.SimlockStatus,
		resp.WanIPAddress,
		resp.PrimaryDNS,
		resp.SecondaryDNS,
		resp.CurrentWifiUser,
		resp.TotalWifiUser,
		resp.ServiceStatus,
		resp.SimStatus,
		resp.WifiStatus,
	}, nil
}

// IsConnected checks the connection status of the Huawei device.
// It returns true if the device is connected (status code "901"), otherwise false.
// If there is an error retrieving the connection status, it returns false along with the error.
func (h *Huawei) IsConnected() (bool, error) {
	status, err := h.GetConnectionStatus()
	if err != nil {
		return false, err
	}
	if len(status) == 0 {
		return false, fmt.Errorf("no connection status received")
	}
	return status[0] == "901", nil
}

// xmlEscape escapes special characters in a string for safe inclusion in XML.
// It converts characters like <, >, &, ', and " to their corresponding XML escape sequences.
//
// Parameters:
//
//	s - The input string that needs to be escaped.
//
// Returns:
//
//	A string with special characters replaced by their XML escape sequences.
func xmlEscape(s string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(s))
	return b.String()
}
