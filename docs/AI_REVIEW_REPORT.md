# AI inceleme raporu — 23 Eylül 2026

## Stage

Aşamalar arası kalite denetimi (Stage 1–8). Yeni bir aşama değildir; mevcut
uygulamanın kanıta dayalı hata taraması, doğrulanmış düzeltmesi ve test
kanıtının teslimidir. Domain modeli, PDC politikası, endorsement politikası
ve anomali kuralı **değiştirilmedi**.

## Kapsam, incelenen sürüm ve ortam

- **İncelenen commit:** `98b356ea859d984fdbb258bcdbc6257a0c1da3eb` ("AgroChain pilot: README.md"),
  2026-09-23 22:08:40 +0300. İnceleme öncesi `git status` çalışma ağacının temiz
  olduğunu doğruladı (`nothing to commit, working tree clean`) — yani incelenen
  kod, kullanıcının `/home/mert/Projects/AgroChain` dizinindeki mevcut haliyle
  birebir aynıdır, üzerine yazılmadı.
- **İnceleme yöntemi:** Depo, `github.com/mertosteron/AgroChain` üzerinden bulut
  kapsayıcısına klonlanarak (HEAD ve çalışma-ağacı temizliği doğrulandıktan
  sonra) orada okundu/derlendi/test edildi; kullanıcının asıl makinesinde hiçbir
  yıkıcı komut çalıştırılmadı. Yalnız `AGENTS.md` (git-ignored) ve `dist/`
  paket çıktısı doğrudan cihazdan okundu.
- **Sandbox kısıtları (raporun "BLOCKED" satırlarının nedeni):** Bu bulut
  kapsayıcısında Docker daemon çalışmıyor (`docker info` başarısız; CLI/Compose
  mevcut) ve kurumsal ağ politikası `proxy.golang.org` ile `repo.maven.apache.org`
  adreslerine erişimi 403 ile reddediyor. Go modülleri için GitHub'daki birebir
  aynı sürümlü aynalara `replace` yönlendirmesiyle (yalnız bu geçici klonda,
  cihaza hiç yazılmadı, iş bitince geri alındı) çalışan bir yol bulundu; Maven
  Central için aynı yol yoktur, bu nedenle `mvn test` bu ortamda **BLOCKED**'dır.
  Kullanıcının kendi makinesinde `backend/target/agrochain-backend-0.3.0.jar`
  dosyasının hazır olması, o makinede Maven erişiminin çalıştığını gösteriyor.
- **Okunan temel belgeler:** `AGENTS.md` (tam, 31 bölüm), `ARCHITECTURE.md`,
  `docs/ARCHITECTURE.md`, `docs/DATA_CONTRACTS.md`, `docs/STATE_MACHINE.md`,
  `docs/SECURITY_AND_PRIVACY.md`, `docs/PROJECT_STATUS_TR.md`,
  `docs/PROJECT_REVIEW_ROADMAP_TR.md` (önceki incelemenin tarihsel kanıtı —
  aşağıda ayrıca değerlendirildi), `docs/competition/*`. Bunlardan
  `PROJECT_STATUS_TR.md` ve önceki `PROJECT_REVIEW_ROADMAP_TR.md`'deki "PASS"
  ifadeleri **tarihsel kanıt** olarak alındı, güncel doğruluk kanıtı olarak
  değil; bu raporda her iddia bu incelemede yeniden koddan ve testten
  doğrulanmıştır (aşağıdaki "Çalıştırılan Testler" ve dosya bazlı analiz).

## Özet (önce / sonra)

