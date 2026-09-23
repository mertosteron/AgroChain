# Aşama 7 — Test ve ölçüm raporu

## Stage

Stage 7 — Testing and Measurement. 23 Eylül 2026.

## Completed

- Mevcut ağ, chaincode, PDC, backend ve gerçek tarayıcı testleri tek sıralı
  `make stage7-check` komutunda birleştirildi.
- Normal, şüpheli ve belge tahrifatı senaryoları; kimlik/yetki, miktar, yaşam döngüsü,
  tekrar, MVCC, kaynak imzası ve ağ/backend kurtarma kontrolleri korunarak çalıştırıldı.
- Gateway gönderiminden VALID onayına kadar, ortak/özel sorgu ve HTTP tam rota
  sürelerini ölçen araçlar eklendi. Süre, boyut ve zincir işlem kimliği kaydediliyor.
- İki temiz tekrar için ayrı kimlik, genesis, backend veritabanı ve yedi ayrı test
  ledger diski kullanan kurulum eklendi. Mevcut diskler/kimlikler korunur.
- Disk öneki kimlik manifestine bağlandı; yanlış önekle temizleme reddediliyor.
  İstatistik yöntemi ve temizleme sınırları otomatik testlerle doğrulandı.
- Kabul testlerindeki SQLite bağlantıları açık bırakılmayacak şekilde düzeltildi.
- Temiz kurulumun ortaya çıkardığı iki hata giderildi: `umask 077` altında üretilen
  chaincode imajına açık çalıştırma izni verildi; doğrulanmış geçici dizinlerdeki
  salt okunur Go önbelleğinin güvenle temizlenmesi sağlandı.
- CI aynı kapıyı çalıştıracak şekilde güncellendi. Uzak CI bu çalışmada çalıştırılmadı.

## Changed Files

```text
.github/workflows/stage3-chaincode.yml
ARCHITECTURE.md
Makefile
README.md
backend/src/test/java/org/agrochain/GatewayMeasurementTest.java
backend/test/integration_test.py
backend/test/stage7_measure.py
backend/test/stage7_run.py
backend/test/test_stage7.py
chaincode/agrochain/Dockerfile
network/README.md
network/compose/compose-agrochain.yaml
network/scripts/network.py
network/tests/test_network.py
docs/STAGE7_REPORT.md
docs/STAGE7_TESTING.md
docs/STAGE7_MEASUREMENTS.json
docs/ARCHITECTURE.md
docs/IMPLEMENTATION_PLAN.md
docs/PROJECT_REVIEW_ROADMAP_TR.md
```

## Environment

| Alan | Değer |
| --- | --- |
| Tarih / kaynak | 23 Eylül 2026; temel commit `60fb465287672e5e79a682d979a3ba7813b32411` + bu aşamanın çalışma ağacı |
| Makine | Intel Core i7-12650H; 16 mantıksal CPU; 32.553.244 kB görünür RAM |
| İşletim sistemi | Arch Linux x86-64; kernel 7.2.6-arch2-1; glibc 2.44 |
| Araçlar | Docker 29.8.1; Compose 5.5.1; Python 3.14.7; Go 1.27.1; Temurin Java 21.0.12.1+1 |
| Ağ | Fabric 2.5.15; 4 peer / uygulama MSP; 3 Raft orderer; 4 TLS CCAAS servisi; tek makine; LevelDB |
| Politika | Retailer + Regulator endorsement; 2 saniye BatchTimeout; blokta en çok 10 ileti |
| Tarayıcı | Playwright 1.62.1; Chromium 153.0.8010.52; 375/768/1280 piksel |
| Veri | 100 kg standart domates; kurum kaynakları SIMULATED; gerçek Fabric işlemleri |

## Tests Executed

Birleşik komut:

```bash
PATH=/usr/bin:/bin:$PATH \
NODE_PATH=<yerel Playwright modül dizini> CHROMIUM_PATH=/usr/bin/chromium \
make stage7-check JAVA="$PWD/network/tools/jdk-21.0.12.1+1/bin/java"
```

| Bileşen | Çalıştırılan kontrol | Durum |
| --- | --- | --- |
| Ağ/temizleme güvenliği | `make test` — 27 test; salt okunur önbellek regresyonu dahil | PASS |
| Go domain/gizlilik/analiz | `make chaincode-test` — 33 test, race + vet; domain %81,4 kapsam | PASS |
| Java birim | `make backend-build` — 20 test; iki canlı test bu turda açıkça atlandı | PASS |
| Sayısal özet | `make stage7-unit` — 2 test | PASS |
| Belge protokolü | `make privacy-vectors` — Java/Go ortak vektör, imza/tahrifat | PASS |
| Ağ/kimlik | `make network-up chaincode-check verify` — TLS, MSP, kanal ve peer restart | PASS |
| Gerçek domain/endorsement | `make chaincode-integration` — 6 test | PASS |
| Gerçek PDC/tahrifat/MVCC | `make privacy-integration` — 9 test | PASS |
| HTTP ve simülatörler | `make stage6-integration` — 14 test | PASS |
| Gateway kurtarma | `make backend-gateway-test` — 1 test, VALID ve MVCC-geçersiz durum | PASS |
| Gerçek tarayıcı | `make stage6-ui-test` — aktör/inceleme, iki parti, mobil QR ve gizlilik | PASS |
| Süre ölçümü | HTTP ölçüm adımı ve `GatewayMeasurementTest` — 1 canlı Java testi | PASS |
| İki temiz kurulum | `make stage7-reproduce` — ayrı diskler/kimliklerle 2 × (9 PDC + 14 HTTP testi) | PASS |
| Özgün veri kalıcılığı | Geri açılan ağda blok 353 VALID işlemi, eski parti ve özel alış sorgusu | PASS |
| Kanıt dosyası | JSON örneklerinden medyan/p95 yeniden hesaplama ve özel alan kontrolü | PASS |
| Değişiklik biçimi | `git diff --check` | PASS |

