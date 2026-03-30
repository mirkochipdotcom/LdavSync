## Piano: GoRubrica — CardDAV Directory Aziendale

Rubrica aziendale con sync LDAP automatica, gestione numeri di gruppo centralino, e protocollo CardDAV per integrazione Thunderbird/client esterni. Architettura identica a GoPulley (Go 1.22+, HTMX, SQLite, Docker single-container).

**TL;DR**: Sync LDAP periodica (ogni 1 ora) popola contatti con mail+numero interno. Admin configura prefisso numero primario (es. 0854321) via pannello web. Numeri di gruppo (es. 459 = Ufficio Protocollo) mostrano membri associati. UI ricerca stile Google Contacts pubblica. Server CardDAV RFC 6352 per Thunderbird.

---

**Steps**

### Fase 1 — Foundation (settimana 1-2)
1. Setup struttura progetto Go: `cmd/server/main.go`, `internal/{ldap,database,carddav,phonebook,config}`, `web/{templates,static}` *(seguire layout GoPulley)*
2. Implementare `internal/ldap/sync.go`: lettura LDAP per `mail`, `telephoneNumber` (interno), `displayName`, `uid` *(riusare pattern auth da GoPulley `internal/auth/ldap.go`)*
3. Schema SQLite in `internal/database/sqlite.go`:
   - Tabella `contacts`: `id, uid, display_name, email, ldap_ext, primary_number, department, last_sync`
   - Tabella `group_numbers`: `id, number, name, description`
   - Tabella `group_members`: `group_id, contact_id`
   - Indici su `uid`, `last_sync`, `group_id`
4. Goroutine background in `main.go` per sync LDAP ogni 1 ora: cron interno con `time.Ticker` *(parallelo al cleanup job di GoPulley)*
5. Config via `.env`: `LDAP_HOST`, `LDAP_BASE_DN`, `LDAP_USER_DN_TEMPLATE`, `LDAP_BIND_DN`, `LDAP_BIND_PASSWORD`, `SYNC_INTERVAL_HOURS=1`, `PRIMARY_NUMBER_PREFIX_TEMPLATE` (es. `0854321{ext}`)
6. Dockerfile multi-stage + `compose.yml` *(copiare da GoPulley, cambiare binary name in `gorubrica`)*

### Fase 2 — Pannello Admin (settimana 2-3) *(dipende da Fase 1)*
7. Login admin con LDAP group check: middleware `requireAdmin` basato su `ADMIN_USERS` (lista semicolon-separated) o `LDAP_ADMIN_GROUP` *(riusare pattern GoPulley `requireAuth` e `requireAdmin`)*
8. Template `web/templates/admin.html` con HTMX + Tailwind CDN *(riusare stile dashboard GoPulley)*
9. Pannello configurazione numero primario:
   - Form per modificare template `PRIMARY_NUMBER_PREFIX_TEMPLATE` (stored in DB `config` table)
   - Preview real-time: mostra esempio con interno LDAP
10. CRUD numeri di gruppo:
    - Lista numeri: table con `number`, `name`, `description`, azioni (edit, delete)
    - Form creazione: `POST /admin/groups` con validazione numero univoco
    - Form associazione membri: ricerca live contatti con HTMX autocomplete, `POST /admin/groups/{id}/members`
11. Override manuale contatto: form per modificare `primary_number` o `email` singolo contatto, flag `manual_override` in DB per non sovrascrivere con LDAP sync

### Fase 3 — Interfaccia Rubrica Pubblica (settimana 3-4) *(dipende da Fase 1)*
12. Template `web/templates/phonebook.html`: layout Material Design con card per contatto *(ispirazione da dashboard GoPulley ma con Material badges/chips)*
13. Ricerca live HTMX: `GET /search?q={query}` ritorna partial HTML con risultati filtrati per nome, cognome, numero, ufficio
14. Endpoint `GET /contacts` (pubblico): ritorna lista paginata contatti ordinata per `display_name`
15. Vista dettaglio contatto: `GET /contacts/{uid}` mostra card espansa con:
    - Nome completo
    - Email cliccabile
    - Numero diretto (primary_number)
    - Badge per ogni gruppo di appartenenza (es. "Gruppo 459 - Ufficio Protocollo")
16. Responsive mobile: media queries per layout single-column su <768px
17. Indicatore visivo numeri di gruppo: badge colorato (es. `badge-group`) vs numero diretto (badge normale)

### Fase 4 — Server CardDAV (settimana 4-5) *(dipende da Fase 1)*
18. Implementare `internal/carddav/server.go`:
    - Handler `PROPFIND` su `/carddav/` → ritorna lista contatti come vCard href
    - Handler `REPORT` (addressbook-query) → filtra contatti per criterio
    - Handler `GET /carddav/{uid}.vcf` → ritorna vCard 3.0 singola
19. Generazione vCard per ogni contatto:
    - `FN`, `EMAIL`, `TEL` (primary_number)
    - Numeri di gruppo come `TEL;TYPE=work,x-group:{numero}` con `X-ABLABEL:Gruppo {name}`
20. Autenticazione CardDAV con Basic Auth: check LDAP username/password *(riusare `internal/auth/ldap.go` Authenticate)*
21. Discovery `/.well-known/carddav` → redirect a `/carddav/`
22. Test con Thunderbird: configurazione manuale account CardDAV con URL `https://{host}/carddav/`, verificare sync contatti
23. Export vCard singolo contatto: button "Export" in vista dettaglio → download `.vcf`