**Önce:** İki geçmiş aşama kapanış raporu (Stage 7, Stage 8) "PASS" ve
"tamamlandı" diyordu; bu bir önceki AI incelemesinin (`PROJECT_REVIEW_ROADMAP_TR.md`,
20 Eylül) bulduğu 4 P0/P1 bulgusunun hepsi sonradan (21–23 Eylül commit'lerinde)
kapatılmıştı. Bu inceleme, o kapanışları **yeniden kodda doğruladı** (aşağıya
bakınız — hepsi hâlâ doğru) ve ayrıca kod tabanının geri kalanını (chaincode,
backend, frontend, ağ betikleri, demo/paket betikleri, CI, gönderim arşivi)
uçtan uca, kanıt zorunluluğuyla (repro/test) yeniden taradı.

**Sonra:** Taramada **P0 veya P1 seviyesinde hiçbir doğrulanmış açık bulunmadı**.
Tek somut, kanıtlanmış ve düzeltilen sorun P3 seviyesinde bir doğrulama
katmanı tutarsızlığıdır (bkz. AI-1). Bunun dışında, önceki incelemenin "dead
code" şüphesiyle not ettiğim `closedEvidence` tipi yeniden incelendi ve
**hata olmadığı** kanıtlandı (kasıtlı, test edilen bir "fail-closed" arayüz
sözleşmesi — bkz. bulgular tablosundaki "belirlenmiş-hata-değil" satırı).
Test kanıtı: bu ortamda çalıştırılabilen **70 otomatik testin tamamı PASS**
(33 Go chaincode + 8 demo-script + 27 network-script + 2 stage7-istatistik);
Docker/Maven gerektiren testler açıkça **BLOCKED** olarak işaretlendi, PASS
veya FAIL olarak yorumlanmadı.

## Bulgular

| ID | İlgili aşama | Önem | Dosya:satır | Durum |
| --- | --- | --- | --- | --- |
| AI-1 | Stage 5 (Backend API) | P3 | `backend/src/main/java/org/agrochain/Json.java:63` | **Düzeltildi** |
| AI-2 | Stage 4 (Privacy/Evidence) | — | `chaincode/agrochain/internal/agrochain/models.go:149-153` | **Belirlenmiş: hata değil** |
| AI-3 (tarihsel, yeniden doğrulandı) | Stage 4 | P0 (tarihsel) | `contract.go` / `models.go` | Kapatıldı — bu incelemede doğrulandı |
| AI-4 (tarihsel, yeniden doğrulandı) | Stage 3 | P1 (tarihsel) | `validation.go` `authorize()` | Kapatıldı — bu incelemede doğrulandı |
| AI-5 (tarihsel, yeniden doğrulandı) | Stage 3/CI | P2 (tarihsel) | `.github/workflows/stage3-chaincode.yml` | Kapatıldı — bu incelemede doğrulandı |
| AI-6 (tarihsel, yeniden doğrulandı) | Stage 4 (kimlikler) | P1 (tarihsel) | `network/scripts/privacy-prepare.py` | Kapatıldı — bu incelemede doğrulandı |

### AI-1 — Detay (Düzeltildi)

- **Tetikleyici:** `Json.id()` regex'i `[A-Z0-9]{8,40}` kabul ediyordu; oysa
  aynı sözleşmenin diğer **beş bağımsız kaynağı** — Go chaincode'un kendi
  `identifier` regex'i (`validation.go:74`, `{8,32}`), `docs/DATA_CONTRACTS.md`
  satır 67 (`{8,32}`), `docs/STAGE1_REPORT.md` satır 78'deki orijinal kabul
  testi (`{8,32}`), frontend `consumer.js` satır 4'teki lot-ID kontrolü
  (`{8,32}`) ve backend'in kendi ID üreticisi `Crypto.id()` (tam 32 karakter
  üretir) — hepsi `{8,32}`'de hemfikir. Yalnız `Json.id()` 8 karakter daha
  geniş bir aralık kabul ediyordu.
- **Beklenen/gerçek davranış:** 33–40 karakter uzunluğunda bir gövdeye sahip
  `operationId`/`batchId` vb., backend'in kendi istek doğrulamasından
  (`Workflow.validate`) geçip Fabric Gateway'e endorsement için gönderiliyor,
  ancak chaincode kendi `{8,32}` sınırında bunu `INVALID_IDENTIFIER` ile
  reddediyordu. Yani "geçersiz ID için açık hata" gereksinimi karşılanıyordu
  ama katman sırası yanlıştı: temiz bir 400 yerine, gecikmeli/uzak bir
  reddediş.
- **Kanıt:** Beş kaynağın karşılaştırmalı `grep` çıktısı (rapor ekindeki analiz
  ile aynı); Python `re` ile birebir eşdeğer desen davranışı doğrulandı
  (33 karakter eski regex'te kabul/yeni regex'te ret, 32/8 karakter her
  ikisinde de kabul) — bkz. "Çalıştırılan Testler".
- **Etki:** Yetkisiz erişim, veri sızıntısı veya yanlış iş sonucu **yok**;
  chaincode nihai otorite olduğu için hiçbir 33-40 karakterlik ID zincire asla
  yazılamazdı. Etki, hatalı girdi için beklenenden daha geç/dolaylı bir ret
  ve backend ile chaincode arasında belgelenen sözleşmeye aykırılıktır — P3
  "düşük etkili tutarlılık sorunu" tanımına tam uyuyor.
