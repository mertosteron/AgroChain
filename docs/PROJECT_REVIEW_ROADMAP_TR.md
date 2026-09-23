# AgroChain — Teknik durum incelemesi ve yarışmaya hazırlık yol haritası

> **Aşama 8 kapanış güncellemesi — 23 Eylül 2026:** [Yarışma paketi](competition/README.md),
> 12 slayt, çevrimdışı HTML, sabit demo yükleyicisi ve kurulum kılavuzu hazır.
> Yeni kimlik/boş test diskleriyle kurulum ve aynı işlemleri koruyan tekrar yükleme
> doğrulandı. [Kabul raporu](STAGE8_REPORT.md) sınırları ve kanıtları listeler.
> Aşağıdaki aşama planları, ilk incelemenin tarihsel kayıtlarıdır.

> **Aşama 7 kapanış güncellemesi — 23 Eylül 2026:** Birleşik test kapısı, yerel
> gecikme ölçümleri ve iki bağımsız temiz kurulum tamamlandı. Her iki turda normal
> ve şüpheli sonuçlar yeniden üretildi; belge tahrifatı/PDC/tekrar kontrolleri geçti.
> Temiz kurulumda bulunan izin ve önbellek temizleme hataları düzeltildi. Özgün
> ledger ve kimlikler korunarak ağ geri açıldı. [Aşama 7 raporu](STAGE7_REPORT.md)
> ölçümleri ve sınırlamaları içerir. Sıradaki aşama **Aşama 8 yarışma paketidir**.

> **Aşama 6 kapanış güncellemesi — 23 Eylül 2026:** Açıklanabilir tamsayı fiyat
> analizi, zincirde sonuç doğrulaması, özel denetçi incelemesi, aktör ekranları ve
> tüketici QR görünümü tamamlandı. Gerçek ağda 14 HTTP kabul testi ve tarayıcıda
> üreticiden inceleme sonucuna kadar akış doğrulandı. Ayrıntılar
> [Aşama 6 raporunda](STAGE6_REPORT.md). Sıradaki iş Aşama 7 test ve ölçüm paketi;
> sonrasında Aşama 8 yarışma sunumu ve kurulum/demo paketidir.

> **Aşama 5 kapanış güncellemesi — 22 Eylül 2026:** Spring Boot API, gerçek Java
> Gateway bağlantısı, dört imzalı kurum simülatörü, işlem kurtarma ve herkese açık
> güvenli veri görünümü tamamlandı. [Aşama 5 raporu](STAGE5_REPORT.md) canlı test
> kanıtlarını içerir. Sıradaki aşama açıklanabilir anomali ve kullanıcı arayüzüdür.

> **Aşama 4 kapanış güncellemesi — 21 Eylül 2026:** Canlı ürün akışı, özel veri
> koleksiyonları, imzalı belge doğrulaması, tekrar/çakışma kontrolü ve dolu kayıtlarla
> yeniden başlatma doğrulandı. Güncel kanıtlar [Aşama 4 raporunda](STAGE4_REPORT.md).
> Sıradaki iş Aşama 5: Spring Boot ve açıkça simüle edilmiş kurum servisleri.
> Aşağıdaki önceki değerlendirmeler tarihsel kayıttır.

> **21 Eylül 2026 güncellemesi:** Aşama 3'ün daha önce seçilmiş çekirdek iş mantığı +
> güvenli kapalı dağıtım kapsamı için testler, tek komutla doğrulama, CI ve kapanış
> belgeleri tamamlandı. Güncel sonuç ve kanıtlar [Aşama 3 raporunda](STAGE3_REPORT.md).
> Aşağıdaki inceleme 20 Eylül tarihli durumun tarihsel kaydıdır; eski rapor ve eksik
> CI bulguları güncel kapanış çalışmasıyla ele alındı. Doğrulanmış canlı ürün akışı
> ve PDC kabulü hâlâ Aşama 4'e aittir; bunlar tamamlanmış sayılmaz.

İnceleme tarihi: **20 Eylül 2026**. Kapsam: bu bilgisayardaki çalışma ağacı, mimari belgeler, kaynak kodu, mevcut otomasyon ve bu incelemede çalıştırılan testler. Yarışmanın harici şartnamesi veya puanlama sistemi incelenmedi; değerlendirme deponun `AGENTS.md` kabul ölçütlerine göredir.

**Karar: Aşama 1 ve Aşama 2 yerel pilot kapsamlarında onaylanabilir. Aşama 3'ün iş mantığı ve kısıtlı ağ dağıtımı çalışıyor, ancak tam kabulü tamamlanmış değil. Aşama 4–8'in ana teslimatları eksik. Proje henüz jüriye sunulabilecek uçtan uca son ürün değil; doğrulanmış bir Fabric altyapısı ve test edilmiş bir chaincode çekirdeği.**

