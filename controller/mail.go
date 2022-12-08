package controller

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"

	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

type Mail struct {
	Subject string          `json:"subject"`
	Body    string          `json:"body"`
	From    []*mail.Address `json:"from"`
	To      []*mail.Address `json:"to"`
	Date    time.Time       `json:"data"`
}

func CreateMail(context *gin.Context) {
	//firstname := c.DefaultQuery("emailname", "Guest")
}

func GetMails(context *gin.Context) {
	log.Println("##in GetMails")
	imap.CharsetReader = charset.Reader

	emailname := context.DefaultQuery("emailname", "test1")

	mailhost := os.Getenv("MAIL_HOST")
	mailport := os.Getenv("MAIL_PORT")
	pass := os.Getenv("MAIL_PASS")

	c, err := client.DialTLS(mailhost+":"+mailport, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	defer c.Logout()

	// Login
	if err := c.Login(emailname, pass); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("*BOX " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last max_count messages
	var max_count uint32 = 100
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > max_count {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - max_count
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 100)
	done = make(chan error, 1)
	var section = &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	log.Println("Last max_count messages:")

	var maillist []Mail

	for msg := range messages {
		var mailInfo Mail
		r := msg.GetBody(section)
		if r == nil {
			log.Fatal("get body from server fail!")
		}
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		header := mr.Header
		var subject string
		if date, err := header.Date(); err == nil {
			mailInfo.Date = date
			//log.Println("Date:", date, reflect.TypeOf(date))
		}
		if from, err := header.AddressList("From"); err == nil {
			mailInfo.From = from
			//log.Println("From:", from, reflect.TypeOf(from))
		}
		if to, err := header.AddressList("To"); err == nil {
			mailInfo.To = to
			//log.Println("To:", to, reflect.TypeOf(to))
		}
		if subject, err = header.Subject(); err == nil {
			mailInfo.Subject = subject
			//log.Println("Subject:", subject, reflect.TypeOf(subject))
		}

		// 处理邮件正文
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal("NextPart:err ", err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// 正文消息文本
				b, _ := ioutil.ReadAll(p.Body)
				// mailFile := fmt.Sprintf("INBOX/%s.eml", subject)
				// f, _ := os.OpenFile(mailFile, os.O_RDWR|os.O_CREATE, 0766)
				// f.Write(b)
				// f.Close()
				//log.Printf("body: %v\n", string(b[:]))
				mailInfo.Body = string(b[:])

			case *mail.AttachmentHeader:
				// 正文内附件
				filename, _ := h.Filename()
				log.Printf("attachment: %v\n", filename)
			}
		}

		maillist = append(maillist, mailInfo)

	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

	context.JSON(http.StatusOK, gin.H{"data": maillist, "code": 0})
}