- **Düzeltme:** `Json.id()` regex'i `{8,40}` → `{8,32}` (tek karakterlik,
  en küçük yeterli değişiklik). Mevcut hiçbir test 32 karakterden uzun sabit
  bir ID kullanmıyor (`Crypto.id()` zaten tam 32 karakter üretiyor), bu yüzden
  değişiklik geriye dönük hiçbir testi bozmaz.
- **Doğrulama durumu:** Regex mantığı Python'da birebir eşdeğer desenle
  test edildi (PASS — bkz. aşağıda). Yeni bir JUnit regresyon testi eklendi
  (`BackendTest.idGrammarMatchesTheDocumentedAndChaincodeEnforcedThirtyTwoCharacterBound`)
  ancak Maven bu ortamda erişilemez olduğundan **derlenip gerçek JUnit ile
  çalıştırılamadı (BLOCKED)** — kullanıcının makinesinde `mvn test` ile
  doğrulanmalıdır (tek komut, aşağıda verilmiştir).

### AI-2 — Detay (Belirlenmiş: hata değil)

Bir önceki taramamda `models.go:149`'daki `closedEvidence{}` tipini ve onu
`dispatch()`'e bağlayan eski kodun kaldırıldığını görünce bunu geçici olarak
"ölü kod, P3 temizlik adayı" diye not almıştım. Daha derin inceleme bunun
**yanlış bir ilk izlenim** olduğunu gösterdi: `closedEvidence`,
`domain_test.go:354-356`'daki `TestLateFreightRecoveryAndPopulatedQueries`
testi tarafından bilinçli olarak kullanılıyor — `evidenceProvider` arayüzünün
"kanıt doğrulaması kullanılamıyorsa yaz işlemini reddet" sözleşmesini
(`EVIDENCE_VERIFICATION_UNAVAILABLE`) kanıtlayan, kasıtlı bir "fail-closed"
savunma testi. Üretim `dispatch()` yolu her zaman `realEvidence{}`'ı
bağlıyor (`contract.go`), yani `closedEvidence` üretimde asla tetiklenmiyor
— ama arayüzün güvenli-başarısız sözleşmesi otomatik testle korunuyor olması
iyi bir tasarımdır, hata değildir. Bu bulgu **kod değiştirilmeden**
kapatılmıştır; kullanıcının isteğinde açıkça belirtilen "kanıtlanmamış
şüpheyi doğrulanmış açık gibi sunma" ilkesi gereği, ilk izlenim düzeltmeden
rapora yazılmamıştır.

### AI-3 – AI-6 — Tarihsel bulguların yeniden doğrulanması

`docs/PROJECT_REVIEW_ROADMAP_TR.md` (20 Eylül tarihli, üstündeki güncelleme
notlarıyla açıkça geçersiz kılınmış önceki bir AI incelemesi) dört P0/P1/P2
bulgusu listeliyordu. Bu incelemede her biri **yeniden, doğrudan güncel
koddan** doğrulandı (yalnız eski rapora güvenilmedi):

- **AI-3 (eski P0):** `contract.go`'da `dispatch()`'in her zaman
  `engine{ps, realEvidence{...}}` oluşturduğu, `closedEvidence`'ın hiçbir
  üretim yoluna bağlı olmadığı satır satır doğrulandı → **kapalı**.
- **AI-4 (eski P1):** `validation.go` `authorize()` fonksiyonu, `Health`
  dışındaki her komut için boş rolü (`role == ""`) açıkça reddediyor
  (`fail("UNAUTHORIZED_ROLE")`); MSP-only geri düşüş yolu yok → **kapalı**.