İlk iki aşamanın onayı, ürünün yüzde 25'inin tamamlandığı anlamına gelmez. Aşamaların iş yükü eşit değil; kullanıcı deneyimi ve uçtan uca güvenlik kanıtlarının çoğu hâlâ ileride.

## 1. Aşama durum tablosu

| Aşama | Mevcut durum | Onay | Eksik kabul kanıtı / sonraki iş |
| --- | --- | --- | --- |
| **1 — Kapsam ve teknik sözleşme** | Roller, yaşam döngüsü, veri sınıfları, simülatör sınırları, formüller ve senaryolar tanımlı. Belge doğrulayıcısı yeniden geçti. | **PASS — tasarım kapsamı** | Sonraki aşamalarda oluşmuş Aşama 3 kabul/rapor tutarsızlığı giderilmeli. Bu onay çalışma zamanı güvenlik kanıtı değildir. |
| **2 — Hyperledger Fabric ağı** | Dört kurum/peer, üç Raft orderer, kanal, TLS, dağıtım araçları, sorgular ve yeniden başlatma doğrulandı. | **PASS — yerel pilot** | Temiz ortam kurulumu önceki raporda kayıtlı; bu incelemede sıfırlama yapılmadı. İkinci bilgisayar provası henüz kanıtlı değil. |
| **3 — Chaincode** | Yedi iş komutu için domain, doğrulama, yetkilendirme, sorgular, testler ve v0.1.2 dağıtımı mevcut. Canlı iş yazmaları kapalı. | **KISMİ — tam aşama PASS değil** | Güncel kapanış raporu; kabul kapsamının tutarlı hale gelmesi; gerçek ürün oluşturma/devir ve dolu defter senaryoları. |
| **4 — Gizlilik ve veri doğrulama** | Üç PDC yapılandırması ve kriptografik sözleşme mevcut. Çalışan özel veri/kanıt akışı yok. | **TAMAMLANMADI** | Rol taşıyan kimlikler, imza/doğrulama, transient veri, atomik PDC işlemleri, yetkili/yetkisiz erişim ve tahrifat testleri. |
| **5 — Backend ve kurum simülatörleri** | API ve adapter sözleşmeleri belgelenmiş. Spring Boot uygulaması ve simülatör implementasyonları yok. | **TAMAMLANMADI** | Gateway, kurum bazlı imzalayıcılar, ÇKS/e-Fatura/HKS/U-ETDS, API, işlem kurtarma ve belge saklama. |
| **6 — Anomali motoru ve arayüz** | Formül, normal/şüpheli örnekler ve inceleme süreci tasarlanmış. Çalışan motor ve UI yok. | **TAMAMLANMADI** | Aktör/denetçi/tüketici ekranları, özel anomali verisi, açıklanabilir sonuç, QR ve insan incelemesi. |
| **7 — Test ve ölçüm** | Önceki aşamaların testleri var. Tam ürün test paketi ve ölçüm raporu yok. | **TAMAMLANMADI** | Gerçek PDC, backend ve UI uçtan uca testleri, tekrar üretilebilir demo, gecikme ölçümleri. |
| **8 — Yarışma paketi** | Ağ ve chaincode kılavuzları var. Nihai ürün kurulum/demo paketi yok. | **TAMAMLANMADI** | Demo verisi yükleme, 10–12 slayt, 3 dakikalık akış, çevrimdışı yedek ve başka bilgisayarda prova. |

## 2. Mevcut uygulamanın teknik değerlendirmesi

### Sağlam temeller

- **Pilot kapsamı açık:** tek ürün sınıfı, sabit rota, tek bütün parti, dört kuruluş. Bölme/birleştirme, kayıp, iade ve kısmi teslimat kapsam dışında. Bunların sessizce düzeltilmesi yerine reddedilmesi tasarlanmış.
- **Mülkiyet ve taşıma sorumluluğu ayrılmış:** lojistik firması taşıma sırasında ürünün sahibi olmuyor; perakendecinin teslim kabulü sahipliği değiştiriyor.
- **Yaşam döngüsü açık ve kodla uyumlu:** `CREATED → PICKUP_PENDING → IN_TRANSPORT → DELIVERY_PENDING → RECEIVED → RETAIL_REPORTED`. Son durum fiyat bildirimidir; gerçekleşmiş satış kanıtı değildir.
- **İş mantığı test edilmiş:** kurum/rol denetimi, miktar ve para sınırları, sürüm çatışmaları, tekrar gönderim, belge referansı tekrar kullanımı, geçersiz durum geçişleri, sorgu sınırları ve paylaşılan kayıt/olaylara özel veri sızmaması için testler var.
- **Sayısal model uygun:** miktar gram, ticari tutarlar kuruş üzerinden tamsayı. Kayan noktalı finans hesabı kullanılmıyor.
- **Güvensiz kısa yol kapalı:** dağıtılan contract doğrulanmamış belgeyi kabul ederek ürün oluşturmuyor. Test kanıt sağlayıcısı yalnız test dosyalarında bulunuyor.
- **Ağ işlemleri otomatik:** mevcut Make hedefleri kurulum, doğrulama, chaincode dağıtımı ve veri koruyan kapatma/başlatmayı kapsıyor. Yeni bir ağ çatısı kurmaya ihtiyaç görünmüyor.