İlk deneme araç ortamının Docker erişim kısıtında durdu. İzinli ortamda ilk canlı
kontrol, önceden durmuş orderer nedeniyle başarısız oldu. Birleşik komuta ağ açma
adımı eklendi; sonraki tur yukarıdaki kapıları geçti. İlk başarısız denemeler PASS
sayılmadı. İlk temiz tur ayrıca konteyner çalıştırma izni ve salt okunur önbellek
temizliği hatalarını yakaladı; bu tur FAIL olarak saklandı. Düzeltmeden sonra
`make chaincode-package network-up chaincode-check` ve `make test stage7-unit`
başarıyla çalıştırıldı. Mevcut ağ tekrar açıldı. Temiz tekrar, `make stage7-reproduce`
ile ayrı çalıştırıldı ve iki tur da PASS oldu. Tek bir düzeltilmiş `stage7-check`
çağrısının baştan sona geçtiği iddia edilmez: etkilenen kontroller ayrı komutlarla
yeniden doğrulandı. Uzak CI sonucu yerel testlerden varsayılmadı.

**113 farklı otomatik test**: 27 ağ/temizleme + 33 Go + 20 Java birim + 2 istatistik
+ 6 canlı domain + 9 canlı PDC + 14 canlı HTTP + 1 Gateway kurtarma + 1 Gateway
ölçüm testi. Tarayıcı, çapraz dil vektörü ve ölçüm rota kontrolleri ayrıca geçti.
Temiz turlarda tekrarlanan testler bu farklı test sayısına yeniden eklenmedi.

## Measurements

| Ölçüm | n | Medyan (ms) | p95 (ms) | Min–maks (ms) |
| --- | ---: | ---: | ---: | ---: |
| Gateway gönderim → VALID | 30 | 2.022,271 | 2.042,957 | 2.018,167–2.044,227 |
| Ortak GetBatch sorgusu | 30 | 2,998 | 3,941 | 2,637–4,432 |
| Özel GetPurchase sorgusu | 30 | 3,340 | 4,130 | 2,671–6,121 |
| HTTP iş komutu → COMMITTED | 42 | 2.045,039 | 2.070,879 | 2.038,847–2.074,508 |
| Tam rota + analiz + tüketici görünümü | 6 | 16.523,849 | 16.592,872 | 16.448,516–16.592,872 |

Her iki ölçüm aracında **0 hata**. HTTP örnekleri yedi komut türünden altışar adet;
tam rota üç normal + üç şüpheli partidir. Isınma olarak bir Gateway işlemi ve iki
HTTP rota dışlandı. Eşzamanlılık 1; p95 en yakın üst sıra (`ceil(0,95 × n)`).
Tüm örnekler monoton saatle alınmıştır. Yaklaşık 2 saniyelik onay süreleri ağın
2 saniyelik blok bekleme ayarıyla uyumludur; bu bir kapasite/azami TPS ölçümü değildir.

HTTP gövde boyutu 301–499 byte; endorse edilmiş Gateway işlem boyutu 9.432–9.444
byte; ortak sorgu yanıtı 597–599 byte, özel sorgu yanıtı 1.168 byte. Bu boyutlar
TLS/HTTP üstbilgilerini içermez; özel içerikler rapora kopyalanmaz. HTTP süresi
doğrulama/simülatör/endorsement/commit'i içerir; Gateway gönderim süresi önceden
endorse edilen işlemin gönderimiyle başlar. Tam rota insan beklemesini dışlar,
100 ms aralıklı analiz/projeksiyon görünürlük kontrolünü içerir.

Zincirle eşleştirilebilen örnek: blok **353**, işlem
`07a77c24061608dc82bcf05bcfee713dae1f7eb3db0b14ae093b3ea843588061`.
Tüm örnek işlem kimlikleri ham JSON kanıtlarında bulunur.

