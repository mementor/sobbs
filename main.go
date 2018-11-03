package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	mediaURL string
	bulkURL  string
)

type message struct {
	user           string
	pass           string
	text           string
	from           string
	label          string
	PTransactionID string
	sendingMethod  string
	buttonText     string
	buttonLink     string
	imageID        string
	phones         []string
	groupID        string
}

type bulkResp struct {
	Code    int    `xml:"code"`
	Message string `xml:"tech_message"`
}

type mediaResp struct {
	ImageID string `json:"image_id"`
	Status  int    `json:"status"`
	Error   string `json:"error"`
}

func sendMsg(msg *message) {
	form := url.Values{
		"txt":  {msg.text},
		"user": {msg.user},
		"from": {msg.from},
	}

	if msg.sendingMethod != "" {
		form.Set("sending_method", msg.sendingMethod)
	}
	if msg.buttonText != "" {
		form.Set("button_text", msg.buttonText)
	}
	if msg.buttonLink != "" {
		form.Set("button_link", msg.buttonLink)
	}
	if msg.imageID != "" {
		form.Set("image_id", msg.imageID)
	}
	if msg.groupID != "" {
		form.Set("group_id", msg.groupID)
	}
	if msg.label != "" {
		form.Set("label", msg.label)
	}
	if msg.PTransactionID != "" {
		form.Set("p_transaction_id", msg.PTransactionID)
	}

	reqTime := time.Now()

	phonesSign := strings.Join(msg.phones, "")
	signString := fmt.Sprintf("%s%s%s%s%s", msg.user, msg.from, phonesSign, msg.text, msg.pass)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(signString)))

	form.Set("sign", sign)

	for _, phoneOne := range msg.phones {
		form.Add("phone", phoneOne)
	}

	resp, errHTTP := http.PostForm(bulkURL, form)

	respTime := time.Now()

	lag := respTime.Sub(reqTime)

	if errHTTP != nil {
		for _, phone := range msg.phones {
			fmt.Printf("%s;%s;%s;%s;error: %s\n", phone, reqTime.String(), respTime.String(), lag, errHTTP)
		}
		return
	}
	defer resp.Body.Close()
	var parsedResp bulkResp
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bodyBytes, &parsedResp)
	if parsedResp.Code == 0 && parsedResp.Message == "OK" {
		for _, phone := range msg.phones {
			fmt.Printf("%s;%s;%s;%s;%d;%s\n", phone, reqTime.String(), respTime.String(), lag, parsedResp.Code, parsedResp.Message)
		}
	} else if parsedResp.Message != "OK" {
		for _, phone := range msg.phones {
			fmt.Printf("%s;%s;%s;%s;%d;%s\n", phone, reqTime.String(), respTime.String(), lag, parsedResp.Code, parsedResp.Message)
		}
	} else {
		for _, phone := range msg.phones {
			fmt.Printf("%s;%s;%s;%s;%s\n", phone, reqTime.String(), respTime.String(), lag, bodyBytes)
		}
	}
}

func worker(wg *sync.WaitGroup, msgChan chan *message, exitChan chan bool) {
	fmt.Fprintln(os.Stderr, "Worker up")

	for {
		select {
		case msg := <-msgChan:
			sendMsg(msg)
			wg.Done()
		case <-exitChan:
			wg.Done()
			return
		}
	}
}