Bu güçlü taraflar korunmalı; tamamlanma için çalışan çekirdeği yeniden yazmak gerekmiyor.

### Kritik bulgular ve öncelikler

| Öncelik | Bulgu ve kanıt | Etkisi | Gerekli iş |
| --- | --- | --- | --- |
| **P0 — aşama geçişi** | `docs/STAGE3_REPORT.md:134` testlerin yapılmadığını, `:182` implementasyonun olmadığını söylüyor. Kod, README ve bu incelemenin testleri bunun aksini gösteriyor. | Ekip hangi kapsamın bittiğini güvenilir biçimde göremiyor. | Eski raporu güncel implementasyon ve test kanıtlarıyla uzlaştır; tarihi engeli güncel sonuçtan ayır. |
| **P0 — kabul bağımlılığı** | `docs/IMPLEMENTATION_PLAN.md` Aşama 3 canlı yazmalarını Aşama 4'e bırakıyor, Aşama 4 için de Aşama 3 PASS istiyor. `docs/OPEN_QUESTIONS.md:5` daraltılmış kapsam kararının zaten alındığını belirtiyor. | Tam canlı Aşama 3 kabulü ile sıralı aşama kuralı arasında döngü oluşuyor. | Mevcut kararın hangi daraltılmış kapıyı kapattığını ve hangi kanıtların Aşama 4'e taşındığını tek kabul tablosunda belirt. |
| **P0 — ürün akışı** | `contract.go:81` iş komutlarını `closedEvidence` üzerinden reddediyor; `models.go:150` sonucu `EVIDENCE_VERIFICATION_UNAVAILABLE`. Canlı Health yanıtı `writesEnabled=false`. | Şu an gerçek ağda ürün oluşturma/devir demosu yapılamaz. | Aşama 4'te doğrulayıcı ve PDC işlemlerini tamamlayarak güvenli yolu aç. Sadece bayrağı değiştirme veya test sağlayıcısını ağa bağlama. |
| **P0 — gizlilik kanıtı** | Collection tanımları var; uygulamada gerçek özel veri okuma/yazma yolu yok. | PDC dosyasının varlığı ticari gizliliği ispatlamaz. | Gerçek üye ve üye olmayan kimlik/peer ile erişim testleri yap. |
| **P1 — kimlik** | `validation.go:273` rol boşsa MSP bazlı denetime izin veriyor. Bu sınırlama belgelenmiş. | Mevcut kapalı yazma durumunda sınırlı; canlı yazmalar açılırken yeterli olmayacak. | Gerekli iş rolü niteliklerini üret, zorunlu kıl; eksik/yanlış rolü doğrudan chaincode çağrısında reddet. |
| **P1 — tekrar üretilebilirlik** | İnceleme başında `chaincode/` Git tarafından izlenmiyordu; `git ls-files chaincode` sıfır dosya döndürdü. Başka yerel değişiklikler de vardı. | Bu makinedeki çalışan durumun mevcut commit'ten klonlanarak elde edilebildiği söylenemez. | Kaynakları ve güncel raporları gözden geçirilmiş bir sürümde kaydet; üretilmiş anahtarları dahil etme. Bu incelemede commit/push yapılmadı. |
| **P1 — CI** | Mevcut workflow yalnız ağ yollarıyla tetikleniyor; chaincode Go testlerini, dağıtımını ve entegrasyonunu çalıştırmıyor. | Yalnız chaincode değişikliği otomatik olarak doğrulanmayabilir. | Mevcut CI'ı chaincode testleri ve uygun entegrasyon adımlarıyla genişlet. GitHub üzerinde çalıştığı ayrıca kanıtlanmalı. |
| **P1 — çalışma ortamı** | Varsayılan `python3` 3.9.23; sistem Python'u 3.14.7. İlk ağ başlatma bu yüzden reddedildi. | Belgelenmiş Python ≥3.10 önkoşulu yanlış PATH yüzünden sağlanmayabilir. | Kurulum/preflight'ta kullanılan Python yolunu ve sürümünü görünür kıl. |

