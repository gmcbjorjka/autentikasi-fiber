# SMTP Configuration Guide untuk Password Reset Feature

## Troubleshooting Gmail SMTP Issues

### Error: "421 service not available (connection refused, too many connections)"

Ini biasanya berarti salah satu dari berikut:

#### 1. **Gmail App Password Tidak Valid**

Gmail sekarang require "App Password" bukan password akun biasa.

**Langkah setup:**

1. Buka: https://myaccount.google.com/apppasswords
2. Pilih "Mail" dan "Windows Computer" (atau device kamu)
3. Google akan generate password 16 character
4. Copy dan paste ke `.env` file sebagai `MAIL_PASSWORD`

**Contoh:**

```dotenv
MAIL_PASSWORD=tcgo vhvt ijwh rang
```

#### 2. **2-Factor Authentication Belum Enabled**

Gmail App Passwords hanya tersedia jika 2FA sudah diaktifkan.

**Langkah:**

1. Buka: https://myaccount.google.com/security
2. Scroll ke "How you sign in to Google"
3. Enable "2-Step Verification"
4. Setelah itu, baru bisa generate App Passwords

#### 3. **Less Secure Apps Access** (untuk akun lama)

Jika menggunakan Gmail lama (sebelum migration), mungkin perlu enable:

1. Buka: https://myaccount.google.com/lesssecureapps
2. Enable "Allow less secure apps"

#### 4. **Credentials Format Salah**

Pastikan `.env` file seperti ini:

```dotenv
MAIL_SERVER=smtp.gmail.com
MAIL_PORT=587
MAIL_USE_TLS=true
MAIL_USERNAME=smt6capstone@gmail.com
MAIL_PASSWORD=tcgo vhvt ijwh rang
MAIL_DEFAULT_SENDER=smt6capstone@gmail.com
```

**Perhatian:**

- `MAIL_USERNAME` = email address Gmail lengkap
- `MAIL_PASSWORD` = App Password (16 char dengan spasi)
- `MAIL_SERVER` = smtp.gmail.com (jangan ubah)
- `MAIL_PORT` = 587 (untuk TLS)
- `MAIL_DEFAULT_SENDER` = bisa sama dengan username atau berbeda

#### 5. **Testing SMTP Connection**

Untuk test apakah SMTP sudah benar, request endpoint ini:

```bash
curl -X POST http://localhost:3000/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"teguh1@example.com"}'
```

**Response Success:**

```json
{
  "code": "200",
  "message": "OTP sent to your email",
  "success": true,
  "data": {
    "email": "teguh1@example.com",
    "expires_in": 900
  }
}
```

**Response Error (jika SMTP config salah):**

```json
{
  "code": "500",
  "message": "Failed to send OTP email",
  "success": false
}
```

Check backend logs untuk error details.

---

## Testing Password Reset Flow

### Step 1: Request OTP

```bash
curl -X POST http://localhost:3000/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@gmail.com"}'
```

### Step 2: Check email untuk OTP code

OTP akan dikirim ke email dalam 15 menit.

### Step 3: Verify OTP

```bash
curl -X POST http://localhost:3000/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@gmail.com","otp":"123456"}'
```

### Step 4: Reset Password

```bash
curl -X POST http://localhost:3000/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@gmail.com","otp":"123456","password":"newpassword123"}'
```

### Step 5: Login dengan password baru

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your-email@gmail.com","password":"newpassword123"}'
```

---

## Common Issues & Solutions

| Issue                       | Cause                    | Solution                                      |
| --------------------------- | ------------------------ | --------------------------------------------- |
| "421 service not available" | Gmail App Password salah | Generate ulang di apppasswords                |
| "Invalid credentials"       | Username/password typo   | Check .env file, no trailing spaces           |
| "TLS required"              | Port 465 bukan 587       | Use port 587 dengan MAIL_USE_TLS=true         |
| Email tidak diterima        | Email blocked            | Check Gmail spam folder atau whitelist sender |
| "Too many connections"      | Rate limit               | Wait a few minutes sebelum retry              |

---

## Production Checklist

- [ ] Gunakan Email Service Provider yang reliable (SendGrid, AWS SES, Mailgun)
- [ ] Jangan hardcode credentials - gunakan environment variables
- [ ] Implement rate limiting untuk /forgot-password endpoint
- [ ] Log semua email transactions untuk audit trail
- [ ] Setup email templates yang lebih professional
- [ ] Add email verification sebelum menggunakan akun baru
- [ ] Implement password complexity rules
- [ ] Add account lockout setelah failed login attempts

---

## References

- Gmail App Passwords: https://support.google.com/accounts/answer/185833
- Gmail 2-Step Verification: https://support.google.com/accounts/answer/185839
- SMTP TLS Ports: https://support.google.com/mail/answer/13287
