# Aşama 5 tamamlanma raporu — 22 Eylül 2026

## Stage

Stage 5 — Backend and Institutional Simulators.
Spring Boot API ile gerçek Fabric bağlantısı yerel pilotta tamamlandı.
Aşama 6 uygulanmadı; anomali değerlendirmesi ve arayüz henüz yoktur.

## Completed

- Spring Boot 3.5.16 / Java 21 uygulaması, Fabric Gateway 1.10.1 bağlantısı.
  Gerçek çalışmada bulunan Protobuf uyumsuzluğu, Gateway ile uyumlu protobuf
  4.33.4 ve gRPC 1.78.0 BOM sürümleriyle giderildi; regresyon testi eklendi.
- Sabit kurum/rol eşlemeli rastgele geliştirme erişim anahtarları; istemcinin
  X-MSP/X-Role başlıkları yetki vermiyor. Her işlem kendi kurumunun sertifikasıyla
  imzalanıyor. Özel veri taşıyan öneriler Retailer Gateway üzerinden yalnızca
  Retailer/Regulator onaylayıcılarına gidiyor.
- Aynı uygulamadaki değiştirilebilir InstitutionalAdapter sınırı üzerinden ÇKS,
  e-Fatura (alış ve taşıma faturası), HKS, U-ETDS imzalı simülatörleri. Her kaynak
  SIMULATED etiketli; gerçek devlet API'si kullanılmıyor.
- Backend, ledger'daki güven anahtarlarıyla imzayı, tuzlu özeti, miktar/taraf/işlem
  bağlarını ve orijinal dosya baytlarını doğrulayıp transient girdiyi oluşturuyor.
  Chaincode bağımsız doğrulama yapmaya devam ediyor; dış servis çağırmıyor.
- Yedi adımlı üretici → taşıyıcı → perakendeci HTTP akışı, özel fiyat bildirimi,
  paylaşılan kayıt ve yetkili özel veri sorguları.
- SQLite üzerinde dayanıklı işlem günlüğü, hazırlanmış işlem baytları/ID'si,
  tekrar önleme ve belirsiz sonuç kurtarma. Başarı yalnızca VALID commit sonrası.
  Aynı anahtarla farklı istek 409; belirsiz sonuç 202 ve durum adresi.
- Kalıcı simülatör yanıtları; tekrar başvuruda aynı belge/nonce/imza/orijinal.
  Korunan dosya arşivi, commit sonrası metadata yayını ve yeniden başlatma onarımı.
- Geçerli olayları tekrar okuyabilen, işlem ID'siyle çoğalmayı önleyen, checkpoint
  saklayan ve yalnızca izin verilen alanları içeren herkese açık parti görünümü.
- Anlamlı ve hassas veri içermeyen HTTP hataları; özel sorguların evaluate-only
  tutulması; bilinmeyen chaincode komutu için genel submit endpoint'i bulunmaması.
- Sabitlenmiş yerel Java/Maven araçları, Make komutları, CI güncellemesi ve
  [kurulum/API rehberi](../backend/README.md).

## Changed Files

Bu isteğin başlangıcında commit edilmemiş bir backend başlangıcı ile Make/ignore
ekleri vardı. Bunlar incelendi, korundu ve Aşama 5 kapsamında tamamlanıp doğrulandı.

- `backend/pom.xml`, `backend/mvnw`, `backend/README.md`.
- `backend/scripts/toolchain.py`, `backend/scripts/prepare.py`.
- `backend/src/main/java/org/agrochain/Actor.java`, `Api.java`, `ApiError.java`,
  `Application.java`, `AuthFilter.java`, `Crypto.java`, `EvidenceService.java`,
  `FabricLedger.java`, `InstitutionalAdapter.java`, `Json.java`, `Ledger.java`,
  `Projection.java`, `Settings.java`, `Simulators.java`, `Store.java`, `Workflow.java`.
- `backend/src/main/resources/application.properties`.
- `backend/src/test/java/org/agrochain/BackendTest.java`, `GatewayRecoveryTest.java`.
- `backend/test/integration_test.py`.
- `.gitignore`, `Makefile`, `.github/workflows/stage3-chaincode.yml`.
- `README.md`, `ARCHITECTURE.md`, `docs/ARCHITECTURE.md`, `docs/DATA_CONTRACTS.md`,
  `docs/SECURITY_AND_PRIVACY.md`, `docs/IMPLEMENTATION_PLAN.md`,
  `docs/OPEN_QUESTIONS.md`, `docs/PROJECT_REVIEW_ROADMAP_TR.md`, bu rapor.

