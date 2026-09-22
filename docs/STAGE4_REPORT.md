# Aşama 4 tamamlanma raporu — 21 Eylül 2026

## Stage

Stage 4 — Privacy and Data Verification / Gizlilik ve Veri Doğrulama.
Yerel Fabric pilotunda tamamlandı. Aşama 5 uygulanmadı.

## Completed

- Üç gerçek Private Data Collection üzerinde alış belgesi, taşıma maliyeti ve
  perakende fiyatı yazma/okuma; kurum ve sertifika rolü denetimleri.
- Sertifikalarda zorunlu `agrochain.role`; eksik rol ve business çağrısı yapan admin
  reddi. Geliştirme kimliklerini ve dört simüle kaynak anahtarını hazırlayan araç.
- Bir kez yapılabilen Regulator-admin güven kaydı; çakışan kaynak anahtarı reddi.
- Ed25519 kaynak imzası, tuzlu SHA-256 belge özeti, işleme/partiye bağlama,
  kaynak belge/nonce ve işlem tekrarının önlenmesi.
- Gerçek yedi adımlı ürün akışı: üretici kaydından teslim ve özel fiyat bildirimine.
  Mülkiyet ve muhafaza değişimleri, makbuzlar ve olaylar canlı bloklarla doğrulandı.
- Yetkili istemci okuması, yetkisiz istemci reddi ve koleksiyona üye olmayan peer'de
  yetkili Regulator kimliğiyle dahi özel verinin alınamaması.
- Orijinal dosya doğrulaması; değiştirilmiş dosya/açılış/imza/bağlam/anahtar reddi.
- Ortak blok yazılarında özel alan bulunmaması, olay alanlarının izin listesi,
  peer ve chaincode servis kayıtlarında örnek gizli tuzların bulunmaması.
- Gerçek eşzamanlı işlem: aynı kayıt için iki başvurunun biri VALID, diğeri
  MVCC_READ_CONFLICT (blok doğrulama kodu 11). Tek ürün/işlem kaydı oluştu.
- Dolu public/PDC kayıtlarıyla veri silmeden ağ kapatma/açma doğrulaması.
- Java–Go kriptografik test vektörü, birleşik `stage4-check`, güncellenen CI ve
  [çalıştırma/API rehberi](STAGE4_PRIVACY.md).

## Changed Files

Bu aşamada değişen dosyalar aşağıdadır. Depoda önceki aşamalardan kalan commit
edilmemiş dosyalar korunmuştur; aşağıdaki liste genel `git status` listesi değildir.

- `chaincode/agrochain/internal/agrochain/`: `evidence.go`, `private.go`,
  `privacy_test.go`, `contract.go`, `domain.go`, `models.go`, `queries.go`,
  `validation.go`, `contract_test.go`, `domain_test.go`, `invariants_test.go`,
  `validation_test.go`.
- `chaincode/agrochain/test/VerifyVector.java`.
- `chaincode/agrochain/test/integration/`: `evidence_support.py`, `privacy_test.py`,
  `fabric_test.py`, `restart_test.py`.
- `network/scripts/`: `privacy-prepare.py`, `chaincode-client.sh`, `chaincode.sh`.
- `network/config/chaincode.env`, `network/tests/test_chaincode.py`, `Makefile`,
  `.github/workflows/stage3-chaincode.yml`.
- `README.md`, `ARCHITECTURE.md`, `network/README.md`,
  `chaincode/agrochain/README.md`, `chaincode/agrochain/collections/README.md`,
  `docs/ARCHITECTURE.md`, `docs/DATA_CONTRACTS.md`, `docs/SECURITY_AND_PRIVACY.md`,
  `docs/IMPLEMENTATION_PLAN.md`, `docs/OPEN_QUESTIONS.md`,
  `docs/PROJECT_REVIEW_ROADMAP_TR.md`, `docs/STAGE4_PRIVACY.md`, bu rapor.

## Tests Executed

Yerel ortam: Linux x86-64, Fabric 2.5.15, Docker 29.8.0 / Compose 5.5.1,
Go 1.27.1, Python 3.14.7 / cryptography 50.0.1, OpenJDK 26.0.2.1;
dört kurum peer'i, üç Raft orderer ve dört TLS chaincode servisi.

| Gerçekten çalıştırılan komut | Sonuç |
| --- | --- |
| `PATH=/usr/bin:/bin:$PATH make privacy-prepare` | PASS — geliştirme rolleri ve kaynak anahtarları |
| `make chaincode-format chaincode-test` | PASS — format, vet, race, test ve derleme |
| `PATH=/usr/bin:/bin:$PATH make test` | PASS — 24 ağ/paket/güvenli temizlik testi |
| `make privacy-vectors JAVA=/usr/lib/jvm/java-26-openjdk/bin/java` | PASS — Java taahhüt/imza/bozulma vektörleri |
| `PATH=/usr/bin:/bin:$PATH make chaincode-upgrade CHAINCODE_VERSION=0.2.1 CHAINCODE_SEQUENCE=5` | PASS — dört kurumda tanım ve sürüm doğrulandı |
| `PATH=/usr/bin:/bin:$PATH make stage4-check JAVA=/usr/lib/jvm/java-26-openjdk/bin/java` | PASS — bütünleşik son kabul kontrolü |

