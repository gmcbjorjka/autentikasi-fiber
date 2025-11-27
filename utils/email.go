package utils

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"autentikasi/config"

	"gopkg.in/mail.v2"
)

// GenerateOTP creates a random 6-digit OTP
func GenerateOTP() string {
	return strconv.Itoa(rand.Intn(900000) + 100000)
}

// SendOTPEmail sends OTP to user email
func SendOTPEmail(cfg *config.Config, recipientEmail, otp string) error {
	m := mail.NewMessage()
	m.SetHeader("From", cfg.MailDefaultSender)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "OTP untuk Reset Kata Sandi - Dompetku")

	body := fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; background-color: #f5f5f5; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 10px;">
        <h2 style="color: #333;">Reset Kata Sandi Dompetku</h2>
        <p>Halo,</p>
        <p>Kami menerima permintaan untuk reset kata sandi akun Anda. Gunakan kode OTP di bawah ini untuk melanjutkan proses reset:</p>
        
        <div style="background-color: #f0f0f0; padding: 20px; border-radius: 5px; text-align: center; margin: 20px 0;">
            <p style="font-size: 14px; color: #666; margin: 0 0 10px 0;">Kode OTP:</p>
            <p style="font-size: 32px; font-weight: bold; color: #6b4cc9; letter-spacing: 5px; margin: 0;">%s</p>
        </div>
        
        <p style="color: #666;">Kode OTP ini berlaku selama 15 menit. Jika Anda tidak melakukan permintaan ini, abaikan email ini.</p>
        
        <p style="color: #999; font-size: 12px; margin-top: 30px;">
            Salam,<br>
            Tim Dompetku
        </p>
    </div>
</body>
</html>
	`, otp)

	m.SetBody("text/html", body)

	port, _ := strconv.Atoi(cfg.MailPort)
	d := mail.NewDialer(cfg.MailServer, port, cfg.MailUsername, cfg.MailPassword)

	// Configure TLS with ServerName and timeout
	d.TLSConfig = &tls.Config{ServerName: cfg.MailServer}
	d.Timeout = 10 * time.Second

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}
	return nil
}

// SendPasswordResetSuccessEmail sends confirmation email after successful reset
func SendPasswordResetSuccessEmail(cfg *config.Config, recipientEmail string) error {
	m := mail.NewMessage()
	m.SetHeader("From", cfg.MailDefaultSender)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "Kata Sandi Berhasil Direset - Dompetku")

	body := `
<html>
<body style="font-family: Arial, sans-serif; background-color: #f5f5f5; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 10px;">
        <h2 style="color: #333;">Kata Sandi Berhasil Direset</h2>
        <p>Halo,</p>
        <p>Kata sandi akun Dompetku Anda telah berhasil direset. Anda sekarang dapat login dengan kata sandi baru Anda.</p>
        
        <p style="margin-top: 30px; color: #666;">Jika ini bukan Anda atau Anda tidak melakukan reset kata sandi, segera hubungi tim support kami.</p>
        
        <p style="color: #999; font-size: 12px; margin-top: 30px;">
            Salam,<br>
            Tim Dompetku
        </p>
    </div>
</body>
</html>
	`

	m.SetBody("text/html", body)

	port, _ := strconv.Atoi(cfg.MailPort)
	d := mail.NewDialer(cfg.MailServer, port, cfg.MailUsername, cfg.MailPassword)

	// Configure TLS with ServerName and timeout
	d.TLSConfig = &tls.Config{ServerName: cfg.MailServer}
	d.Timeout = 10 * time.Second

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}
	return nil
}