**Güven sınırı:** Dört MSP ve üç orderer aynı bilgisayarda çalışıyor. Bu, kuruluşlar arası mantıksal ayrımı gösterir; bağımsız kurum işletimi veya tek bilgisayar arızasına dayanıklılık ispatı değildir. Mevcut Retailer + Regulator onay politikası denetçi işlemlerinin perakendeci erişilebilirliğine bağımlılığını da beraberinde getirir. Sunumda açıklanmalı; pilotu bitirmeden çok makineli altyapıya genişlemek gerekmiyor.

## 3. Bu incelemede çalıştırılan doğrulamalar

| Kontrol | Sonuç | Kanıtın sınırı |
| --- | --- | --- |
| `docs/STAGE1_REPORT.md` içindeki JavaScript doğrulayıcısının Node ile çalıştırılması | **PASS:** 9 belge, 9 JSON örneği, 42 yerel bağlantı; imza/hash/tahrifat ve 7 hesaplama sınır örneği | Örnek sözleşmeleri doğrular; çalışan anomali motoru veya kurumsal bağlantı değildir. |
| `make test` | **PASS:** 24 test | Ağ yapılandırması, güvenli temizlik, paketleme ve kaynak güvenlik kontrolleri. |
| `make chaincode-test` | **PASS** | Biçim kontrolü, modül doğrulama, `go vet`, race denetimli test ve çalıştırılabilir dosya üretimi. |
| Chaincode test kapsamı | **İş mantığı paketi %89,2; tüm paketler %87,3** | Sunucu başlangıç kodu %0. Kapsam oranı tek başına güvenlik veya E2E doğruluk sertifikası değildir. |
| `PATH=/usr/bin:/bin:$PATH make network-up` | **PASS**, uygun Docker erişimiyle | Mevcut kimlikler/defterler korundu; temiz kurulum yapılmadı. |
| Aynı PATH ile `make verify` | **PASS** | 7 ağ düğümü, TLS/mTLS ve negatif TLS kontrolleri, 4 MSP/anchor, 3 Raft consenter, kurum sorguları ve peer yeniden başlatma. |
| Aynı PATH ile `make chaincode-integration` | **PASS:** 6 test | Dört kimlikle sorgu; kapalı iş yazmaları; yanlış kurum/girdi/fiyat alanı reddi; çift onaylı Health commit'i; tek onayın reddi. |
| Aynı PATH ile `make chaincode-restart-check` | **PASS** | Tam kapatma/başlatma sonrası dört defterin yüksekliği/hash'i, contract ve sorgu sonucu korundu. |
| `git diff --check` | **PASS**, rapor eklenmeden önce | İncelenen mevcut değişikliklerde boşluk biçimi hatası yoktu. İşlevsel kanıt değildir. |

Canlı entegrasyon dört geçerli Health işlemini 27–30 numaralı bloklara ekledi. Eksik onaylı işlem geçersiz olarak blokta yer aldı; geçerli iş durumu oluşturmadı. Son yeniden başlatma kanıtında dört peer'ın blok yüksekliği **32**, contract sürümü **0.1.2**, `writesEnabled=false` ve test partisinin varlığı `false` idi. Bunlar gerçek Fabric işlem/kalıcılık kanıtlarıdır; başarılı ürün yaşam döngüsü kanıtı değildir.

İlk ağ başlatma denemesi eski Python nedeniyle, sonraki sandbox denemesi Docker erişimi nedeniyle başarısız oldu. Sistem Python'u ve izinli Docker erişimiyle tekrar çalıştırıldığında geçti. Belge doğrulayıcısını ilk ayıklama denemem de yanlış kod bloğu ayrıştırması nedeniyle hata verdi; tam blok doğru biçimde çalıştırıldığında geçti. Bu hazırlık hataları ürün testi başarısızlığı olarak yorumlanmadı.

Ortam: Linux `7.2.6-arch2-1` / x86-64; Docker `29.8.0`; Compose `5.5.1`; Fabric `2.5.15`; proje Go aracı `1.27.1`; canlı ağ komutlarında sistem Python'u `3.14.7`.

**Bu oturumda yapılmayanlar:** temiz defter/kimlik sıfırlaması ve `make smoke`, yeni chaincode release dağıtımı, bağımsız bilgisayarda kurulum, GitHub CI koşusu, başarılı canlı ürün oluşturma/devir, gerçek PDC erişimi, backend/UI E2E ve performans ölçümü. Önceki Aşama 2 temiz kurulum kanıtı `docs/STAGE2_REPORT.md` içindeki tarihsel kayda dayanıyor.