Son birleşik kontrol; 24 ağ testi, 30 üst düzey Go test işlevi ve alt vakaları,
Java vektörü, canlı TLS/kanal/kurum doğrulaması, 6 Fabric regresyon testi ve
9 canlı gizlilik testi içerir. Go iş paketi statement kapsamı %82,7;
sunucu başlangıcını da içeren toplam %81,7. Bu kapsam canlı Docker testlerini
ölçmez. Altı canlı regresyon testi 17,877 saniye, dokuz gizlilik testi 31,383 saniye
sürdü; bunlar kapasite veya gecikme benchmark'ı değildir.

İlk sandbox çalıştırmasında port açma testleri ortam izni nedeniyle çalışamadı;
izinli yeniden çalıştırma ve son birleşik kontrol geçti. Aynı sürüm numarasıyla
yükseltme denemesi araç tarafından güvenli biçimde reddedildi; yeni 0.2.1 sürümüyle
yükseltme tamamlandı. Bunlar gizlenmiş PASS sonuçları değildir.

Son yerel sürüm **0.2.1 / sequence 5**. Binary SHA-256:
`49e92e96494e8d7297fc49bc9b38fad5ae6f06162338248728becd3f3e4f2a16`.
Son kabul partisinin kimliği `BAT-2WVD5NFYZJ7N2V4CIB2BBFHCYE`; ürün işlemleri
60–66 numaralı bloklarda. Eşzamanlı test sonrasında yükseklik 68; yeniden başlatma
sonrası aynı kaldı. Yerel, git dışında tutulan kanıt:
`network/runtime/chaincode/privacy-evidence.json`. Raporda özel değer/tuz/anahtar
yayımlanmadı. CI tanımı güncellendi; uzak CI çalıştırıldığı iddia edilmiyor.

## Acceptance Criteria

- [x] Yetkili taraf amaçlanan özel veriye gerçek PDC'den erişebiliyor.
- [x] Yetkisiz taraf erişemiyor; üye olmayan peer özel veriyi sağlayamıyor.
- [x] Tasarlanan ortak parti verisi dört kuruma da açık kalıyor.
- [x] Orijinal belge geçerli imza, tuzlu özet ve dosya özetiyle doğrulanıyor.
- [x] Değiştirilmiş belge/açılış doğrulamadan geçmiyor.
- [x] Aşama 3'ten devreden doğrulanmış canlı ürün akışı tamamlanıyor.
- [x] Canlı tekrar, MVCC çakışması, olaylar ve dolu public/private yeniden
  başlatma kontrolleri geçiyor.
- [x] Zorunlu roller, Go–Java vektörü, blok/olay/kayıt sızıntı kontrolleri geçiyor.

## Assumptions

- Kurum verileri açıkça **SIMULATED** test belgeleridir; gerçek devlet servisi yoktur.
- Tek bilgisayarlı pilotun yöneticisi ve geliştirme CA'ları güven sınırına dahildir.
- Test sırasında ağda başka yazıcı çalışmaz; rastgele ID ve tuzlar tekrar çalıştırmayı
  sağlar. Sabit iş tutarları değiştirilmeden yeni test kayıtları eklenir.
- Bir kez kaydedilmiş güven anahtarları ledger ile birlikte korunur.
- Kanonik JSON, tanımlı ASCII/tamsayı şema alt kümesidir; genel Unicode/ondalıklı
  JCS işlemcisi olduğu iddia edilmez.

## Known Limitations

- Belgenin orijinal bayt kontrolü çalışan off-chain kabul yardımcısındadır;
  Spring Boot API, kurumsal servisler ve kontrollü dosya arşivi henüz yoktur.
- PDC yetkili alıcının veriyi dışarı kopyalamasını veya ortak host yöneticisini
  engellemez. Özel sorgular evaluate/query olarak kullanılmalıdır; bunları submit
  etmek yanıtın blokta kalmasına yol açabilir. Backend izin listesi Aşama 5 işidir.
- Geliştirme sertifikaları ve değişmez kaynak güven kaydı kullanılır; üretim
  sertifika/anahtar yenileme-iptal yönetimi yoktur. CCAAS karşılıklı TLS yoktur.
- Canlı MVCC kanıtı aynı oluşturma komutunun yarışıdır; her handoff yarışının
  ayrı canlı kanıtı olduğu iddia edilmez. Handoff değişmezleri ayrıca birim testlidir.
- Son kaynak durumu, korunan mevcut ağda yükseltmeyle doğrulandı; yeni bir ikinci
  bilgisayarda sıfırdan kurulmuş olduğu veya uzak CI geçtiği iddia edilmez.
- Anomali sonucu, kullanıcı arayüzü, QR görünümü ve üretim ölçeklenebilirliği yoktur.
- Açık zorunlu Aşama 4 kabul kusuru kalmadı.

## Deferred to Future Stages

- **Aşama 5:** Spring Boot/Gateway, kimlik oturumu, evaluate/submit ayrımı,
  ÇKS/e-Fatura/HKS/U-ETDS simülatör servisleri ve güvenli off-chain belge deposu.
- **Aşama 6:** Açıklanabilir anomali, denetçi işlemleri ve tüketici/aktör arayüzleri.
- **Aşama 7–8:** Geniş uçtan uca ölçümler, temiz demo tekrarları, başka bilgisayarda
  kurulum provası ve yarışma teslim paketi.

## Result

**PASS — Aşama 4, tanımlı yerel pilot kapsamındaki zorunlu kabul ölçütlerini geçti.**
Sonraki aşamaya otomatik geçilmedi.
