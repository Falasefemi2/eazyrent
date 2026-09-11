# Sending real email with Resend

The app currently logs auth mail to stdout via the dummy `auth.EmailSender`
(`internal/auth/email.go`). This guide swaps it for
[Resend](https://resend.com) using only the standard library (no SDK needed).

## 1. Free tier

Resend's free plan covers side projects: a few thousand emails per month
with a daily cap. Check the current numbers on
https://resend.com/pricing before relying on them.

## 2. Resend account setup

1. Sign up at https://resend.com and open **API Keys**.
2. Create a key (e.g. `easyrent-dev`) and copy it. It starts with `re_`.
3. For development you can send **to your own signup address only**, using
   Resend's test sender:
   - From: `EasyRent <onboarding@resend.dev>`
   - To: the email address of your Resend account.
4. For production (any recipient, your own sender name):
   - Add your domain under **Domains**, add the shown SPF/DKIM TXT records
     at your DNS provider, wait for verification.
   - Then send from e.g. `EasyRent <noreply@yourdomain.com>`.

## 3. Config

Add to `.env` / `.env.example`:

```env
RESEND_API_KEY=re_your_key_here
EMAIL_FROM=EasyRent <onboarding@resend.dev>
APP_URL=http://localhost:8080
```

Extend `internal/config/config.go` (same pattern as the existing keys):

```go
ResendAPIKey string
EmailFrom    string
AppURL       string
```

`RESEND_API_KEY` stays required only when the real sender is wired; keep
the dummy as the default in `main.go` until you are ready.

## 4. Replace the dummy sender

`internal/auth/email.go` becomes a thin HTTP client. Full replacement:

```go
package auth

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

// EmailSender delivers auth mail through Resend's HTTPS API.
type EmailSender struct {
    APIKey string
    From   string
    AppURL string
    Client *http.Client // nil = http.DefaultClient
}

func (s EmailSender) client() *http.Client {
    if s.Client != nil {
        return s.Client
    }
    return &http.Client{Timeout: 10 * time.Second}
}

func (s EmailSender) send(to, subject, html string) error {
    body, _ := json.Marshal(map[string]string{
        "from":    s.From,
        "to":      to,
        "subject": subject,
        "html":    html,
    })
    req, err := http.NewRequest(
        http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body),
    )
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+s.APIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client().Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode < 200 || resp.StatusCode > 299 {
        return fmt.Errorf("resend: unexpected status %s", resp.Status)
    }
    return nil
}

// SendVerificationEmail mails the email-verification link.
func (s EmailSender) SendVerificationEmail(to, fullName, token string) error {
    link := fmt.Sprintf("%s/auth/verify?token=%s", s.AppURL, token)
    return s.send(to, "Verify your EasyRent email", fmt.Sprintf(
        `<p>Hi %s,</p><p>Confirm your email:</p><p><a href="%s">%s</a></p><p>Link expires in 24 hours.</p>`,
        fullName, link, link,
    ))
}

// SendPasswordResetEmail mails the password-reset link.
func (s EmailSender) SendPasswordResetEmail(to, fullName, token string) error {
    link := fmt.Sprintf("%s/reset-password?token=%s", s.AppURL, token)
    return s.send(to, "Reset your EasyRent password", fmt.Sprintf(
        `<p>Hi %s,</p><p>Reset your password:</p><p><a href="%s">%s</a></p><p>Link expires in 1 hour. Ignore this mail if you did not ask.</p>`,
        fullName, link, link,
    ))
}
```

Wire it in `cmd/api/main.go` (replacing `auth.EmailSender{}`):

```go
auth.EmailSender{
    APIKey: cfg.ResendAPIKey,
    From:   cfg.EmailFrom,
    AppURL: cfg.AppURL,
},
```

Notes:

- The service layer calls these the same way as the dummy, so
  `service.go` does not change. Mail failures stay best-effort (logged,
  auth still succeeds), matching the current behavior.
- `APP_URL` should be the **frontend** origin once you have one, so links
  land in the UI instead of the API.
- Keep tokens out of server logs in production: the dummy's `token=%s`
  log lines are for development only.

## 5. Testing

1. `task dev`, sign up in Swagger with **your own** Resend account email.
2. The mail arrives from `onboarding@resend.dev`; open the link (or paste
   the token into `GET /auth/verify`).
3. Same for forgot-password: the reset link arrives within seconds.
4. Resend's dashboard (**Emails** tab) shows delivery status per message;
   check there first when a mail does not arrive.

## 6. Going to production

- Verify your domain (section 2, step 4) and switch `EMAIL_FROM` to it.
- Restrict the API key to the **Sending** permission.
- Never commit `.env`: the key lives in the environment / secret manager.