## 4. Açık yol haritası

Uygulama sırası değişmemeli: **Aşama 3 kapanışını netleştir → Aşama 4 → Aşama 5 → Aşama 6 → Aşama 7 → Aşama 8.** Her aşamanın kendi testleri o aşamada yazılmalı; güvenlik testleri Aşama 7'ye bırakılmamalı.

### İlk iş — Aşama 3'ün kapanışını tutarlı hale getir

**Teslimat:** Güncel `STAGE3_REPORT.md`, tutarlı kabul tablosu, sürüm kontrolüne alınmış mevcut kaynak ve chaincode'u kapsayan otomasyon.

1. Rapordaki “implementasyon yok”, “test çalışmadı” ve “kullanıcı yanıtı bekleniyor” ifadelerini güncel kanıtlarla düzelt.
2. `OPEN_QUESTIONS.md` içinde zaten kayıtlı daraltılmış kapsam kararını diğer belgelerle eşleştir. Kullanıcıdan geçmişte alınmış aynı kararı tekrar istemek gerekmiyor.
3. İki kanıt grubunu ayır: **çekirdek/domain + güvenli kapalı dağıtım** ve **doğrulanmış belgelerle canlı ürün yaşam döngüsü**. Birincisi bu incelemede olumlu doğrulandı; ikincisi Aşama 4 bağımlılığıdır.
4. Orijinal tam Aşama 3 kabulünü geçti diye işaretleme. Mevcut sıralı plan, daraltılmış kapı ve ertelenen canlı ölçütlerin yeri açıkça uzlaştırılmadan Aşama 4 uygulamasına geçme.
5. Mevcut kaynakları, testleri ve raporları sürümle; temiz checkout üzerinde en azından yerel test/derleme komutlarının çalışmasını doğrula.

**Çıkış ölçütü:** Ekip tek bir kabul tablosundan neyin geçtiğini, neyin Aşama 4'e bağlı olduğunu görebiliyor; “Aşama 3 bitti” ifadesi canlı ürün akışı çalışıyor anlamına gelmiyor. Bu rapor mevcut normatif belgeleri kendiliğinden değiştirmez ve aşama kuralını kaldırmaz.

### Aşama 4 — Gizlilik ve doğrulanmış canlı işlemler

**Amaç:** Güvenli biçimde ilk gerçek ürün yaşam döngüsünü açmak. Projenin şu anki en kritik teknik işi budur.

**İş sırası:**

1. Producer/carrier/retailer ve düzenleyici rollerine uygun kimlikleri üret; zorunlu rol kontrollerini uygula.
2. Defterde güvenilen simülatör kaynak anahtarlarını ve belge türü eşleşmelerini tanımla.
3. Canonical JSON, Ed25519 imza, 32 bayt gizli tuzlu commitment, batch/operation/source bağları ve belge/nonce tekrar kullanımını doğrula. Belgelenmiş vektörü Go ve Java'da aynı sonuçla üret.
4. `tradePrivate`, `freightPrivate`, `retailAuditPrivate` okuma/yazma yollarını uygula. Özel değerleri transient map üzerinden yalnız uygun peer/endorser'lara gönder.
5. Paylaşılan durum, PDC kayıtları, belge kullanım indeksleri ve işlem makbuzlarını tek atomik işlemde üret. Eksik özel veriyi sıfır fiyat olarak yorumlama.
6. Belgenin özgün byte'larını kaydedilen kanıta göre doğrula; değiştirilmiş belge için yeni hash üretip “geçerli” deme.
7. Domain motorunu yalnız bu doğrulamadan geçen üretim yoluna bağla ve dolu gerçek defter üzerinde yaşam döngüsünü çalıştır.

**Çıkış ölçütleri:** Yetkili taraf kendi özel verisini okur; yetkisiz istemci ve üye olmayan peer okuyamaz. Paylaşılan izlenebilirlik çalışır. Özgün belge geçer; değişmiş belge/imza/bağ/nonce reddedilir. Blok, olay, hata ve loglarda özel fiyat/tuz sızıntısı yoktur. Gerçek ürün oluşturma, taşıma, teslim ve fiyat bildirimi tamamlanır; tekrar gönderim ve yarışan kabul işlemleri ikinci geçerli kayıt oluşturmaz. Yeniden başlatma sonrası bu kez **dolu ürün ve özel kayıtlar** doğrulanır.

**Sınır:** Buradaki imzalı test belgeleri ÇKS/e-Fatura/HKS/U-ETDS uygulamalarının tamamlandığı anlamına gelmez; adapter ve senaryo servisleri Aşama 5'tir.

### Aşama 5 — Spring Boot ve dört kurum simülatörü

