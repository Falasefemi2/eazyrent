# Avatar uploads with Cloudinary

## How it works

```
Frontend  ──file──▶  Cloudinary  ──secure_url──▶  Frontend
Frontend  ──PUT /me/avatar {avatar_url} + bearer──▶  API  ──stores URL──▶  Postgres
```

The image bytes never touch your backend. The API only validates the URL
and stores it in `users.avatar_url`. No background jobs, no disk, no
multipart handling server-side.

## Do you need Cloudinary keys on the backend?

**No.** This flow uses an **unsigned upload preset**, which is designed for
browser/mobile uploads without a secret:

| Value              | Where it lives | Secret? |
| ------------------ | -------------- | ------- |
| Cloud name         | Frontend env   | No (public, it's in every image URL) |
| Unsigned preset    | Frontend env   | No (scoped by the preset's own settings) |
| API key / secret   | Nowhere        | Not used in this flow |

The unsigned preset itself carries the constraints (target folder, max
file size, allowed formats, transformations), so clients can't abuse it
beyond what you configured. You'd only need the API secret on the backend
if you later switch to **signed** uploads (server-generated signatures for
per-request constraints) — not needed now.

## 1. Cloudinary setup (dashboard, ~5 minutes)

1. Sign up at https://cloudinary.com (free tier) and note your **cloud
   name** on the dashboard.
2. Go to **Settings → Upload → Upload presets → Add upload preset**.
3. Set **Signing Mode** to **Unsigned**, name it e.g. `easyrent_avatar`.
4. Recommended preset settings:
   - **Folder**: `easyrent/avatars` (keeps uploads organized).
   - **Allowed formats**: `jpg,png,webp`.
   - **Max file size**: ~5 MB (`5242880` bytes).
   - **Transformation** (optional): incoming `c_fill,w_512,h_512` so every
     avatar is stored square at 512px regardless of what the user picks.
5. Save. The two values your frontend needs:
   - `CLOUDINARY_CLOUD_NAME=<cloud name>`
   - `CLOUDINARY_UPLOAD_PRESET=easyrent_avatar`

Put them in the **frontend** `.env`, not the API `.env`. Nothing is added
to the Go backend.

## 2. Frontend upload (example)

```js
// 1. Upload the file straight to Cloudinary.
const form = new FormData();
form.append("file", pickedFile);
form.append("upload_preset", process.env.CLOUDINARY_UPLOAD_PRESET);

const up = await fetch(
  `https://api.cloudinary.com/v1_1/${process.env.CLOUDINARY_CLOUD_NAME}/image/upload`,
  { method: "POST", body: form },
);
const { secure_url } = await up.json();

// 2. Tell the API to store it.
const res = await fetch("http://localhost:8080/me/avatar", {
  method: "PUT",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Bearer ${accessToken}`,
  },
  body: JSON.stringify({ avatar_url: secure_url }),
});
const user = await res.json(); // user.avatar_url is set
```

To remove an avatar: `PUT /me/avatar` with `{"avatar_url": ""}`.

## 3. Backend reference

- `PUT /me/avatar` (bearer required) — `internal/web/auth_handlers.go`.
- Accepts any valid `http(s)` URL on purpose: the endpoint is
  provider-agnostic, so switching from Cloudinary to R2/S3 later changes
  zero backend code.
- Rejects non-URLs and non-http(s) schemes with 400; missing/invalid token
  with 401. `Store.UpdateAvatar` sets `avatar_url` (or `NULL` when
  cleared) and bumps `updated_at`.
- The URL is returned as `avatar_url` (string or `null`) on `GET /me`.

## 4. Later hardening (optional, not now)

- Restrict the host allowlist to `res.cloudinary.com` in `updateAvatar`
  if you want to guarantee avatars only come from your Cloudinary account.
- Switch to signed uploads (backend holds `CLOUDINARY_API_SECRET` and
  mints per-request signatures) if you need per-user folders or quotas.
- Add an **Auto Moderation** add-on to the preset if user content needs
  screening.