func uploadImage(imagePath, user, pass string) (imgID string) {
	var uploadResp mediaResp

	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	hash := md5.New()
	if _, errIO := io.Copy(hash, file); errIO != nil {
		log.Fatal(err)
	}
	md5file := fmt.Sprintf("%x", hash.Sum(nil))
	hash.Reset()

	toSign := fmt.Sprintf("%s%s%s", user, md5file, pass)
	io.WriteString(hash, toSign)
	sign := fmt.Sprintf("%x", hash.Sum(nil))

	params := map[string]string{
		"sign":  sign,
		"login": user,
	}
	req, errUR := newfileUploadRequest(mediaURL, params, "image", imagePath)
	if errUR != nil {
		log.Fatal(errUR)
		return
	}

	client := &http.Client{}
	resp, errHTTP := client.Do(req)
	if errHTTP != nil {
		log.Fatal(errHTTP)
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	errJSON := json.Unmarshal(bodyBytes, &uploadResp)
	if errJSON != nil {
		log.Fatal(errJSON)
		return
	}
	if uploadResp.Status == 0 && uploadResp.ImageID != "" {
		imgID = uploadResp.ImageID
	} else if uploadResp.Error != "" {
		fmt.Printf("Image upload failed: %s\n", uploadResp.Error)
	}

	return
}

// Creates a new file upload http request with optional extra params
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func main() {

	var user string
	var pass string
	var text string
	var from string
	var buttonLink string
	var buttonText string
	var imageID string
	var sendingMethod string
	var imageFile string
	var groupID string

	var threads int
	var batchSize int
	var label string
	var ptransactionID string

	sentCounter := 0

	flag.StringVar(&user, "user", "", "Bulk API user")
	flag.StringVar(&pass, "pass", "", "Bulk API pass")
	flag.StringVar(&text, "text", "", "text to send")
	flag.StringVar(&from, "from", "", "text to send from")

	flag.StringVar(&buttonLink, "buttonlink", "", "Link on button click")
	flag.StringVar(&buttonText, "buttontext", "", "Text on button")
	flag.StringVar(&imageID, "imageid", "", "Image ID loaded at media.sms-online.com")
	flag.StringVar(&sendingMethod, "sendingmethod", "", "Sending method")
	flag.StringVar(&imageFile, "imagefile", "", "Image filepath")
	flag.StringVar(&groupID, "groupid", "", "ID for the dispatch")

	flag.StringVar(&bulkURL, "bulkurl", "https://bulk.sms-online.com/", "Bulk API URL")
	flag.StringVar(&mediaURL, "mediaurl", "https://media.sms-online.com/upload/", "Media API URL")
	flag.IntVar(&batchSize, "batchsize", 10, "Number of phones in one http request")
	flag.IntVar(&threads, "threads", 1, "Parallel threads")
	flag.StringVar(&label, "label", "", "Label in message")
	flag.StringVar(&ptransactionID, "p_transaction_id", "", "PTransactionID for message")

	flag.Parse()

	if imageFile != "" && imageID == "" {
		imageID = uploadImage(imageFile, user, pass)
		if imageID == "" {
			return
		}
		fmt.Fprintf(os.Stderr, "image_id: %s\n", imageID)
	}

	msgChan := make(chan *message, 1)
	exitChan := make(chan bool, 1)

	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		go worker(&wg, msgChan, exitChan)
	}

	inPhones := make([]string, 0, batchSize)
	counter := 1

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		inPhones = append(inPhones, scanner.Text())
		if counter < batchSize {
			counter++
		} else {
			wg.Add(1)
			msg := &message{
				user:   user,
				pass:   pass,
				text:   text,
				from:   from,
				phones: inPhones,
			}
			if label != "" {
				msg.label = label
			}
			if ptransactionID != "" {
				msg.PTransactionID = ptransactionID
			}
			if sendingMethod != "" {
				msg.sendingMethod = sendingMethod
			}
			if buttonText != "" {
				msg.buttonText = buttonText
			}
			if buttonLink != "" {
				msg.buttonLink = buttonLink
			}
			if imageID != "" {
				msg.imageID = imageID
			}
			if groupID != "" {
				msg.groupID = groupID
			}
			msgChan <- msg
			sentCounter += counter
			fmt.Fprintf(os.Stderr, "Sent: %d\n", sentCounter)
			inPhones = make([]string, 0, batchSize)
			counter = 1
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading standard input: %s\n", err)
	}
	if counter > 1 {
		wg.Add(1)
		msg := &message{
			user:   user,
			pass:   pass,
			text:   text,
			from:   from,
			phones: inPhones,
		}

		if label != "" {
			msg.label = label
		}
		if ptransactionID != "" {
			msg.PTransactionID = ptransactionID
		}
		if sendingMethod != "" {
			msg.sendingMethod = sendingMethod
		}
		if buttonText != "" {
			msg.buttonText = buttonText
		}
		if buttonLink != "" {
			msg.buttonLink = buttonLink
		}
		if imageID != "" {
			msg.imageID = imageID
		}
		if groupID != "" {
			msg.groupID = groupID
		}
		msgChan <- msg
	}
	fmt.Fprintln(os.Stderr, "Done!")

	wg.Wait()

	fmt.Fprintln(os.Stderr, "Cleaning...")
	close(exitChan)
	wg.Add(threads)
	wg.Wait()
}