**Amaç:** Komut satırında kanıtlanan akışı gerçek uygulama API'sine taşımak.

1. Tek Spring Boot uygulamasında kimlik doğrulama, rol denetimi ve kurum başına Fabric imzalayıcı eşlemesini kur.
2. Fabric Gateway bağlantısını, uygun endorser seçimini ve `VALID` commit beklemeyi uygula. Tüm aktörleri tek düzenleyici kimliğiyle imzalama.
3. Dört adapter'ı tamamla: **ÇKS**, **e-Fatura** (alış + taşıma faturası), **HKS**, **U-ETDS**. HKS depodaki kabul edilmiş mimarinin parçasıdır; üç simülatör yeterli sayılmamalı.
4. Her kaynak çıktısında `sourceMode=SIMULATED`, kontrollü kaynak anahtarı ve tekrar istekte aynı belge/işlem bağı korunsun.
5. Ürün, teslim, taşıma maliyeti, fiyat raporu, belge doğrulama, özel sorgu ve işlem durumu API'lerini uygula.
6. Belgelenmiş SQLite işlem günlüğü, kontrollü belge saklama, commit sonucu belirsiz isteğin uzlaştırılması ve olay okuma checkpoint'ini tamamla.
7. Tüketiciye sunulacak izinli alan listesini ayrı projection olarak üret; özel kayıtların genel JSON çıktısını kullanma.

**Çıkış ölçütü:** Üretici → taşıyıcı → perakendeci rotası API üzerinden gerçek Fabric commit'leriyle tamamlanır. Simülatör kesintisi, eksik belge, yanlış rol, tekrar istek, commit sonrası timeout ve servis yeniden başlatması anlamlı hata/pending durumlarıyla yönetilir. Chaincode hiçbir dış API çağırmaz.

### Aşama 6 — Açıklanabilir anomali ve jüri ekranları

**Amaç:** Jürinin sistemi kendi gözleriyle anlayıp deneyebilmesi.

1. Mevcut sözleşmedeki tam sayılı hesaplamayı uygula; alış ve bildirilen perakende birim fiyatı aynı para birimi/vergi temelinde karşılaştır.
2. Yapılandırılabilir pilot eşiğini sürümlü, değişmez politika kaydıyla ilişkilendir. Eşik **kesin olarak aşıldığında** sinyal üret; yüzde 50'nin kendisi sinyal değildir.
3. Backend sonucunu chaincode'da özel kayıtlardan yeniden hesaplayarak doğrula. Hesaplanmamış sonuç `EVALUATION_PENDING` kalsın.
4. Aktör ekranı: parti oluşturma, teslim teklif/kabulü, belge durumu, işlem makbuzu ve commit hataları.
5. Denetçi ekranı: alış/fiyat/taşıma bağlamı, formül, eşik, gerekçe, `OPEN → IN_REVIEW → RESOLVED` inceleme akışı.
6. Tüketici QR ekranı: uygun köken ve kabul edilmiş taşıma geçmişi. Ticari fiyatlar, özel belgeler ve denetim sınıflandırması görünmesin.
7. “Simüle kurum verisi” ve “gerçek blockchain işlemi” ayrımını ekranda açık göster.

**Çıkış ölçütü:** Jüri normal ve şüpheli partiyi karşılaştırır, sonucu açıklayabilir, yetkili incelemeyi yapabilir. QR'dan özel veri sızmaz. Yüzde 50, hemen üstü, fiyat düşüşü, sıfır alış ve eksik kanıt testleri geçer. Sonuç yasal ihlal veya kesin dolandırıcılık hükmü olarak sunulmaz.

**Tasarım tercihi:** Küçük, anlaşılır üç görünüm yeterli. Gösterişli dashboard, ayrı mikroservisler veya ML eğitim hattı kabul için gerekli değil.

### Aşama 7 — Uçtan uca kanıt ve ölçüm

1. Mevcut Make arayüzüne tek test ve deterministik demo verisi yükleme akışı ekle.
2. Normal, şüpheli, tahrifat ve tekrar senaryolarını API/gerçek peer/UI seviyesinde birleştir.
3. Gerçek PDC erişim reddi, eşzamanlı transfer kabulü, commit belirsizliği, servis restart ve projection kurtarmasını doğrula.
4. Kontrollü temiz reset ve seed işlemini iki kez tekrar et; aynı iş sonuçlarını al. Rastgele kimlik/tuz nedeniyle blok hash'lerinin aynı olmasını bekleme.
5. Donanım, sürümler, payload, örnek sayısı, sıcak/soğuk başlangıç ve hata sayısını kaydet. Gönderimden `VALID` commit'e, paylaşılan/özel sorguya ve tam demo rotasına ait süreleri ölç; medyan/p95 hesapla.
6. Komut, senaryo, beklenen sonuç, gerçekleşen sonuç, işlem kimliği ve sınırlama içeren bir test raporu üret.

