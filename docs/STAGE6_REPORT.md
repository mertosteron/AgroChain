# Aşama 6 tamamlanma raporu — 23 Eylül 2026

## Stage

Stage 6 — Açıklanabilir anomali analizi ve kullanıcı arayüzü.

## Completed

- Backend'in tamsayı fiyat hesabı, gerçek Fabric sözleşmesinde PDC girdilerinden
  yeniden hesaplanıp doğrulanıyor. Yanlış oracle önerisi reddediliyor.
- %40 normal ve %85 inceleme sinyali senaryoları, değişmez %50 pilot eşiğiyle
  doğrulandı. Tam eşik, eşik üstü, yuvarlama sınırı, düşen fiyat, sıfır alış ve
  eksik kanıt durumları test edildi. Eşik yasal sınır olarak sunulmuyor.
- Değerlendirme bekleme/yeniden deneme akışı yeniden başlatmadan sonra çalışıyor.
  Ortam ayarı zincir politikasıyla uyuşmazsa sonuç normal sayılmıyor.
- Regulator/oracle değerlendirme yapıyor; Regulator/reviewer incelemeyi açıp
  gerekçeyle sonuçlandırıyor. Özgün hesap ve sınıflandırma değişmiyor.
- Üretici, taşıyıcı, perakendeci işlemleri; denetçi karşılaştırması; tüketici QR
  sayfası Türkçe arayüzde mevcut. Arayüz işlemleri gerçek HTTP → Gateway → Fabric
  yolunu kullanıyor; gerekli özellikler taklit yanıtlarla çalışmıyor.
- Tüketici yanıtlarında fiyat, sınıflandırma, inceleme metni ve salt yok.
  Üretici/taşıyıcı/public-reader özel analize erişemiyor.
- Tarayıcı girişi, yükleme/boş/hata/bekleme durumları, klavye odakları, mobil
  görünüm, yerel QR üretimi ve açık SIMULATED etiketleri eklendi.
- Hızlı yenilemede yinelenen ayrıntı gösterimi bulundu, düzeltildi ve tarayıcı
  testi yeniden başarıyla çalıştırıldı.
- Mevcut **0.3.0 / sıra 6** sözleşmesinin dört kuruluştaki tanımı ve çalışması
  yeniden doğrulandı; ağ verileri silinmedi veya gereksiz yükseltme yapılmadı.
  Arayüz `make backend-run` ile `http://localhost:8080` adresinde açılır.

## Changed Files

Toplam **45 dosya**. Teslim yöntemi: kullanıcı isteğiyle her dosya ayrı commit
olarak mevcut `main` dalına alınır ve her commit GitHub'a ayrı push edilir.

```text
.github/workflows/stage3-chaincode.yml
Makefile
README.md
backend/README.md
backend/pom.xml
backend/src/main/java/org/agrochain/Actor.java
backend/src/main/java/org/agrochain/Analysis.java
backend/src/main/java/org/agrochain/Api.java
backend/src/main/java/org/agrochain/ApiError.java
backend/src/main/java/org/agrochain/Application.java
backend/src/main/java/org/agrochain/AuthFilter.java
backend/src/main/java/org/agrochain/EvidenceService.java
backend/src/main/java/org/agrochain/FabricLedger.java
backend/src/main/java/org/agrochain/Json.java
backend/src/main/java/org/agrochain/LotQr.java
backend/src/main/java/org/agrochain/Projection.java
backend/src/main/java/org/agrochain/Workflow.java
backend/src/main/resources/static/app.js
backend/src/main/resources/static/consumer.html
backend/src/main/resources/static/consumer.js
backend/src/main/resources/static/index.html
backend/src/main/resources/static/style.css
backend/src/test/java/org/agrochain/AnalysisTest.java
backend/test/integration_test.py
backend/test/stage6_test.py
backend/test/ui_test.cjs
chaincode/agrochain/README.md
chaincode/agrochain/internal/agrochain/analysis.go
chaincode/agrochain/internal/agrochain/analysis_test.go
chaincode/agrochain/internal/agrochain/contract.go
chaincode/agrochain/internal/agrochain/domain.go
chaincode/agrochain/internal/agrochain/domain_test.go
chaincode/agrochain/internal/agrochain/private.go
chaincode/agrochain/internal/agrochain/queries.go
chaincode/agrochain/internal/agrochain/validation.go
docs/ARCHITECTURE.md
docs/DATA_CONTRACTS.md
docs/IMPLEMENTATION_PLAN.md
docs/OPEN_QUESTIONS.md
docs/PROJECT_REVIEW_ROADMAP_TR.md
docs/SECURITY_AND_PRIVACY.md
docs/STATE_MACHINE.md
docs/STAGE6_REPORT.md
docs/STAGE6_UI.md
network/config/chaincode.env
```

## Tests Executed

Ortam: mevcut Arch Linux yerel pilot; Fabric 2.5.15, dört peer ve üç Raft
orderer; Go 1.27.1; Java 21.0.12.1; Chromium 153.0.8010.52; Playwright 1.62.1.

| Çalıştırılan komut / kontrol | Sonuç |
| --- | --- |
| `PATH=/usr/bin:/bin:$PATH make stage6-check` içindeki `backend-build` ve `chaincode-test` | PASS — 20 Java birim testi ve 33 Go test fonksiyonu; domain paketinde %81,4, tüm Go kodunda %80,5 kapsam; canlı Gateway testi bu birim turunda açıkça atlandı |
| `PATH=/usr/bin:/bin:$PATH make chaincode-check stage6-integration backend-gateway-test` | PASS — dört kuruluşta sözleşme kontrolü, 14 gerçek HTTP/Fabric testi (104,910 saniye) ve gerçek geçerli/MVCC-geçersiz işlem sonucu kurtarma testi |
| `NODE_PATH=<Playwright modül dizini> CHROMIUM_PATH=/usr/bin/chromium make stage6-ui-test` | PASS — Chromium; iki parti karşılaştırması, yedi aktör işlemi, iki inceleme işlemi, tüketici gizliliği, eksik lot, çıkış, 375/768/1280 görünüm |
| `node --check` — app.js, consumer.js, ui_test.cjs | PASS |
| `git diff --check` | PASS |