Yöntem: [test kılavuzu](STAGE7_TESTING.md). Ham kanıtlar Git dışındaki
`network/runtime/stage7/http-measurements.json` ve `gateway-measurements.json`
dosyalarındadır. `check-report.json` ilk birleşik turdaki temiz kurulum hatası
nedeniyle FAIL olarak korunur; önceki başarılı bileşenleri ve ölçümleri içerir.
Süre/boyut/işlem kimliklerinin paylaşılabilir kopyası
[STAGE7_MEASUREMENTS.json](STAGE7_MEASUREMENTS.json) dosyasındadır. Özel belge,
salt, token ve anahtar içermez; ham sürelerden medyan/p95 yeniden hesaplanabilir.

## Clean Reproduction

Her tur sıfır ledger diskleri, yeni kimlik/genesis ve boş backend veritabanıyla
başladı. Önceden indirilmiş araçlar ve Go önbelleği kullanıldı; testler mevcut
veritabanına yeni kimlik eklemekle sınırlı değildi.

| Tur | Kurulum | PDC testleri | HTTP testleri | İş sonuçları |
| --- | --- | --- | --- | --- |
| 1 | PASS, 31,290 s | 9/9 PASS | 14/14 PASS | %40 NO_SIGNAL; %85 REVIEW_REQUIRED |
| 2 | PASS, 31,349 s | 9/9 PASS | 14/14 PASS | %40 NO_SIGNAL; %85 REVIEW_REQUIRED |

Her iki turda eşik %50; orijinal belge geçerli, değişmiş belge reddedildi;
yinelenen işlemler/PDC dışlama/MVCC/restart kontrolleri geçti. Bağımsız genesis kanıtı:

- Tur 1: `a6903c0efac750b258bde503ad6d5c8e9e341fac96b739956c65307d5ae38772`
- Tur 2: `d2cc0926c0a6e80a9da6646fa31709fe4f9a53eefa032a91ad4c96e57d05e4f9`

Her turdaki yedi test diski temizlendi. Özgün yedi disk yeniden yaratılmadı,
kimlik manifesti değişmedi; ağ 7,982 saniyede geri açılıp sözleşme doğrulandı.
Ardından önceki ölçüm işleminin VALID kaldığı, eski perakendeci partisinin ve
yetkili özel alış kaydının okunabildiği ayrıca doğrulandı.
`network/runtime/stage7/reproduce-report.json`: **PASS**, `originalRestored=true`.

## Acceptance Criteria

- [x] Chaincode, yetki, miktar, durum ve yinelenen işlem kontrolleri geçti.
- [x] Gerçek üye/üye olmayan peer ve istemci PDC erişimi doğrulandı.
- [x] Orijinal belge doğrulandı; değiştirilmiş içerik reddedildi.
- [x] Backend/simülatörler, commit kurtarma ve tüketici gizliliği doğrulandı.
- [x] Gerçek tarayıcıda normal/şüpheli parti ve inceleme akışı geçti.
- [x] İki temiz kurulum aynı sonuçları üretti ve mevcut ağ geri açıldı.
- [x] Ölçüm örnekleri, ortam, medyan/p95 ve hata sayısı raporlandı.

## Assumptions

- Mevcut yarışma pilotu ve dört kuruluşlu model korunur. Gerçek devlet API erişimi yoktur.
- Geliştirme kimlikleri/salt'lar rastgeledir; tekrarda iş sonuçları eşit olmalıdır.
- Önceden indirilmiş araç ve Go bağımlılıkları temiz kurulumlarda tekrar kullanılır.
- Bu aşamada GitHub push veya sonraki aşamaya geçiş yapılmaz.

## Known Limitations

- Açık, aşamayı engelleyen kabul hatası yoktur. İlk başarısız denemeler ve bunlara
  yönelik düzeltmeler yukarıda ayrıca belirtilmiştir.

- Tek makinede, düşük eşzamanlılıkta yerel pilot; kapasite, TPS, SLA, ülke çapında
  ölçek veya üretim güvenliği iddiası yoktur.
- Temiz işletim sisteminde sıfır bağımlılık indirme/başka takım üyesi provası yapılmaz.
- Ölçümde JIT/işletim sistemi önbellekleri sıfırlanmaz; tekil Gateway ısınma örneği
  ve iki HTTP ısınma rotası dışlanır. Küçük örneklem p95'i kararlı bir kapasite tahmini değildir.
- Ağ düğümleri aynı yönetici ve fiziksel makinededir. Gizlilik PDC üyeliği sınırıdır;
  yetkili kurum yöneticisinden veya makine yöneticisinden sır saklama iddiası yoktur.
- Hata/normal kesintide geri açma denenir; SIGKILL/elektrik kesintisinde otomatik
  geri dönüş garanti edilmez. Kılavuzdaki elle kurtarma sınırları geçerlidir.
- İnsan fiyat incelemesi hile/hukuki ihlal kararı değildir; eşik pilot parametresidir.

## Deferred to Future Stages

**Aşama 8:** başka bilgisayarda kurulum provası, sabit demo seed komutu, 10–12
slayt, üç dakikalık demo ve çevrimdışı video/alternatif. Üretim/ulusal yük testi
bu pilotun tamamlanma koşulu değildir.

## Result

**PASS — Aşama 7 tamamlandı.** Aşama 8'e otomatik geçilmedi.