**Çıkış ölçütü:** Yarışma açısından kritik otomatik testler geçer ve demo yeniden üretilebilir. Ölçümler yerel pilot gözlemidir; ulusal ölçek kapasite iddiası değildir.

### Aşama 8 — Son yarışma ürünü

1. Tek başlangıç kılavuzunu finalize et: desteklenen ortam, kaynak sürümü, önkoşullar, bağımlılıkların önceden indirilmesi, başlatma, seed, test, stop/reset ayrımı.
2. İki ana parti ve bozulmuş belge/tekrar örneğini deterministik yükle; yalnız elle girilmiş yerel veriye dayanma.
3. 10–12 slayt hazırla: problem, dar pilot kapsamı, aktörler, mimari, gerçek Fabric kanıtı, PDC, simülatör sınırı, anomali açıklaması, test/ölçüm, demo, sınırlamalar/sonraki işler.
4. Canlı demo için preflight, tek operatör akışı ve geri dönüş planı hazırla.
5. Aynı çalışan sürümden yerel video, ekran görüntüsü ve doğrulanmış işlem/test kanıtlarını kaydet. Yedek gösterimi canlıymış gibi sunma.
6. Başka bir takım üyesi başka desteklenen bilgisayarda yalnız belgeleri izleyerek kurup demoyu çalıştırsın.

**Çıkış ölçütü:** Ekipten biri kurulumdan normal/şüpheli sonuçlara ve gizlilik kanıtına kadar tüm akışı tekrarlayabiliyor; sunum anında dış kurum internet servislerine ihtiyaç duyulmuyor.

## 5. Nihai demoda gösterilecek somut sonuçlar

| Senaryo | Sabit iş verisi | Jürinin görmesi gereken |
| --- | --- | --- |
| **Normal: BAT-NORMAL01** | 100 kg domates; alış 20 TL/kg; bildirilen fiyat 28 TL/kg; taşıma 200 TL toplam; pilot eşik %50 | %40 artış, `NO_SIGNAL`, gerçek işlem geçmişi ve tüketici köken görünümü. |
| **Şüpheli: BAT-SUSPECT1** | Aynı miktar/alış/taşıma; bildirilen fiyat 37 TL/kg | %85 artış, `REVIEW_REQUIRED`, yetkili inceleme ve açıklama. |
| **Bütünlük/tekrar: BAT-TAMPER01** | Geçerli belge, sonra değiştirilmiş byte veya imza; ardından tekrar gönderim | Özgün belge geçer; değişmiş belge reddedilir; tekrar ikinci ürün/devri oluşturmaz. |

Artış hesabı `(bildirilen fiyat − alış fiyatı) / alış fiyatı × 100` anlamındadır; implementasyonda tamsayı ve tam karşılaştırma kullanılmalı. Taşıma maliyeti bağlamdır; bu formülde düşülmez. Sonuç net kâr veya yasal fiyat sınırı değildir.

Önerilen üç dakikalık sunum akışı:

| Süre | Gösterim |
| --- | --- |
| 0:00–0:25 | Sorun, dört aktör, gerçek blockchain / simüle kurum ayrımı. |
| 0:25–1:00 | Önceden yüklenmiş normal partinin yolculuğu, makbuzu ve tüketici QR görünümü. |
| 1:00–1:40 | Şüpheli partide 20 → 37 TL, %85 artış ve %50 pilot eşiği; yetkili canlı inceleme işlemi. |
| 1:40–2:15 | Yetkisiz özel fiyat erişiminin reddi ve değiştirilmiş belgenin doğrulamadan kalması. |
| 2:15–2:40 | Tekrar gönderimin ikinci işlem oluşturmadığının gösterilmesi. |
| 2:40–3:00 | Test kanıtı, pilot sınırlamaları ve sistemin neyi ispatladığı. |

Bu bir gelecek demo planıdır; bugün çalışır durumda olduğu iddia edilmiyor.

## 6. Kapsam ve planlama kararları

**Şimdi korunacaklar:** Hyperledger Fabric, mevcut Go domain'i, Spring Boot tercihi, tek kanal/üç PDC, dört simülatör, bütün parti modeli, iki fiyat senaryosu, kullanıcı dışındaki yetkilendirme ve açıklanabilir kurallar.

**Ertelenecekler:** Kubernetes, Kafka, yeni mikroservis katmanları, ML, bulut yayını, token/ödeme, gerçek devlet entegrasyonu, ürün sayısını artırma, çoklu rakip perakendeci modeli, bölme/birleştirme ve iade. Bunlar mevcut pilotun eksik kabul ölçütlerini kapatmıyor.

