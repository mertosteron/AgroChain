# Aşama 6 — Fiyat analizi ve arayüz

## Çalıştırma

Önkoşul: Aşama 5 ağı, rol sertifikaları, kaynak anahtarları ve backend araçları
hazır olmalı. Yeni bilgisayarın tam kurulumu için kök README ve ağ rehberini izleyin.

```bash
make network-up
# Mevcut ağda 0.2.1 / sıra 5 kullanılıyorsa:
make chaincode-upgrade CHAINCODE_VERSION=0.3.0 CHAINCODE_SEQUENCE=6
make backend-build
make backend-run
```

Başka bir ağda son tanımın bir sonraki sırasını kullanın. Yeni kanalda
`make chaincode-deploy` sıra 1 kullanır. Kaynak anahtarlarını ve veri birimlerini
silmeyin. Uygulama: `http://localhost:8080`.

`make backend-prepare` ile oluşturulan geliştirme anahtarları
`network/runtime/backend/tokens.json` içindedir. Kullanılacak rolün değerini erişim
alanına girin; dosyayı paylaşmayın veya Git'e eklemeyin. Anahtar HTML içinde
bulunmaz, tarayıcı deposuna yazılmaz; sayfa kapanınca oturum biter. API rolü anahtardan
seçer; zincir ayrıca sertifika rolünü kontrol eder.

| Oturum | Ekrandaki işlemler |
| --- | --- |
| producer:producer | Parti oluştur, taşıyıcıya teslim teklif et |
| logistics:carrier | Teslim al, nakliye belgesini kaydet, mağazaya teslim teklif et |
| retailer:retailer | Teslim al, raf fiyatını bildir, yetkili analiz/karşılaştırma ve QR |
| regulator:auditor | İki partinin fiyatlarını ve geçmişini karşılaştır |
| regulator:reviewer | Denetçi görünümü; incelemeyi aç ve gerekçeyle sonuçlandır |
| regulator:oracle | Backend değerlendirme servisinin kimliği; insan incelemesi yapamaz |
| Kimliksiz tüketici | `/consumer.html?lot=LOT-...`: kaynak, miktar, teslimler |

Pilot formu 100 kg standart domates ve Antalya ile sınırlıdır; bu mevcut imzalı
simülatör sözleşmesidir. Kurum verileri **SIMULATED**, işlemler gerçek yerel
**FABRIC** olarak görünür. Fiyat, gerçekleşmiş satış değil vergi hariç raf teklifidir.

## Karşılaştırılabilir iki senaryo

1. Üretici partiyi oluşturup teslim teklif eder; taşıyıcı kabul eder, nakliyeyi
   kaydeder ve mağazaya teklif eder; perakendeci teslim alır.
2. Normal parti için 28,00 TL/kg; ikinci parti için 37,00 TL/kg raf fiyatı girilir.
3. İmzalı alış belgesi her ikisinde 20,00 TL/kg, nakliye toplamı 200,00 TL'dir.
4. Değerlendirme tamamlanana kadar ekran **bekleniyor** der. Yenile ile güncellenir.
5. Denetçi karşılaştırmada %40 ve %85 artışı, %50 pilot eşiğini ve iki geçmişi görür.
6. İncelemeci şüpheli partide incelemeyi başlatır; gerekçe ve sonuç girer.
   Sonuç sınıflandırmayı değiştirmez; işlem geçmişi silinmez.

Oran `(raf − alış) / alış × 100`; karar tam sayı çarpımlarıyla verilir.
Tam %50 sinyal değildir; %50'yi aşan değer sinyaldir. Görüntülenen baz puan
matematiksel aşağı yuvarlanır; sınırdaki karar yuvarlanmış değerden hesaplanmaz.
Nakliye oran hesabından düşülmez, açıklayıcı bağlamdır. Eşik yasal sınır veya
sinyal bir suç/ihlâl kararı değildir.

`ANOMALY_PRICE_INCREASE_THRESHOLD_PERCENT=50` ayarı ondalık olarak tam baz puana
dönüşmeli ve değişmez `CFG-PRICE001` zincir politikasıyla uyuşmalıdır. Ortam
değişkenini değiştirmek mevcut politika veya sonuçları değiştirmez. Uyuşmazlık
değerlendirmeyi bekletir; ayarı düzeltip backend'i yeniden başlatmak güvenli
yeniden denemeyi sağlar. Yeni eşik, gelecekte yönetilen yeni politika sürümü gerektirir.

## HTTP ekleri

Mevcut `command/scenario/privateInput` zarfı ve aynı operationId'yi taşıyan
`Idempotency-Key` kullanılır. Yeni komutlar da batch expectedVersion'ı artırır.

