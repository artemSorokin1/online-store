package notifiction

import (
	"context"
	"fmt"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/mail"
	"log/slog"
	"notification_service/internal/config"
	"strings"
)

type EmailSender struct {
	cfg          *config.NotifyConfig
	emailService *mail.Mail
}

func New(cfg *config.NotifyConfig) *EmailSender {
	emailService := mail.New(cfg.AppName, cfg.AppName)

	emailService.AuthenticateSMTP("", cfg.EmailFrom, cfg.Password, "smtp.yandex.ru")

	return &EmailSender{
		cfg:          cfg,
		emailService: emailService,
	}
}

// msg - сообщение об успешной оплате вида: payment sucess email: "email"
func NotifyUser(s *EmailSender, msg string) {
	splitMsg := strings.Fields(msg)
	emailTo := splitMsg[len(msg)-1]

	noti := notify.New()
	noti.UseServices(s.emailService)
	s.emailService.AddReceivers(emailTo)

	err := noti.Send(context.Background(), fmt.Sprintf("Поздравляем! Ваша оплата прошла успешно. (%s)"), msg)
	if err != nil {
		slog.Warn("Error sending email", slog.String("email", emailTo), slog.String("error", err.Error()))
		return
	} else {
		slog.Info("Email sent successfully", slog.String("email", emailTo))
	}

}