**Takvim yaklaşımı:** Teslim tarihi, ekip kapasitesi ve hedef demo bilgisayarı bilinmediğinden kesin gün/hafta taahhüdü verilmedi. İlerleme aşama kapılarıyla izlenmeli. En büyük belirsizlik Aşama 4'teki gerçek özel veri ve kimlik akışları; Aşama 5'teki işlem kurtarmasıdır. UI paralelinde sahte backend oluşturarak hız kazanılmış sayılmamalı.

## 7. İnceleme tamamlanma raporu

### Stage

Aşama 1–8 durum incelemesi ve yol haritası. Yeni bir aşama implementasyonu yapılmadı.

### Completed

- Repo, mimari, kabul sözleşmeleri, ağ, domain, testler ve CI incelendi.
- İlk iki aşama kapsamları ve kanıt sınırlarıyla onaylandı.
- Aşama 3'ün çalışan parçaları, kapanış sorunu ve canlı ürün akışı eksiği ayrıştırıldı.
- Aşama 4–8 için sıralı teslimat ve geçiş ölçütleri çıkarıldı.
- Mevcut testler ve canlı ağ/entegrasyon/kalıcılık kontrolleri çalıştırıldı.

### Changed Files

- `docs/PROJECT_REVIEW_ROADMAP_TR.md` — bu yeni rapor.
- Kaynak kodu ve önceki aşama raporları değiştirilmedi; mevcut kullanıcı değişiklikleri korundu.
- Testler tarafından üretilen/yenilenen dosyalar: `chaincode/agrochain/build/coverage.out`, `chaincode/agrochain/build/agrochain`, ağ doğrulama çıktıları ve `network/runtime/chaincode/` içindeki entegrasyon/yeniden başlatma kanıtları. Bunlar sürüm kontrolü dışında tutulan çalışma çıktılarıdır.
- Canlı testler deftere Health ve eksik onay testi kayıtlarını ekledi; mevcut veriler sıfırlanmadı. Son durumda ağ yeniden çalışır halde bırakıldı.

### Tests Executed

Yukarıdaki doğrulama tablosunda belirtilen belge kontrolü, 24 yerel test, chaincode derleme/testleri, canlı ağ doğrulaması, 6 entegrasyon testi ve yeniden başlatma kontrolü **PASS**. Testlerin kanıtlamadığı tam ürün davranışları ayrıca belirtildi.

### Acceptance Criteria

- [x] Tamamlanmış aşamalar kanıtlarıyla ayrıştırıldı.
- [x] Eksik aşamalar ve önkoşullar listelendi.
- [x] Her kalan aşama için somut teslimat ve kabul ölçütü yazıldı.
- [x] Çalıştırılan ve çalıştırılmayan kontroller ayrıldı.
- [ ] Aşama 3'ün orijinal tam canlı yaşam döngüsü kabulü.
- [ ] Aşama 4–8'in son ürün kabul ölçütleri.

### Assumptions

- Kaynak gerçekliği bu yerel çalışma ağacıdır; uzak depoyla aynı olduğu varsayılmadı.
- Kurum API erişimi yok; dört kaynak simülatör olarak kalacak.
- Tek makine, sentetik veri, tek bütün parti ve mevcut fiyat kuralı korunacak.
- Daha önce belgelenmiş daraltılmış Aşama 3 kararı yeni bir onay bekliyormuş gibi ele alınmayacak; rapor ve bağımlılık tablosunda tutarlı hale getirilecek.

### Known Limitations

- Önceki Aşama 3 raporu güncel kodla çelişiyor; bu inceleme raporu onu sessizce değiştirmedi.
- Temiz kurulum, başka bilgisayar, hosted CI ve tam ürün E2E bu oturumda doğrulanmadı.
- Kod incelemesi ve geçen testler bağımsız kapsamlı güvenlik denetimi yerine geçmez.

### Deferred to Future Stages

Gizlilik/kanıt implementasyonu, backend/simülatörler, anomali/UI, uçtan uca ölçüm ve yarışma paketi; Bölüm 4'teki sırayla.

### Result

**İnceleme ve yol haritası: PASS. Aşama 1: PASS. Aşama 2: PASS — yerel pilot. Aşama 3 tam kabulü: FAIL / tamamlanmamış. Nihai yarışma ürünü: FAIL / henüz tamamlanmamış.**

Buradaki FAIL, geçen testlerin başarısız olduğu anlamına gelmez; zorunlu son ürün kabul ölçütlerinin henüz tamamlanmadığını belirtir.