| Uç nokta | Kimlik / davranış |
| --- | --- |
| GET /api/v1/session | Oturumdaki org/role |
| GET /api/v1/batches | Yetkili ortak parti listesi; geçerli olay projeksiyonu |
| GET /api/v1/batches/{id}/anomaly | Perakendeci, denetçi, incelemeci, oracle; özel sonuç veya EVALUATION_PENDING |
| GET /api/v1/batches/{id}/reviews | Aynı özel okuma yetkisi; eklemeli inceleme kayıtları |
| POST /api/v1/batches/{id}/anomaly-evaluations | Yalnız oracle; EvaluatePrice |
| POST /api/v1/anomalies/{id}/review-actions | Yalnız reviewer; OpenReview/ResolveReview |
| GET /api/v1/public/lots/{id}/qr | Kimliksiz; tüketici bağlantısı içeren gerçek PNG QR |

EvaluatePrice payload: `{anomalyId,reportId}`, privateInput: `{}`. Backend alış,
nakliye, rapor ve politika kayıtlarını okur; kendi hesapladığı öneriyi transient
olarak gönderir. İstemci hesap veya fiyat seçemez. Zincir yetkili oracle'ın yanlış
önerisini de reddeder. Raporlar otomatik işlendiği için bu API'yi elle çağırmak
genellikle gerekmez.

OpenReview payload: `{anomalyId}`, privateInput: `{reviewerRef:"REV-PILOT001"}`.
ResolveReview aynı payload ve `{reviewerRef,outcome,explanation}` alır.
Sonuçlar: EXPLAINED, FOLLOW_UP_RECOMMENDED, INSUFFICIENT_EVIDENCE. Gerekçe 1–2.000
yazdırılabilir Unicode karakteri olmalı; kişisel veri kullanılmaz. reviewerRef
pilot içi etikettir; güvenilen kimlik Fabric Regulator/reviewer sertifikasıdır.
Arayüzde yalnız tek sentetik incelemeci vardır. Yeni rastgele kayıt saltları
backend tarafından üretilir. Kaynak belge kanonikleştirmesi ASCII/tamsayı olarak kalır.

## QR ve yerel ağ

QR varsayılan bağlantısı `http://localhost:8080/consumer.html?lot=...` olur.
`AGROCHAIN_PUBLIC_BASE_URL` güvenilir kurulum ayarıdır; istek Host başlığından
türetilmez. Telefonla açılacaksa base URL aynı yerel ağdaki sunucu adresine ve
ayrıca Spring `SERVER_ADDRESS` uygun ağ arayüzüne açıkça ayarlanmalıdır. Varsayılan
backend yalnız loopback dinler; dışarıya yayın veya TLS dağıtımı bu aşamada yapılmaz.
QR üretimi [ZXing 3.5.4](https://github.com/zxing/zxing/releases/tag/zxing-3.5.4)
ile yerelde gerçekleşir, harici QR servisine veri gönderilmez.

## Doğrulama

```bash
make stage6-check
# Ek gerçek tarayıcı kontrolü (bu çalışmada Node + Playwright 1.62.1 kullanıldı):
# Playwright modülü normal Node çözümlemesinde ya da NODE_PATH içinde bulunmalı.
CHROMIUM_PATH=/usr/bin/chromium make stage6-ui-test
```

Playwright kurulu değilse test araçlarını Git dışı bir dizine kurabilirsiniz:

```bash
npm install --prefix network/runtime/ui-tools --no-save playwright@1.62.1
network/runtime/ui-tools/node_modules/.bin/playwright install chromium
NODE_PATH="$PWD/network/runtime/ui-tools/node_modules" make stage6-ui-test
```

Bu yalnız test bağımlılığıdır; uygulamanın çalışması için Node, internet veya
Playwright gerekmez. Sistem Chromium'u varsa `CHROMIUM_PATH` indirme gereksinimini kaldırır.

Tarayıcı testi önce stage6-check'in ürettiği `stage6-summary.json` senaryolarını
kullanır; sonra UI üzerinden yeni bir parti ve inceleme oluşturur. Kendi backend'ini
18081 portunda başlatıp kapatır. HTTP kabul testleri 18080 portunu kullanır.
Ekran görüntüleri ve test verileri Git dışı `network/runtime/backend-ui-acceptance`
dizinindedir. Araçlar kişisel/üretim verisiyle çalıştırılmamalıdır.

## Sınırlar / sonraki aşamalar

- Tek backend, tek yerel ağ, basit manuel token girişi. Kurum SSO'su veya kullanıcı
  bazlı sertifika kayıt sistemi yoktur. PDC üyesi kuruluş yöneticisine karşı gizlilik
  iddiası yoktur.
- Liste ve değerlendirme taraması küçük pilot veri kümesi içindir; sayfalı operatör
  listesi ve performans ölçümleri Aşama 7 kapsamında değerlendirilir.
- Salt ve fiyatlar yetkili API'lerde bulunabilir, tüketici yanıtlarında bulunamaz.
- HTTP 202 veya bağlantı belirsizliğinde UI aynı işlemi tekrar kontrol eder.
  Açık sayfadaki bekleyen operationId saklanır; sayfa kapatıldıktan sonra operatör
  API'den operationId ile sonucu kontrol etmelidir. Arayüz sessizce yeni komut yaratmaz.
- Nihai sabit demo seed'i, kurulum paketi, sunum/video ve çevrimdışı alternatif
  Aşama 8'e; kapsamlı ölçümler Aşama 7'ye ertelidir.