Chaincode kaynakları, mevcut koleksiyonlar ve akıllı sözleşme sürümü değiştirilmedi.
API erişim anahtarları, kaynak özel anahtarları, SQLite, belgeler, loglar, Maven
önbelleği ve derlenmiş JAR Git dışında tutuluyor.

## Tests Executed

Ortam: Linux x86-64; Temurin 21.0.12.1+1, Maven 3.9.11, Python 3.14.7;
Fabric 2.5.15, Docker 29.8.1 / Compose 5.5.1. Dört kurum peer'i ve üç orderer,
mevcut chaincode 0.2.1 / sequence 5. Backend sürümü 0.3.0.

| Çalıştırılan komut | Sonuç |
| --- | --- |
| `PATH=/usr/bin:/bin:$PATH make network-up backend-prepare` | PASS — veri koruyan ağ açılışı ve geliştirme kimlikleri |
| `PATH=/usr/bin:/bin:$PATH make backend-prerequisites` | PASS — sabitlenmiş yerel araçlar mevcut |
| `bash backend/mvnw -U dependency:tree -Dincludes=com.google.protobuf:protobuf-java,org.hyperledger.fabric:fabric-protos,io.grpc:grpc-protobuf package` | PASS — sürüm uyumu ve paketleme |
| `PATH=/usr/bin:/bin:$PATH make backend-build` | PASS — derleme, birim testleri, çalıştırılabilir JAR |
| `PATH=/usr/bin:/bin:$PATH make backend-integration` | PASS — gerçek HTTP/Gateway/Fabric senaryoları |
| `PATH=/usr/bin:/bin:$PATH make backend-gateway-test` | PASS — yeniden kurulan Gateway ile VALID ve MVCC-invalid commit sorgusu |
| `PATH=/usr/bin:/bin:$PATH make stage5-check` | PASS — son birleşik kabul kontrolü |
| `git diff --check` | PASS — izlenen değişikliklerde biçim kontrolü |

Son birleşik çalıştırmada **15 birim testi**, **10 gerçek HTTP/Fabric testi** ve
**1 ek gerçek Gateway kurtarma testi** geçti. Gateway canlı testi normal Maven
birim aşamasında bilinçli olarak atlanır; ardından ayrı hedefte gerçekten çalışır
(sonuç: 1 test, 0 skipped). Dolayısıyla 26 ayrı test geçti; atlanan testi geçmiş
gibi sayma yapılmadı. Son HTTP paketi 84,925 saniye, Gateway testi 6,333 saniye
sürdü; bunlar pilot gözlemidir, kapasite benchmark'ı değildir.

Canlı kabul: iki senaryoda toplam 14 geçerli yaşam döngüsü komutu; kimlik sahteciliği
reddi; negatif miktar/geçersiz geçiş; tüm özel veri kurum matrisleri; orijinal ve
değiştirilmiş fatura; tekrar/çatışma; simülatör ve Fabric kesintisi; aynı işlemle
iyileşme; commit edilmiş fakat yerel yayını tamamlanmamış kaydın onarımı; kamuya
açık görünümün yeniden oluşturulması ve erişim anahtarı/tuzların loglara sızmaması.

Kurtarma kanıtı iki ayrı düzeyde tutuldu: birim testleri commit öncesi/sonrası
yanıt kaybını modeller; HTTP testi gerçek VALID commit sonrasında SQLite'daki
yerel yayın/checkpoint durumunu geri alarak çökme aralığını açıkça enjekte eder.
Ayrı Gateway testi iki öneriyi aynı başlangıç sürümünde onaylatıp gerçek MVCC
reddi oluşturur ve yeni Gateway bağlantısıyla hem başarılı hem geçersiz sonucu
doğrular. Bunların rastgele zamanlanmış gerçek bir process crash olduğu iddia edilmez.

Son kabul partileri:

- Normal: `BAT-A58EFDD11227B298794071781C23A3BC` / `LOT-90F4901822E26969C1CDDC64341ABAB3`.
- Şüpheli girdili: `BAT-61ECFE748DDF057C4CBC84999FAD0E63` / `LOT-517FA4E0755DA6EFD8724A0EEF0E23AE`.

