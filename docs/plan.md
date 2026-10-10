# Pomosky: to'liq loyiha rejasi (Go SSR, Azure)

**Pomosky** = pomodoro + sky. Host (siz) ish jarayonida taymerni boshqarasiz, tomoshabinlar havola orqali qo'shilib **birga ishlaydi**. Butun xonaning fon osmoni host joylashuvidagi **haqiqiy ob-havo va quyosh holati**ga qarab jonli o'zgaradi ("Seulga ochiq deraza").

Maqsad: TikTok/Instagram videolarida monitor ekranini yozib olish. Ekranda havola/QR va "hozir N kishi ishlayapti" hisoblagichi o'sish aylanasini yaratadi: video → havola → tomoshabin kiradi → son o'sadi → yangi video.

> Belgilar: `[ ]` bajarilmagan, `[x]` bajarilgan. **(tekshirish)** = hozircha tasdiqlanmagan, sinab yoki rasmiy hujjatdan aniqlash kerak.

---

## 1. Qarorlar

| Mavzu | Qaror |
|---|---|
| Nom | **Pomosky** (domen/npm/GitHub bo'sh ekani 2026-10-10 da tekshirilgan; handle'lar va savdo belgisi tekshirilmagan) |
| Arxitektura | **Hammasi Go'da, SSR** (bitta ilova, bitta deploy) |
| Interaktiv qism | Oddiy **TypeScript modullar** (Vue/framework yo'q), Vite yoki esbuild bilan yig'iladi |
| Pattern | **MVC** (Model = sof yadro + store, Controller = handler, View = `html/template`) |
| Ma'lumotlar bazasi | **PostgreSQL** |
| Redis | MVP'da **yo'q** |
| Login | **Ixtiyoriy.** Anonim kirish asosiy (guard emas) |
| Hosting | **Azure VM** (sizning serveringiz, doimiy ishlaydi) |
| Reverse proxy / TLS | **Caddy** (avto HTTPS) |
| Ob-havo | Open-Meteo. Bepul API **faqat nokommersiya** (reklama/obuna bo'lsa litsenziya kerak), < 600 so'rov/daqiqa, CC BY 4.0 atributsiya. 2026-10-10 da hujjatdan tekshirildi |

Bepul hostingning (Render va hokazo) uxlash muammosi **endi yo'q**, shuning uchun u variant rejadan olib tashlandi.

### Maqsad emas (birinchi versiyada)
Video/audio aloqa, erkin chat, to'lov, mobil ilova, ko'p tilli interfeys, bir necha host, Vue/React.

---

## 2. Mahsulot tavsifi

### Rollar
- **Host** (siz): xona yaratadi, taymerni boshqaradi, joylashuvni beradi.
- **Tomoshabin:** havola bilan kiradi, taymerni ko'radi, reaksiya yuboradi. Hisob shart emas.
- **Ro'yxatdan o'tgan foydalanuvchi (ixtiyoriy):** statistika va streak saqlanadi.

### Sahifalar (SSR)
| Yo'l | Nima |
|---|---|
| `/` | Landing: nima bu, "hozir ishlayotgan xona" va keyingi sessiya vaqti |
| `/r/{id}` | Xona (tomoshabin): osmon, taymer, "N kishi ishlayapti", reaksiyalar |
| `/r/{id}/host` | Host panel (host token yoki egasi hisobi) |
| `/r/{id}?mode=record` | Yozib olish rejimi: to'liq ekran, minimal UI, havola/QR |
| `/login`, `/account` | Ixtiyoriy login va profil/statistika |
| `/privacy` | Maxfiylik bayoni |

Ijtimoiy tarmoq havola ko'rinishi uchun har bir xona sahifasida **Open Graph meta teglar** (SSR ustunligi).

### Foydalanuvchi hikoyalari (qabul mezoni bilan)
- Tomoshabin havolani ochadi va **hisob yaratmasdan** taymerni ko'radi. (Mezon: login so'ralmaydi.)
- Host taymerni boshlasa, hamma tomoshabinda **0.5 soniyadan kam farq** bilan boshlanadi.
- Host sahifani yangilasa yoki internet uzilsa, taymer **to'xtamaydi va tiklanadi**.
- Osmon rangi quyosh balandligiga qarab o'zgaradi: tong, kun, oltin soat, shom, tun.
- Tomoshabindan joylashuv ruxsati **so'ralmaydi**.
- Anonim foydalanuvchi hamma narsani ishlata oladi, ro'yxatdan o'tgan statistikasini ko'radi.

---

## 3. Arxitektura

### MVC + sof yadro
- **Model:** `room` (sof yadro: `reduce(state, event) → state`, soat/tarmoq/fayl yo'q) va `store` (Postgres).
- **Controller:** `http` va `ws` handler'lar. Ingichka: so'rovni o'qiydi, Modelni chaqiradi, Viewga beradi. Mantiq yozilmaydi.
- **View:** `html/template` (layout + sahifalar), `embed` bilan binarga qo'shiladi.
- Vaqt va tasodifiy narsalar (`now`, ID) yadroga tashqaridan beriladi (testlash uchun).

### Papka tuzilmasi
```
pomosky/
  cmd/server/main.go
  internal/
    room/            # sof yadro: State, Event, reduce, testlar
    models/          # store: Postgres (pgx + sqlc), foydalanuvchi, xona, statistika
    controllers/     # sahifa handler'lari, auth, API
    ws/              # WebSocket Hub (xona boshiga bitta)
    weather/         # Open-Meteo adapteri + kesh
    middleware/      # OptionalAuth, rate limit, CSRF, logging, recover
    views/           # *.html shablonlar (layout, sahifalar, qismlar)
  web/               # frontend manbalari (TypeScript)
    src/
      timer.ts       # taymer va server vaqti offseti
      ws.ts          # WebSocket mijozi, qayta ulanish
      sky/           # quyosh hisobi, bulut, yomg'ir, yulduzlar (Canvas/WebGL)
      host.ts        # host panel
      record.ts      # yozib olish rejimi
    dist/            # yig'ilgan fayllar (hash'li nomlar)
  migrations/        # goose
  deploy/            # docker-compose.yml, Caddyfile
  docs/
  .github/workflows/
```

### Interaktiv qism (SSR + kichik JS)
- Sahifa serverda tayyor HTML bo'lib keladi, taymer/WebSocket/osmon **TypeScript modullari** bilan ishlaydi.
- Vue/React yo'q. Host forma (xona nomi, vaqt) oddiy POST yoki WebSocket xabari.
- Statik fayllar hash'li nomda, uzoq kesh sarlavhalari bilan (Caddy beradi). Cloudflare proksi ixtiyoriy (CDN, DDoS himoyasi) **(tekshirish)**.

### Xona holati (yadro)
```
State {
  RoomID, HostTokenHash
  Timer { Phase: idle|focus|break|paused,
          StartedAt (server vaqti), DurationMs,
          PausedRemainingMs, Label }
  Presence { Count }
  Reactions { counters }
  Location { City, LatRounded, LonRounded, Tz }
}
Event: TimerStart, TimerPause, TimerResume, TimerReset,
       Join, Leave, Reaction, SetLocation, HostRecover
```

### Sinxron taymer
- Server `startedAt` va `durationMs` ni saqlaydi, qolgan vaqtni mijoz hisoblaydi: `remaining = startedAt + durationMs - (clientNow + offset)`.
- Ulanganda server `serverNow` yuboradi, mijoz `offset = serverNow - (clientReceiveTime - rtt/2)` hisoblaydi va vaqti-vaqti bilan tuzatadi.
- Pauzada `PausedRemainingMs` saqlanadi.
- Server doimiy ishlagani uchun xona holati xotirada turadi; qayta ishga tushganda faol xona taymerlari Postgres snapshot yoki host qayta ulanishi (`HostRecover`) bilan tiklanadi.

### WebSocket protokoli (qoralama)
- Mijoz → server: `join{room, nickname}`, `host.auth{token}`, `timer.start{durationMs, label}`, `timer.pause`, `timer.resume`, `timer.reset`, `reaction{kind}`, `ping`.
- Server → mijoz: `state{snapshot, serverNow}`, `timer.updated{...}`, `presence{count}`, `reaction.tick{kind, count}`, `weather{...}`, `error{code}`.
- JSON, kichik xabarlar. **Ping/pong har ~25-30 soniyada**: Azure tarmog'ida bo'sh TCP ulanishlar qisqa vaqtdan keyin uzilishi mumkin (standart ~4 daqiqa), shuning uchun yurak urishi shart **(tekshirish)**.

### HTTP yo'llar (qoralama)
- `GET /`, `GET /r/{id}`, `GET /r/{id}/host`, `POST /r/new` (xona yaratish → host tokeni bir marta ko'rsatiladi, serverda faqat xesh).
- `GET /ws/{id}` (WebSocket), `GET /api/weather?room={id}` (keshdan), `GET /healthz`.
- Auth: `GET /login`, `POST /login`, `GET /auth/callback`, `POST /logout` (usul tanlanganidan keyin).
- Formalar uchun **CSRF token**.

### Ob-havo va osmon
- Go ob-havoni xona uchun 5-10 daqiqa keshlaydi, hamma tomoshabinga bitta so'rov.
- Quyosh/oy holati mijozda hisoblanadi (kenglik, uzunlik, vaqt).
- Xaritalash:

| Kirish | Ko'rinish |
|---|---|
| Quyosh balandligi | osmon gradienti: tong, kun, oltin soat, shom, tun |
| Bulut ulushi (%) | bulut qatlami zichligi, yulduzlar xiraligi |
| Shamol tezligi va yo'nalishi | bulut harakati |
| Yog'in intensivligi | yomg'ir/qor tomchilari soni va tezligi |
| Ob-havo kodi | tuman, chaqmoq, qor |
| Oy fazasi (hisoblanadi) | tunda oy |

- Sahifada doim "yangilandi N daqiqa oldin" (ma'lumot ~15 daqiqa kechikishi mumkin).
- Open-Meteo parametr nomlari, limit va shartlari rasmiy hujjatdan tekshiriladi **(tekshirish)**.

### Ma'lumotlar bazasi (PostgreSQL)
- `users(id, email NULL, provider, nickname, created_at)`
- `sessions(id, user_id, expires_at)`
- `rooms(id, owner_user_id NULL, host_token_hash, name, city, lat_rounded, lon_rounded, tz, created_at)`
- `focus_sessions(id, room_id, started_at, ended_at, peak_viewers)`
- `user_stats(user_id, total_focus_ms, streak_days, last_active_date)` (ixtiyoriy)
- Joylashuv **yaxlitlangan** (~1 km). Kod: `pgx`, `sqlc`, `goose`.
- **Qayerda:** MVP uchun **Azure VM'da** (docker) qo'yish pul sarflamaydi. Shart: VM RAM'i yetishi (1 GB tor bo'lishi mumkin), va **zaxira nusxa** (kunlik `pg_dump`, VM'dan tashqariga, masalan Azure Blob) **(tekshirish)**. Muqobil: Neon bepul (alohida joyda, kamroq ish, lekin 5 daqiqada uxlaydi) yoki Azure'ning boshqariladigan Postgres'i (pullik).

### Redis
MVP'da kerak emas (bitta nusxa, xona holati xotirada). Keyin kerak bo'lsa VM'da lokal o'rnatiladi yoki Upstash. `ports` interfeyslari orqali qo'shish oson.

### Login (ixtiyoriy)
- `OptionalAuth` middleware: cookie bo'lsa foydalanuvchini aniqlaydi, **yo'q bo'lsa hech narsani bloklamaydi**.
- Faqat host amallari host tokeni yoki xona egasi hisobi bilan himoyalanadi.
- Usul: **Google OAuth** yoki **email magic link** (parolsiz). Qaror 4-bosqichda.
- Sessiya cookie: `httpOnly`, `Secure`, `SameSite=Lax`, Postgres'da. Sahifa va API bir domen bo'lgani uchun cross-domain cookie muammosi yo'q.

### Muhit o'zgaruvchilari
`PORT`, `DATABASE_URL`, `SESSION_SECRET`, `HOST_TOKEN_SALT`, `WEATHER_CACHE_TTL`, `OPEN_METEO_BASE`, `BASE_URL`, login usuliga qarab `OAUTH_GOOGLE_*` yoki email xizmati kaliti. Hech narsa kodga yoki repoga tushmaydi (`.env` fayl `.gitignore`da, VM'da ruxsatlari cheklangan).

---

## 4. Azure va deploy

### Kerakli ma'lumot (sizdan)
- [ ] VM turi va RAM/CPU (kamida 1 vCPU, 2 GB RAM tavsiya: Go + Postgres + Caddy uchun) **(tekshirish)**.
- [ ] Hudud (region): tomoshabinlar asosan qaerda? Seul/Koreya yoki Yevropa/AQSh. Latency uchun muhim, lekin WebSocket yengil.
- [ ] Ochiq IP statikmi? DNS shu IP'ga qaraydi (dinamik IP o'zgarib qolsa sayt yo'qoladi).
- [ ] Xarajat: VM, IP, trafik (egress) hisobi. Virusli video trafikni oshirishi mumkin, Azure hisobini kuzatish.

### Tavsiya etilgan sxema
```
Internet → (ixtiyoriy Cloudflare) → Azure VM
                                     └─ Caddy (80/443, avto HTTPS)
                                          └─ Go ilova (SSR + WebSocket) :8080
                                     └─ Postgres (docker, faqat localhost)
```
- **Docker Compose:** `app`, `caddy`, `postgres`. Mahalliy va serverda bir xil.
- **Deploy:** GitHub Actions: test → binar/image yig'ish → VM'ga yuborish (SSH) → qayta ishga tushirish. Yoki qo'lda `docker compose pull && up -d`.
- **Domen:** `pomosky.com` → A yozuvi VM IP'ga.

### Xavfsizlik (VM)
- [ ] NSG (tarmoq qoidalari): faqat 80/443 ochiq, SSH faqat sizning IP'ingizdan yoki kalit bilan.
- [ ] SSH kalit, parol bilan kirish o'chiq; avtomatik xavfsizlik yangilanishlari.
- [ ] Postgres tashqariga ochilmaydi (faqat docker tarmog'i/localhost).
- [ ] Sirlar `.env` da, ruxsati 600; zaxira kalitlari repoda yo'q.
- [ ] Caddy: HSTS, gzip, kesh sarlavhalari, rate limit (kerak bo'lsa Go ichida).
- [ ] Kuzatuv: `healthz` ni tashqi uptime xizmati tekshiradi (bepul), log aylanishi.

---

## 5. Xavfsizlik va maxfiylik (ilova)
- Host tokeni tasodifiy, serverda faqat xesh. Tomoshabin taymerni boshqara olmaydi.
- Tomoshabin: faqat taxallus (filtr bilan), email/hisob majburiy emas.
- Aniq koordinata saqlanmaydi (yaxlitlangan), tomoshabinga shahar nomi.
- Rate limit: IP bo'yicha ulanish va xabar soni; WebSocket xabar hajmi chegarasi; noto'g'ri xabarni rad etish.
- Forma CSRF himoyasi, shablonlarda avto-escape (`html/template`), Content-Security-Policy.
- Erkin **chat yo'q** (moderatsiya yuki), faqat tayyor reaksiyalar.
- Maxfiylik sahifasi, hisobni va ma'lumotni o'chirish imkoniyati.
- Musiqa: mualliflik huquqi, litsenziyali yoki o'zingiz yaratgan ovoz.

---

## 6. Test strategiyasi
- **Birlik testlari:** reducer jadvalli testlar (event ketma-ketligi → holat), taymer chegaralari, ikkita host bo'lmasligi.
- **Integratsiya:** haqiqiy WebSocket bilan 3 mijoz, taymer farqi < 0.5 soniya; shablonlar render testi.
- **Uzilish testi:** host ulanishi uzilib qayta ulansa taymer tiklanadi; server qayta ishga tushganda tiklash.
- **Auth testi:** anonim foydalanuvchi hamma narsani ishlata oladi (login bloklamaydi).
- **Yuklama testi:** yuzlab ulanish, VM xotira va CPU o'lchovi.
- **Kamerada sinash:** monitorni telefon bilan olib muar va titroqni ko'rish.
- **Unumdorlik:** telefonda animatsiya FPS va batareya, tab yashirilganda to'xtashi.
- **CI (GitHub Actions):** `go test ./...`, `go vet`, linter, frontend build.

---

## 7. Bosqichlar (kun hisobi bilan)

### 1-bosqich: Tekshiruv va tayyorgarlik (2 kun)
- [ ] Repo (`pomosky`), litsenziya, `.gitignore`, README.
- [ ] Azure VM: turi, RAM, hudud, statik IP, NSG, SSH, Docker o'rnatilishi.
- [ ] Open-Meteo hujjati: maydonlar, limit, shartlar (tekshirish).
- [ ] Azure'da WebSocket sinovi: bo'sh ulanish qancha turadi, ping/pong kerakmi (tekshirish).
- [ ] Raqobatchilarni ko'rish (CoFocusRoom, Coffee Focus, Studyzone), farqni yozish.
- [ ] Domen (`pomosky.com`, `.app`) va handle'larni band qilish (o'zingiz).
- **Tugash mezoni:** VM tayyor va WebSocket sinovdan o'tgan, shartlar tekshirilgan.

### 2-bosqich: Skelet va sof yadro (2 kun)
- [x] `go mod init`, papkalar, `Makefile`, CI.
- [x] `room`: `State`, `Event`, `reduce` va jadvalli testlar.
- [x] Minimal HTTP server, layout shablon, `healthz`.
- **Tugash mezoni:** `go test ./...` o'tadi, yadroda tashqi paket yo'q.

### 3-bosqich: Sinxron taymer va WebSocket (3 kun)
- [x] `ws` Hub: xona boshiga bitta, ulanganda to'liq holat.
- [x] Server vaqti va mijoz offset hisobi (`room-client.ts`, `clock-sync.ts`).
- [x] Ulanganlar soni, qayta ulanish, ping/pong, `HostRecover` (server tomoni; server qayta ishga tushsa xona yo'qoladi, host yangi xona yaratadi).
- [x] Host token bilan boshqarish, rate limit, orqaga bosim, IP bo'yicha ulanish chegarasi, bo'sh xonani o'chirish.
- **Tugash mezoni:** 3 brauzer oynasida taymer bir xil (farq < 0.5 soniya), host yangilasa tiklanadi.

### 4-bosqich: Postgres va ixtiyoriy login (4 kun)
- [x] Migratsiyalar (goose), repozitoriylar (sqlc), ulanishlar puli.
- [x] Ikkala login usuli: Google OAuth (haqiqiy kalitlar bilan sinalmagan) va email magic link; sessiya cookie, CSRF.
- [x] `OptionalAuth`, xona egaligi, fokus sessiyasi tarixi, oylik statistika, hisobni o'chirish.
- [ ] Zaxira nusxa (kunlik `pg_dump`) va tiklash sinovi.
- **Tugash mezoni:** anonim foydalanuvchi to'liq kira oladi, login qilgan statistikasi saqlanadi, zaxiradan tiklash ishlaydi.

### 5-bosqich: Sahifalar (SSR) va interaktiv qism (4 kun)
- [x] Shablonlar: layout, landing, xona, host panel, login/profil (statistika UI bilan), maxfiylik, xato sahifasi.
- [ ] TypeScript modullar va ularni yig'ish (Vite), hash'li nomlar (Vite stabil nomlar ishlatadi; hash keyin).
- [x] Taxallus, reaksiyalar, "N kishi ishlayapti", havola va QR (lokal generatsiya).
- [ ] Open Graph teglar, ulanish holatlari (uzildi, qayta ulanmoqda).
- **Tugash mezoni:** host va tomoshabin to'liq oqim bilan ishlaydi (osmonsiz ham).

### 6-bosqich: Jonli osmon (5-6 kun, eng qiziq qism)
- [x] `weather` adapteri (Open-Meteo, hujjatdan tekshirilgan), kesh, `/api/weather`, shahar qidirish (`/api/places`).
- [x] Quyosh/oy hisobi, osmon gradienti, yulduzlar.
- [x] Bulutlar, yomg'ir, qor, tuman, chaqmoq (Canvas 2D).
- [x] Unumdorlik: tab yashirinsa to'xtash, `prefers-reduced-motion`, zaif qurilma rejimi.
- [x] "Yangilandi N daqiqa oldin" va Open-Meteo atributsiyasi (CC BY 4.0).
- **Tugash mezoni:** turli ob-havo kodlarida to'g'ri ko'rinish, telefonda silliq (FPS o'lchanadi).

### 7-bosqich: Videoga mos rejim va o'sish (2-3 kun)
- [ ] Yozib olish rejimi, vertikal (9:16) tartib.
- [ ] Ekranda havola, QR, hisoblagich; keyingi sessiya vaqti (tomoshabin vaqt mintaqasida).
- [ ] Kamerada sinash va tuzatish; oddiy tugash signali.
- **Tugash mezoni:** haqiqiy video yozib, telefon kamerasida o'qish qulay va titroq yo'q.

### 8-bosqich: Deploy (2 kun)
- [ ] Docker Compose (`app`, `caddy`, `postgres`), Caddyfile (avto HTTPS).
- [ ] GitHub Actions orqali deploy yoki qo'lda skript.
- [ ] Domen ulash, HTTPS, NSG va xavfsizlik ro'yxati.
- [ ] Yuklama testi, monitoring (uptime), log aylanishi.
- **Tugash mezoni:** to'liq oqim `https://pomosky.com` da ishlaydi.

### 9-bosqich: Ishga tushirish va o'lchash (davomiy)
- [ ] Anonim o'lchov: bir vaqtdagi eng ko'p odam (peak), sessiya soni, qaytib kelganlar. Shaxsiy ma'lumotsiz.
- [ ] Birinchi 5-10 video bilan sinash, fikr yig'ish.
- [ ] Keyingi g'oyalar: haftalik jadval, Telegram eslatma, ko'p host, reaksiya turlari, leaderboard.

---

## 8. Muddat

| Bosqich | Kun |
|---|---|
| 1. Tekshiruv | 2 |
| 2. Yadro | 2 |
| 3. Taymer + WebSocket | 3 |
| 4. Postgres + login | 4 |
| 5. Sahifalar (SSR) | 4 |
| 6. Jonli osmon | 5-6 |
| 7. Video rejim | 2-3 |
| 8. Deploy | 2 |
| **Jami** | **~24-26 ish kuni (taxminan 5 hafta, bir kishi)** |

Eng noaniq qismlar: 6-bosqich (vizual) va Azure sozlamasi. Login keyinga qoldirilsa, MVP ~4 hafta bo'ladi.

---

## 9. Xavflar

| Xavf | Chora |
|---|---|
| Video virusli bo'lib, yuk va Azure trafik hisobi oshadi | Rate limit, yengil WebSocket xabarlari, statik fayllar keshi (yoki Cloudflare), yuklama testi, xarajat kuzatuvi |
| VM kichik (RAM yetmaydi) | O'lchash, kerak bo'lsa VM'ni kattalashtirish yoki Postgres'ni Neon'ga ko'chirish |
| Postgres'da ma'lumot yo'qolishi | Kunlik zaxira, tiklashni sinash |
| Bo'sh WebSocket ulanish uziladi | Ping/pong, avto qayta ulanish |
| Ob-havo "real-time" emas | Halol yozuv "yangilandi N daqiqa oldin", quyosh holati hisoblanadi |
| Open-Meteo shartlari/limiti noaniq | 1-bosqichda tekshirish, kesh, bitta so'rov butun xona uchun |
| Telefon kamerasida muar va titroq | Dizaynni kamerada sinash, silliq harakat |
| Moderatsiya | Chat yo'q, taxallus filtri |
| Musiqa mualliflik huquqi | Litsenziyali yoki o'z ovozi |
| "Pomo" nomi bahsli bo'lishi mumkin (savdo belgisi) | USPTO/EUIPO/KIPRIS tekshiruvi |
| Murakkab UI kelajakda SSR'da qiyinlashadi | Hozir kerak emas; keyin kerak bo'lsa Vue/htmx qo'shish mumkin |
| VM'ni xavfsizlik bo'yicha qo'riqlash sizning zimmangizda | NSG, SSH kalit, yangilanishlar, `fail2ban` (ixtiyoriy) |
| Bir kishi uchun hajm katta | Login va murakkab animatsiyani keyinga qoldirish (MVP anonim va oddiy osmon) |

---

## 10. Ochiq savollar
1. Azure VM turi/RAM, hudud, statik IP bormi? (4-bo'lim)
2. Postgres: VM'da (docker) yoki Neon? (Tavsiya: VM'da + zaxira.)
3. Login usuli: Google OAuth, email magic link yoki ikkalasi?
4. Bir vaqtda faqat sizning xonangiz (MVP) yoki boshqa hostlar ham? (Tavsiya: MVP'da faqat siz.)
5. Cloudflare proksi ishlatiladimi (CDN, DDoS himoyasi)?
6. Fon tovushi: faqat tugash signali yoki sokin musiqa?
7. Savdo belgisi tekshiruvini qachon qilamiz?

---

## 11. Hozir sizning qadamlaringiz (repo yaratish)
- [ ] GitHub'da `pomosky` repo yarating (public yoki private).
- [ ] Litsenziya (masalan MIT) va `.gitignore` (Go, Node, `.env`) tanlang.
- [ ] `pomosky.com` va `pomosky.app` domenlarini band qiling, Instagram/TikTok handle'larini tekshiring.
- [ ] Azure VM ma'lumotlarini (tur, RAM, hudud, IP, ochiq portlar) ayting.
- [ ] Repo tayyor bo'lgach ayting, papka tuzilmasi va `go mod init` bilan boshlaymiz.