### Fase 5 — Rifinitura (settimana 5-6) *(parallelo a Fase 2-4)*
24. i18n in `internal/i18n/messages.go`: supporto italiano e inglese *(copiare struttura GoPulley con `messages` map, `ResolveLocale`, `T`)*
25. Logging strutturato: usare `log.Printf` con prefissi `[LDAP]`, `[SYNC]`, `[CARDDAV]`, `[HTTP]` *(come GoPulley)*
26. Health check endpoint: `GET /health` → JSON `{"status":"ok","ldap":"connected","db":"ok","last_sync":"2026-03-10T..."}`
27. README.md e README.it.md: architettura, quick start, config, esempi .env *(template da GoPulley)*
28. `.env.example` con tutte le variabili commentate
29. CI/CD GitHub Actions: workflow build + push GHCR `ghcr.io/mirkochipdotcom/gorubrica:latest` *(copiare da GoPulley workflow)*
30. Versioning: file `VERSION`, injection via ldflags `-X main.AppVersion=...`

---

**File rilevanti**

**Da creare** (struttura nuova):
- `cmd/server/main.go` — HTTP server, routing, sync goroutine, logging
- `internal/ldap/sync.go` — `SyncContacts()` legge LDAP e popola DB
- `internal/database/sqlite.go` — schema, CRUD `contacts`, `group_numbers`, `group_members`, query ricerca
- `internal/carddav/server.go` — handlers PROPFIND, REPORT, GET .vcf
- `internal/phonebook/service.go` — business logic: genera `primary_number` da template, associa gruppi
- `internal/config/config.go` — `Config` struct + `Load()` da .env *(pattern GoPulley `getEnv`, `getEnvInt`, `getEnvBool`)*
- `internal/i18n/messages.go` — bundle it/en per UI
- `web/templates/phonebook.html` — ricerca pubblica + card contatti
- `web/templates/admin.html` — pannello admin CRUD gruppi + config
- `web/templates/login.html` — form login admin
- `web/static/css/style.css` — Tailwind CDN + custom Material Design badges
- `Dockerfile` — multi-stage Go builder + Alpine runtime
- `compose.yml` — single service, volume `/data`
- `.env.example` — template configurazione
- `README.md` + `README.it.md` — doc completa

**Da riusare come riferimento** (pattern GoPulley):
- `GoPulley/internal/auth/ldap.go` → `Authenticate()` per auth CardDAV + admin
- `GoPulley/internal/auth/ldap.go` → pattern `ldap.DialURL`, `StartTLS`, `Bind`, `Search`
- `GoPulley/internal/config/config.go` → `Load()`, `getEnv*` helpers
- `GoPulley/internal/i18n/messages.go` → struttura `messages map[string]map[string]string`, `T()`, `ResolveLocale()`
- `GoPulley/cmd/server/main.go` → pattern goroutine cleanup job con ticker *(da adattare per sync)*
- `GoPulley/cmd/server/main.go` → middleware `requireAuth`, `requireAdmin`, session management
- `GoPulley/Dockerfile` → multi-stage build con CGO per SQLite
- `GoPulley/compose.yml` → volume `/data`, ports, env_file
- `GoPulley/internal/database/sqlite.go` → pattern `InitDB`, `migrate`, query prepared statements

---

**Verification**

1. **Sync LDAP**: `docker logs gorubrica` mostra `[SYNC] synced 150 contacts from LDAP` ogni 1h
2. **Admin Panel**: login con credenziali LDAP admin, creare gruppo "459 - Ufficio Protocollo", associare 3 contatti, verificare salvataggio DB
3. **Ricerca pubblica**: aprire `http://localhost:8080`, cercare "mario", vedere risultati filtrati, cliccare dettaglio, verificare badge gruppi
4. **CardDAV Thunderbird**:
   - Thunderbird → Nuovo indirizzo → CardDAV
   - URL: `http://localhost:8080/carddav/`
   - Username/password LDAP
   - Verificare lista contatti sincronizzata
   - Cercare contatto con numero di gruppo, verificare campo `TEL` multiplo
5. **Health check**: `curl http://localhost:8080/health` → JSON con status OK + timestamp last_sync
6. **i18n**: cambiare browser locale it-IT vs en-US, verificare UI tradotta (navbar, button, placeholder)
7. **Container build**: `podman build -t gorubrica .` → success, size <50MB
8. **Compose startup**: `podman compose up -d` → container running, accessible su :8080

---

**Decisioni**

- **Sync frequency**: ogni 1 ora (bilanciamento tra freschezza dati e carico LDAP)
- **Prefisso numero primario**: configurabile come template `0854321{ext}` via `.env` + override in pannello admin
- **Admin auth**: lista esplicita `ADMIN_USERS` semicolon-separated (flessibile, no dipendenza gruppo LDAP)
- **Interfaccia pubblica**: nessuna autenticazione per rubrica web (come da richiesta), solo CardDAV richiede Basic Auth
- **Numeri di gruppo**: solo informativo/rubrica, GoRubrica non comunica con centralino (documentazione chiara in README)
- **CardDAV version**: vCard 3.0 per compatibilità Thunderbird legacy, supportare 4.0 in future se richiesto
- **Database overrides**: campo `manual_override` boolean in `contacts` → sync LDAP skip se true

---

**Further Considerations**

1. **Sync error handling**: come gestire contatti rimossi da LDAP?
   - **Opzione A** (raccomandato): soft-delete con flag `deleted_at`, mantenere storico
   - **Opzione B**: hard-delete immediato, pulizia completa
   
2. **Gruppi vs contatti esterni**: permettere numeri di gruppo senza membri LDAP?
   - **Opzione A** (raccomandato): sì, per numeri generici (es. reception, centralino principale)
   - **Opzione B**: no, solo gruppi con membri validi
   
3. **CardDAV sync direction**: supportare write-back da client a DB?
   - **Opzione A**: read-only (più semplice, dato che LDAP è source of truth)
   - **Opzione B** (raccomandato): read-write per override locali (sync conflict detection necessario)