Yerel test kanıtları `backend/target/surefire-reports/` ve
`network/runtime/backend-acceptance/acceptance-summary.json` içindedir; hassas
çalışma verileri GitHub'a gönderilmez. İlk çalışmada Protobuf uyumsuzluğu ve sandbox
ağ izinleri nedeniyle başarısızlık görüldü; bunlar giderildikten sonra son birleşik
kontrol geçti. Güncellenen uzak CI'ın geçtiği bu raporda iddia edilmiyor.

## Acceptance Criteria

- [x] Simülatör belgeleriyle parti, backend üzerinden tüm ürün akışını tamamlıyor.
- [x] Backend kaynak verisini doğrulayıp gerçek Fabric Gateway'e gönderiyor.
- [x] ÇKS/e-Fatura/U-ETDS ve mimarideki HKS açıkça SIMULATED olarak ayrılıyor.
- [x] Çağıran kurum/rol, backend ve chaincode sınırlarında denetleniyor.
- [x] API hataları uygun 4xx/5xx veya belirsiz commit için 202 dönüyor.
- [x] Aynı işlem tekrarında ikinci etki oluşmuyor; değiştirilmiş istek reddediliyor.
- [x] Özel veri/orijinal belge sorguları ve bütünlük kontrolü çalışıyor.
- [x] Yanıt kaybı, yeniden başlatma ve yerel kayıt yayını kurtarma testleri geçiyor.
- [x] Herkese açık görünüm PDC alanlarını dışlıyor; checkpoint ve replay çalışıyor.

## Assumptions

- Tek bilgisayarlı, tek backend süreçli, sabit domates/100 kg pilotu.
- Kurum simülatörleri aynı uygulamadadır; ayrı HTTP servisleri zorunlu değildir.
- İstemci bir opaque operationId seçer; backend aynı ID'yi doğrulayıp kalıcılaştırır.
  Bu somut HTTP sözleşmesi mimari belgedeki önceki ID tahsis ifadesini netleştirir.
- Sertifika, kaynak anahtarları ve token dosyası yerel pilot operatörünün denetimindedir.
- Kamuya açık görünüm sonunda tutarlıdır; anlık ledger görünümü olduğu iddia edilmez.

## Known Limitations

- Gerçek kurum bağlantısı yoktur; imza, simülatörün iddiasını doğrular, fiziksel
  ürün gerçeğini veya bakanlık onayını kanıtlamaz.
- HTTP yalnızca loopback'te dinler. Geliştirme bearer kimlikleri, tek Retailer
  Gateway ve sıralı yazma kullanılır; üretim oturum/anahtar yönetimi, uzak TLS ve
  yüksek erişilebilirlik kapsam dışıdır.
- SQLite ve dosya izinleri yerel erişimi kısıtlar; host yöneticisine karşı şifreleme
  sağlamaz. Başarısız/belirsiz işlemlerin dosyaları korunur; otomatik süreli temizlik
  yoktur. Büyütmeden önce kontrollü saklama/temizleme politikası gerekir.
- Tam sıfırdan ikinci bilgisayar provası ve uzak CI sonucu bu yerel kabulün yerine
  varsayılmadı. Testler mevcut Stage 4 ağına karşı gerçekten çalıştırıldı.
- Anomali hesaplama, denetçi kararları, QR/aktör arayüzü Aşama 6'ya aittir.
- Zorunlu Aşama 5 kabul ölçütlerinde açık kusur kalmadı.

## Deferred to Future Stages

- Aşama 6: açıklanabilir anomali, denetçi aksiyonları, aktör ve tüketici arayüzleri.
- Aşama 7: daha geniş uçtan uca yük/kurtarma ölçümleri ve temiz kurulum tekrarları.
- Aşama 8: başka bilgisayarda demo provası, sunum, kurulum ve yedek demo paketi.

## Result

**PASS — Aşama 5, yerel pilot kapsamındaki kabul ölçütlerini geçti.**
Kullanıcının isteği doğrultusunda doğrulanmış dosyalar ayrı commit'ler halinde,
her commit'ten sonra ayrı push yapılarak yayımlanacaktır. Aşama 6'ya geçilmedi.
