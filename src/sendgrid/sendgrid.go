package sendgrid

import (
	"context"
	"fmt"

	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/pralolik/templgrid/pkg"
	"github.com/pralolik/templgrid/src/logging"
	"github.com/pralolik/templgrid/src/queue"
	"github.com/pralolik/templgrid/src/templatemanager"
)

type SendGrid struct {
	log       logging.Logger
	storage   *templatemanager.EmailStorage
	client    *sendgrid.Client
	isSandBox bool
}

func NewSendGrid(apiKey string, isSandBox bool, log logging.Logger, storage *templatemanager.EmailStorage) *SendGrid {
	return &SendGrid{
		log:       log,
		storage:   storage,
		client:    sendgrid.NewSendClient(apiKey),
		isSandBox: isSandBox,
	}
}

func (sg *SendGrid) Run(ctx context.Context, q queue.Interface) error {
	queueChannel, err := q.GetChannel()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case email := <-queueChannel:
			err = sg.sendEmail(email)
			if err != nil {
				sg.log.Error("error send email %s: %v ", email.TemplateName, err)
				continue
			}
			sg.log.Info("email sent %s", email.TemplateName)
		}
	}
}

func (sg *SendGrid) sendEmail(email *pkg.TemplgridEmailEntity) error {
	var err error
	var subject, emailHTML string

	if subject, emailHTML, err = sg.buildEmail(email); err != nil {
		return err
	}
	sg.log.Debug(
		"email %s for locale %s built subject: %s content: %s",
		email.TemplateName,
		email.Locale,
		subject,
		emailHTML)

	sgMail := &email.SendGridParameters
	sgMail.Subject = subject
	sgMail.AddContent(mail.NewContent("text/html", emailHTML))
	sg.setSandBox(sgMail)
	sg.log.Debug("email object prepared %v", sgMail)

	var res *rest.Response
	if res, err = sg.client.Send(sgMail); err != nil {
		return err
	}

	sg.log.Debug("response from sendgrid :%v ", res)
	if err = sg.processResponse(res); err != nil {
		return err
	}

	return nil
}

func (sg *SendGrid) buildEmail(email *pkg.TemplgridEmailEntity) (string, string, error) {
	subject, emailHTML, err := sg.storage.BuildEmail(email.TemplateName, email.Locale, email.EmailParameters)
	if err != nil {
		return "", "", err
	}
	return subject, emailHTML, nil
}

func (sg *SendGrid) setSandBox(sgMail *mail.SGMailV3) {
	isSandBox := sg.isSandBox
	if sgMail.MailSettings == nil {
		sgMail.MailSettings = mail.NewMailSettings()
	}
	sgMail.MailSettings.SetSandboxMode(&mail.Setting{Enable: &isSandBox})
}

func (sg *SendGrid) processResponse(response *rest.Response) error {
	if response.StatusCode >= 400 {
		return fmt.Errorf("%d response, body: %s", response.StatusCode, response.Body)
	}

	return nil
}