**68 otomatik test** (33 Go + 20 Java birim + 14 canlı HTTP + 1 canlı Gateway)
ve ayrıca gerçek tarayıcı senaryosu başarıyla tamamlandı. Tarayıcı senaryosu
yenileme düzeltmesinden sonra yeniden geçti; yakalanan JavaScript istisnası yok.
104,910 saniye kabul testlerinin toplam süresidir; performans kapasitesi ölçümü değildir.

İlk `stage6-check` çalışması birim testlerinden sonra araç ortamının Docker erişim
kısıtında durdu. Kalan ağ kontrolleri izinli ortamda yukarıdaki birleşik komutla
başarıyla tamamlandı. Tarayıcı testi de ilk denemede yerel soket kısıtına takıldı;
izinli ortamda yeniden çalıştırıldı. Bunlar uygulama test hatası değildir.
HTTP testinde Python 3.14, test yardımcısının SQLite bağlantıları için kapanış
uyarısı verdi; tüm doğrulamalar geçti. GitHub CI tanımı
güncellendi; uzak CI çalıştırılmış gibi bir iddia yoktur.

Yerel kanıtlar Git dışında tutulur:

- `network/runtime/backend-acceptance/stage6-summary.json`: iki senaryonun gerçek
  parti/lot kimlikleri ve özel analiz sonuçları; yetkili yerel test dosyası.
- `network/runtime/backend-ui-acceptance/comparison.png`: %40/%85 karşılaştırması.
- Aynı dizinde `review.png`, `review-mobile.png`, `consumer-375.png`,
  `consumer-768.png`, `consumer-1280.png`: incelenen ekran görüntüleri.

## Acceptance Criteria

- [x] Jüri normal partiyi ve şüpheli partiyi inceleyebilir.
- [x] Geçmişler, alış/raf farkı, eşik ve gerekçe yan yana görülebilir.
- [x] Backend sonucu gerçek zincirde yeniden doğrulanır.
- [x] Tam %50, üzeri, düşen fiyat, sıfır alış ve eksik kanıt güvenle ele alınır.
- [x] İzin verilen insan inceleme geçişleri çalışır; geçersiz geçişler reddedilir.
- [x] Yetkisiz kuruluşlar özel fiyat/analiz/incelemeye erişemez.
- [x] Tüketici sayfası ve QR uç noktası özel ticari verileri açığa çıkarmaz.
- [x] Bekleyen değerlendirme yeniden başlatmada kaybolmaz.
- [x] Uygulama, test, dokümantasyon ve demo uyumu sağlandı.

## Assumptions

- Tek üretici, taşıyıcı, perakendeci ve düzenleyici kuruluşlu mevcut pilot korunur.
- Kurumların gerçek API erişimi yoktur. ÇKS, e-Fatura, HKS, U-ETDS simülasyondur.
- Zincir politikası CFG-PRICE001 / 5.000 baz puandır. Ortam ayarı yalnız bununla
  uyumu denetler; geçmiş kayıtların anlamını değiştirmez.
- Kullanıcılar yerel geliştirme anahtarlarını kurulum dosyasından alır; token
  sahipliği ilgili pilot rolüne erişim sağlar. reviewerRef kişisel kimlik değildir.

## Known Limitations

- Açık, aşamayı engelleyen kabul hatası yoktur.
- Bu tek makinede çalışan pilot; üretim SSO'su, yüksek erişilebilirlik, ülke çapında
  ölçek, gerçek kurum entegrasyonu veya kesin hile tespiti iddiası yoktur.
- Liste/değerlendirme taraması küçük veri kümesi içindir. Yerel dosya/PDC üyesi
  yöneticilere karşı gizlilik veya disk şifrelemesi sağlandığı iddia edilmez.
- UI bekleyen işlem kimliğini açık sayfada tutar. Sayfa kapanırsa operationId ile
  API üzerinden durum kontrolü gerekir. Backend günlüğü kalıcıdır.
- QR varsayılanı localhost:8080'dir. Telefon erişimi için güvenilir yerel ağ URL'si
  ve dinleme adresi ayrıca ayarlanmalıdır; dış ağa dağıtım yapılmadı.
- Üretici formu mevcut simülatörün 100 kg standart domates/Antalya senaryosudur.
- İnceleme sonucu hile/suç kararı veya fiyat artışının ekonomik gerekçesinin
  otomatik ispatı değildir. Nakliye yalnız bağlam olarak gösterilir.

## Deferred to Future Stages

- **Aşama 7:** birleşik test/ölçüm raporu, sıfırdan iki tekrar kurulumu, kontrollü
  gecikme ölçümleri, ortam ve örneklem açıklaması.
- **Aşama 8:** sabit demo seed'i, başka bilgisayarda kurulum paketi, 10–12 slayt,
  üç dakikalık sunum ve çevrimdışı video/alternatif.

## Result

**PASS — Aşama 6 tamamlandı.** Aşama 7'ye otomatik geçilmedi.

Kullanım, API ekleri ve tekrar çalıştırma: [Aşama 6 rehberi](STAGE6_UI.md).