- **AI-5 (eski P2):** `.github/workflows/stage3-chaincode.yml` içeriği okundu;
  `backend/**, chaincode/**, network/**` yollarında tetikleniyor ve
  `make stage7-check` (chaincode + backend + Playwright + iki temiz ağ
  bootstrap'ı) çalıştırıyor → **kapalı**.
- **AI-6 (eski P1):** `network/scripts/privacy-prepare.py` okundu; her rol
  için Fabric-CA standardı öznitelik formatında (`agrochain.role`, OID
  `1.2.3.4.5.6.7.8.1`) imzalı sertifika üretiyor, `cryptogen`'in tek başına
  yapamadığı iş rolü atamasını doğru şekilde tamamlıyor → **kapalı**.

## Değiştirilen dosyalar

| Dosya | Neden |
| --- | --- |
| `backend/src/main/java/org/agrochain/Json.java` | AI-1: ID regex sınırı `{8,40}` → `{8,32}`, chaincode/frontend/DATA_CONTRACTS.md ile hizalandı. Tek satır değişti. |
| `backend/src/test/java/org/agrochain/BackendTest.java` | AI-1 için regresyon testi eklendi (7 satır); mevcut test stiliyle tutarlı, başka hiçbir test değiştirilmedi/silinmedi. |

Başka hiçbir dosyada değişiklik yapılmadı. Chaincode (Go), ağ yapılandırması,
frontend, demo/paket betikleri ve tüm dokümantasyon **olduğu gibi** bırakıldı
— bu alanlarda kanıtlanmış bir kusur bulunamadı.

## Çalıştırılan testler

Aşağıdaki komutlar bu incelemede, klonlanan çalışma kopyası üzerinde
gerçekten çalıştırıldı (kullanıcının orijinal makinesine hiçbir şey
yazılmadı; sonuçlar aşağıda özetlenmiştir, ayrıntılı çıktı bu oturumda
mevcuttur):

| Komut | Sonuç | Not |
| --- | --- | --- |
| `cd chaincode/agrochain && go vet ./...` | **PASS** | Temiz |
| `gofmt -l cmd internal` | **PASS** | Çıktı yok (biçim sorunu yok) |
| `go test ./... -race -cover -count=1` | **PASS** | `agrochain/chaincode/internal/agrochain` — **33/33 alt-test PASS**, %81.2 satır kapsamı, veri yarışı yok |
| `python3 -m unittest discover -s scripts -p 'test_*.py' -v` | **PASS** | **8/8** (`test_demo.py`, `test_package.py` — Stage 8 birim testleri) |
| `python3 -m unittest discover -s network/tests -v` | **PASS** | **27/27** (`test_network.py`, `test_chaincode.py`; `docker compose config` daemon gerektirmeden render edildi, gerçek betik `bash -n` sözdizim kontrolünden geçti) |
| `python3 -m unittest backend/test/test_stage7.py -v` | **PASS** | **2/2** (istatistik yardımcı fonksiyonları) |
| SHA-256 doğrulama: `dist/AgroChain-competition.sha256` vs. gerçek arşiv özeti | **PASS** | Birebir eşleşiyor |
| Arşiv manifesti çapraz kontrolü: `package-demo.py`'nin kendi `permitted()` mantığı temiz `git ls-files` + `AGENTS.md` üzerinde simüle edilip arşivin 146 kaydıyla (145 dosya + `PACKAGE_MANIFEST.json`) `diff` edildi | **PASS** | Birebir eşleşiyor; beklenmeyen/eksik dosya yok |
| `PACKAGE_MANIFEST.json` içindeki 145 SHA-256'nın tümü, çıkarılan arşiv dosyalarına karşı yeniden hesaplandı | **PASS** | 0 uyuşmazlık |
| Arşiv içinde `.pem/.key/.p12/.sqlite/.env/tokens.json/_sk` deseni taraması | **PASS** | Bulunamadı (yalnız kaynak koddaki dize sabitleri/alan adları eşleşti, gerçek gizli dosya yok) |
| Çalıştırılabilir betiklerin (`*.sh`, `*.py`) yürütme izinlerinin kaynak↔arşiv karşılaştırması | **PASS** | Fark yok |
| Python `re` ile AI-1 regex eşdeğerlik kontrolü (33/32/8/7 karakter sınır durumları) | **PASS** | Beklenen davranış doğrulandı |
| `mvn test` (backend, tam JUnit takımı: `AnalysisTest`, `BackendTest`, `GatewayMeasurementTest`, `GatewayRecoveryTest`) | **BLOCKED** | `repo.maven.apache.org` bu sandbox'ın ağ politikasınca 403 ile engelleniyor; iş yerinden bağımsız, kod incelemesiyle telafi edildi (bkz. kod analizi) |
| `make stage3-check` / `chaincode-integration` / `chaincode-restart-check` / `privacy-integration` | **BLOCKED** | Docker daemon bu kapsayıcıda çalışmıyor (`docker info` başarısız) |
| `make backend-integration` / `backend-gateway-test` / `stage6-integration` / `stage6-ui-test` | **BLOCKED** | Canlı Fabric ağı ve derlenmiş backend gerektirir; Docker yok |
| `make stage7-check` / `stage8-rehearse` / `stage8-offline-test` (Playwright) | **BLOCKED** | Docker/canlı ağ gerektirir |
| Kullanıcının kendi makinesindeki `backend/target/agrochain-backend-0.3.0.jar` derleme çıktısı | **Tarihsel kanıt** | O makinede Maven erişiminin çalıştığını gösterir; bu sandbox'ın kısıtı yalnız buraya özgüdür |

**Not:** Yukarıdaki BLOCKED satırlarının hiçbiri "geçti" ya da "başarısız
oldu" şeklinde yorumlanmamıştır — kullanıcının açık talimatı doğrultusunda,
engelin somut nedeni (Docker daemon kapalı / Maven Central ağ politikasınca
engelli) belirtilmiş ve koşulamayan test alanı, mümkün olan en kapsamlı
manuel kod incelemesiyle (bu raporun "Bulgular" ve dosya-bazlı analiz kısmı)
telafi edilmiştir.

## Kabul kriterleri (etkilenen aşamalar)

Hiçbir aşamanın zorunlu kabul kriteri bu incelemeyle **yeniden açılmadı**.
AI-1 düzeltmesi Stage 5'in "girdi doğrulama/hata eşleme" kriterini daha sıkı
hale getirdi (zayıflatmadı); Stage 5 kabul kriterleri hâlâ karşılanıyor.
AI-2 hiçbir kabul kriterini etkilemedi (kod değişmedi). AI-3–AI-6, önceki
incelemenin bulduğu kapanmamış kriterlerin şu an tamamının karşılandığını
teyit etti; bu inceleme onları yeniden açmadı, yalnız doğruladı.

Bu incelemede **Maven ve Docker'a bağımlı testler koşulamadığından**,
Stage 5/6/7/8'in yalnız *canlı entegrasyon* kanıtı gerektiren kısımları
(örn. `backend-gateway-test`, `stage6-ui-test`, `stage8-rehearse`) bu
ortamda **yeniden koşulup PASS olarak kanıtlanamadı** — bunlar kullanıcının
kendi makinesinde daha önce PASS olarak kapatılmıştı (`docs/STAGE7_REPORT.md`,
`docs/STAGE8_REPORT.md`) ve bu inceleme onları bozacak hiçbir değişiklik
yapmadı, ama bu ortamda bağımsız olarak yeniden doğrulayamadı da. Bu, "kanıt
eksikliği" olarak açıkça işaretlenmiştir; "geçti" diye sunulmamıştır.

## Kalan riskler, varsayımlar, doğrulanamayanlar ve ileriye bırakılanlar

- **Varsayım:** Kullanıcının makinesi ile bu incelemedeki klon arasındaki tek
  fark, incelemenin kendisinin uyguladığı iki dosyalık AI-1 düzeltmesidir
  (aşağıdaki "Demo/Kurulum/Paket son durumu" bölümünde cihaza nasıl
  yazılacağı belirtilmiştir).
- **Doğrulanamayan:** `mvn test`, canlı Fabric entegrasyon testleri, Playwright
  UI testi (`stage6-ui-test`, `stage8-offline-test`) ve Stage 8 provası
  (`stage8-rehearse`) bu sandbox'ta çalıştırılamadı (yukarıya bakınız).
  Öneri: kullanıcı kendi makinesinde `make backend-test chaincode-check
  stage6-ui-test` çalıştırarak AI-1 düzeltmesinin JUnit'te de PASS ettiğini
  doğrulamalı.
- **Küçük, kasıtlı olarak düzeltilmeyen gözlem (bulgu değil):**
  `docs/DATA_CONTRACTS.md` satır 67, yeni ID'lerin "uppercase base32" ile
  kodlanacağını söylüyor; ancak `Crypto.id()` (backend) fiilen büyük harf
  onaltılık (hex/base16) üretiyor. Üretilen ID hâlâ `[A-Z0-9]{8,32}` grameriyle
  tam uyumlu ve ≥128 bit rastgelelik sağlıyor; işlevsel, güvenlik veya kabul
  kriteri etkisi yok. Bu saf bir belge ifadesi tutarsızlığıdır — kullanıcının
  "saf stil tercihini gerçek hata gibi sınıflandırma" talimatı gereği bulgu
  tablosuna P-seviyeli bir madde olarak eklenmedi, yalnız burada şeffaflık
  için not edilmiştir. İstenirse tek satırlık bir belge düzeltmesiyle
  kapatılabilir.
- **Kapsam dışı bırakılan (geleceğe ertelenen) hiçbir yeni özellik fikri
  üretilmedi** — inceleme, kullanıcının talimatına uygun şekilde yalnız
  mevcut mimari/kabul sözleşmesinin gerektirdiği ölçekte çalıştı; yeni
  framework, mikroservis veya kurumsal entegrasyon önerilmedi.
- **"Hata bulunmadı" bir güvenlik garantisi değildir.** Chaincode, backend,
  frontend, ağ betikleri ve paket betiklerinin büyük çoğunluğu satır satır
  okundu ve mevcut otomatik testler (koşulabilenler) PASS etti; ancak bu,
  yalnızca *bu incelemenin bulabildiği* sorunların yokluğu anlamına gelir,
  mutlak güvenlik kanıtı değildir — özellikle Maven/Docker engeliyle
  koşulamayan canlı entegrasyon/UI/performans testleri için.

## Demo / kurulum / paket son durumu ve yeniden çalıştırma adımları

- **Kaynak paket (`dist/AgroChain-competition.tar.gz`):** Sağlama toplamı ve
  145 dosyalık manifest bu incelemede yeniden doğrulandı, tam eşleşme;
  gizli anahtar/token/ortam dosyası deseni bulunamadı; yürütme izinleri
  korunmuş. **Değişiklik gerekmiyor**, ancak AI-1 düzeltmesinden sonra
  kullanıcı isterse `make competition-package` ile yeniden üretip aynı
  doğrulamayı tekrarlayabilir (arşiv hash'i değişecektir, çünkü `Json.java`
  ve `BackendTest.java` içeriği değişti).
- **Çevrimdışı HTML / PPTX:** İçerik, `DEMO_3MIN.md`/`OFFLINE.md`/`TECHNICAL.md`
  ile satır satır karşılaştırıldı; %40/%85/%50 eşik rakamları, parti kimlikleri
  (`BAT-DEMONORMAL01`/`BAT-DEMOSUSPICIOUS01`) ve komut adları ("make demo-setup"
  vb.) her yerde tutarlı. `test-offline.cjs` testinin kodu, sayfanın gerçekten
  sıfır ağ isteği yaptığını ve "ÇEVRİMDIŞI" uyarısının her slaytta göründüğünü
  doğruluyor (bu test Playwright/Docker gerektirmediği için normalde
  koşulabilir olmalı, ancak bu sandbox'ta `playwright` Python değil Node
  paketi olduğundan ve tarayıcı ikili dosyaları ayrıca doğrulanmadığından
  bu incelemede fiilen çalıştırılmadı — **BLOCKED**, kod incelemesiyle
  telafi edildi).
- **Kullanıcının AI-1 düzeltmesini kendi makinesine alması için:** Bu
  konuşmadaki `device_commit_files` çağrısıyla yalnız iki dosya
  (`backend/src/main/java/org/agrochain/Json.java`,
  `backend/src/test/java/org/agrochain/BackendTest.java`) ve bu rapor
  (`docs/AI_REVIEW_REPORT.md`) cihaza yazılacaktır. Sonrasında:

  ```bash
  cd /home/mert/Projects/AgroChain
  git diff --stat                      # yalnız 3 dosya değişmeli
  make backend-test                    # AI-1 regresyon testini içerir
  make chaincode-test verify chaincode-check   # ağ zaten ayaktaysa
  ```

  Ağ hâlâ ayaktaysa `make network-down` gerekmez; ledger diskleri ve
  kimlikler korunur. Yeniden derleme yalnız `backend-build`/`backend-run`
  gerektirir.
- **Push/merge/yayın yapılmadı; hiçbir yarışma dosyası harici bir yere
  gönderilmedi** — kullanıcının açık talimatına uygun olarak.

## Sonuç

**Result: PASS** (bu incelemenin kapsamı ve bu ortamda koşulabilen testler
için) — **BLOCKED maddelerle birlikte**, canlı Fabric entegrasyonu, Maven
JUnit takımı ve Playwright UI testleri bu sandbox'ta çalıştırılamadığından
bunlar için genel bir "tamamlandı" iddiası yapılmamaktadır. Zorunlu bir kabul
kriterinin başarısız olduğuna dair hiçbir kanıt yoktur; yalnız bazı kanıtlar
bu ortamın ağ/Docker kısıtları nedeniyle **eksiktir** ve öyle
işaretlenmiştir. Kullanıcının kendi makinesinde `make backend-test
stage3-check stage6-ui-test` çalıştırması, bu raporun tek açık kalan
doğrulama boşluğunu kapatacaktır.
