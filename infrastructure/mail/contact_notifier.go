package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"yutagame-backend/domain/model"
)

type ContactNotifier interface {
	SendContactNotification(ctx context.Context, inquiry *model.ContactInquiry) error
}

type GameRecommendationNotifier interface {
	SendGameRecommendationNotification(ctx context.Context, recommendation *model.GameRecommendation) error
}

type SMTPContactNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       string
	adminURL string
	useTLS   bool
	skipSend bool
}

func NewSMTPContactNotifierFromEnv() *SMTPContactNotifier {
	port, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SMTP_PORT")))
	if err != nil || port <= 0 {
		port = 587
	}
	return &SMTPContactNotifier{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     port,
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		to:       strings.TrimSpace(os.Getenv("CONTACT_NOTIFY_TO")),
		adminURL: strings.TrimRight(strings.TrimSpace(os.Getenv("ADMIN_SITE_URL")), "/"),
		useTLS:   strings.EqualFold(strings.TrimSpace(os.Getenv("SMTP_USE_TLS")), "true"),
		skipSend: strings.EqualFold(strings.TrimSpace(os.Getenv("SMTP_SKIP_SEND")), "true"),
	}
}

func (n *SMTPContactNotifier) Enabled() bool {
	return !n.skipSend && n.host != "" && n.from != "" && n.to != ""
}

func (n *SMTPContactNotifier) SendContactNotification(ctx context.Context, inquiry *model.ContactInquiry) error {
	if !n.Enabled() || inquiry == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	subject := "【PACKAGE FROESST】お問い合わせが届きました"
	body := n.buildBody(inquiry)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", n.from),
		fmt.Sprintf("To: %s", n.to),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	var auth smtp.Auth
	if n.username != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}
	if n.useTLS {
		return sendMailTLS(addr, n.host, auth, n.from, []string{n.to}, []byte(message))
	}
	return smtp.SendMail(addr, auth, n.from, []string{n.to}, []byte(message))
}

func (n *SMTPContactNotifier) SendGameRecommendationNotification(ctx context.Context, recommendation *model.GameRecommendation) error {
	if !n.Enabled() || recommendation == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	subject := "【PACKAGE FROESST】おすすめゲーム投稿が届きました"
	body := n.buildGameRecommendationBody(recommendation)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", n.from),
		fmt.Sprintf("To: %s", n.to),
		fmt.Sprintf("Subject: %s", mime.QEncoding.Encode("utf-8", subject)),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	var auth smtp.Auth
	if n.username != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}
	if n.useTLS {
		return sendMailTLS(addr, n.host, auth, n.from, []string{n.to}, []byte(message))
	}
	return smtp.SendMail(addr, auth, n.from, []string{n.to}, []byte(message))
}

func (n *SMTPContactNotifier) buildBody(inquiry *model.ContactInquiry) string {
	adminLine := ""
	if n.adminURL != "" && inquiry.ID > 0 {
		adminLine = fmt.Sprintf("\n管理画面URL: %s/admin/contacts/%d\n", n.adminURL, inquiry.ID)
	}
	return fmt.Sprintf(`お問い合わせが届きました。

ID: %d
名前: %s
メールアドレス: %s
件名: %s

本文:
%s
%s`, inquiry.ID, inquiry.Name, inquiry.Email, inquiry.Subject, inquiry.Message, adminLine)
}

func (n *SMTPContactNotifier) buildGameRecommendationBody(recommendation *model.GameRecommendation) string {
	adminLine := ""
	if n.adminURL != "" && recommendation.ID > 0 {
		adminLine = fmt.Sprintf("\n管理画面URL: %s/admin/recommendations/%d\n", n.adminURL, recommendation.ID)
	}
	return fmt.Sprintf(`おすすめゲーム投稿が届きました。

ID: %d
ゲーム名: %s

おすすめ理由:
%s
%s`, recommendation.ID, recommendation.GameName, recommendation.Reason, adminLine)
}

func sendMailTLS(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Quit()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		return err
	}
	return writer.Close()
}
