# Sending real email with Resend

The app sends auth mail through [Resend](https://resend.com) using the
official SDK (`github.com/resend/resend-go/v4`). `auth.EmailSender`
(`internal/auth/email.go`) is the only place that touches Resend; the
service layer just calls `SendVerificationEmail` /
`SendPasswordResetEmail` and treats failures as best-effort (logged, auth
still succeeds).

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

## 4. How it is wired (already done)

`internal/auth/email.go` builds a `resend.NewClient` per send and posts a
`resend.SendEmailRequest{From, To, Subject, Html}`. `cmd/api/main.go`
fills `auth.EmailSender{APIKey, From, AppURL}` from config; `APP_URL`
should become the **frontend** origin once you have one, so links land in
the UI instead of the API.

Keep tokens out of server logs in production: the `auth email sent`
log line records the Resend id, recipient and subject only.

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
