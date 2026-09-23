# Aşama 8 tamamlanma raporu — 23 Eylül 2026

## Stage

Stage 8 — Son yarışma paketi. Önkoşul, mevcut [Aşama 7 PASS raporu](STAGE7_REPORT.md).
Bu aşamada domain, PDC politikası veya anomali kuralı değiştirilmedi.

## Completed

- Yeni kurulum, yeniden açılış, güvenli durdurma, kimlik kullanımı ve hata giderme
  kılavuzu hazır. `make demo-setup` mevcut ağı sıfırlamadan yeni kurulum yapar.
- Sabit iki sentetik parti `make demo-seed` ile gerçek HTTP → Gateway → Fabric
  üzerinden yüklenir. `demo-check` %40/%85, %50 eşik, 14 COMMITTED işlem, tüketici
  alan sınırı ve yetkisiz analiz erişimini doğrular. Tekrar yükleme aynı işlemleri kullanır.
- `demo-ready` dört peer'in yüklü chaincode'u ve diğer kuruluşları keşfettiğini
  kontrol eder. Bu, temiz provada saptanan özel veri yayılımı başlangıç hatasını giderir.
- Kısa teknik açıklama, Mermaid mimari diyagramı, düzenlenebilir mimari/tablolar
  ve fiyat grafiği içeren **12 slaytlık PPTX** hazır. Konuşmacı notları kaynakları belirtir.
- Üç dakikalık canlı akış, teknik soru cevap, kesinti geçişi ve video çekim planı hazır.
- Tek dosyalı çevrimdışı HTML; tüm slaytlar ve sabit sentetik partilerin gerçek
  arayüzden alınmış iki ekran görüntüsü gömülü. Canlı sorgu olmadığı açıkça belirtilir.
- Kaynak ve sunumları checksum/manifest ile paketleyen `make competition-package`
  hazır. Araç önbelleği, ledger diskleri, özel anahtarlar ve tokenlar pakete girmez.
- Temiz kaynak kopyası, yeni kimlikler ve boş test diskleriyle kurulum rehberi
  uygulandı; iki seed aynı işlem kimliklerini korudu. Özgün ağ/kimlikler geri getirildi.

## Changed Files

Aşama 7'den kalan çalışma ağacı değişiklikleri korundu. Aşağıdakiler Aşama 8 kapsamıdır:

```text
.gitignore
ARCHITECTURE.md
Makefile
README.md
network/README.md
docs/IMPLEMENTATION_PLAN.md
docs/PROJECT_REVIEW_ROADMAP_TR.md
docs/STAGE8_REPORT.md
docs/competition/README.md
docs/competition/INSTALLATION.md
docs/competition/DEMO_3MIN.md
docs/competition/TECHNICAL.md
docs/competition/OFFLINE.md
docs/competition/slides.json
docs/competition/AgroChain.pptx
docs/competition/offline.html
docs/competition/assets/comparison.png
docs/competition/assets/consumer.png
scripts/demo.py
scripts/demo-ready.sh
scripts/demo-setup.sh
scripts/test_demo.py
scripts/rehearse-demo.py
scripts/capture-demo.cjs
scripts/build-offline.py
scripts/build-presentation.mjs
scripts/test-offline.cjs
scripts/package-demo.py
scripts/test_package.py
```

Üretilen dağıtım dosyaları Git dışında `dist/AgroChain-competition.tar.gz` ve
`dist/AgroChain-competition.sha256`. Bu istek kapsamında commit/push yapılmadı.

## Tests Executed

| Komut / kontrol | Sonuç |
| --- | --- |
| `PATH=/usr/bin:/bin:$PATH make stage8-unit` | PASS — 8 test: sabit/benzersiz kimlik, aynı işlemle yeniden deneme, yetki hatası, uzak token gönderimini engelleme, paket veri/dizin sınırları ve PEM blok denetimi |
| `make demo-seed demo-check` | PASS — özgün gerçek ağda normal/şüpheli veri ve tekrar yükleme |
| `PATH=/usr/bin:/bin:$PATH make stage8-rehearse` | PASS — yeni kaynak kopyası, kimlikler, boş ledger ve backend günlüğü; kurulum 68,679 sn, ilk seed 35,303 sn, ikinci seed 0,164 sn, kontrol 0,114 sn |
| Temiz kurulumun `privacy-integration` adımı | PASS — 9 gerçek PDC, yetki, tahrifat, tekrar, MVCC ve yeniden başlatma testi; 33,071 sn |
| Temiz kurulumun `backend-build` adımı | PASS — 20 Java birim testi; bu komutta etkinleştirilmeyen 2 canlı Gateway testi açıkça SKIP |
| `make demo-ready` ve geri yüklenen ağda `make demo-check` | PASS — dört kuruluş keşfi, özgün kimlik/disklerin korunması ve eski sabit partilere erişim |
| `node scripts/capture-demo.cjs` (Playwright/Chromium) | PASS — sabit sentetik karşılaştırma ve tüketici ekranı; JavaScript hatası yok |
| `node scripts/test-offline.cjs` (Playwright/Chromium, offline=true) | PASS — 12 slayt, klavye gezinmesi, gömülü ekranlar, 375/768/1280 genişlik, sıfır HTTP isteği ve sıfır JavaScript hatası |
| `python3 scripts/build-offline.py` | PASS — bağımsız HTML üretimi |
| `node scripts/build-presentation.mjs` (sağlanan Artifact Tool çalışma zamanı) | PASS — 12 slayt, PPTX paket/düzen/font, 2 düzenlenebilir tablo, gömülü veri tablosuna sahip düzenlenebilir grafik ve yeniden içe aktarma |
| `bash -n scripts/demo-setup.sh scripts/demo-ready.sh` | PASS |
| `make competition-package` ve arşiv manifest denetimi | PASS — 144 kaynak/sunum dosyası; tüm dosya özetleri eşleşti; gerçek token değerleri ve runtime/kimlik dosyaları yok |
| `sha256sum -c AgroChain-competition.sha256` (`dist` içinde) | PASS |
| Rehber yerel bağlantıları, 12 son PPTX slaydının görsel incelemesi ve `git diff --check` | PASS |

Buradaki süreler tek prova gözlemidir, kapasite ölçümü değildir. Aşama 7'nin
113 farklı test/ölçüm sonucu sunumda tarihsel kanıt olarak kullanılır; bu aşamada
113 testin tamamı yeniden çalıştırılmış gibi sunulmaz.

İlk temiz prova PDC yayılımında “0 eligible peers” ile başarısız oldu.
Peer keşfi hazır olma kapısı eklenip temiz prova baştan tekrarlandı ve geçti.
İlk başarısız kanıt `network/runtime/stage8-rehearsal-first-failure/` altında
korunur. İlk araç denemelerindeki Docker/soket/alt süreç kısıtları izinli
ortamda yeniden çalıştırılarak aşıldı; başarısız denemeler PASS sayılmadı.
Sunumun ilk doğrulaması grafik için gömülü çalışma kitabı eksikliğini buldu;
sabit 40/85 ve 50/50 değerleri düzenlenebilir çalışma kitabıyla paketlenip geçti.
İlk paket taraması kaynak kodundaki PEM ayrıştırma metnini anahtar sanarak durdu;
gerçek PEM blok başlangıcını ayıran kontrol ve regresyon testi eklendi. Özel
anahtar dosyalarını dışlayan dizin/uzantı sınırları korundu.

Yerel kanıtlar:

- `network/runtime/stage8-rehearsal/report.json`: PASS, aynı işlem kimlikleri,
  yeni genesis özeti ve originalRestored=true.
- `network/runtime/stage8-presentation/validation.json`: son PPTX özeti ve doğrulayıcı sonuçları.
- `network/runtime/stage8-presentation/final-*.png`: son PPTX'in render edilmiş slaytları.
- `network/runtime/demo/evidence.json`: yalnız sabit sentetik senaryoların alan listesi ve işlem kimlikleri.
- `network/runtime/stage8-offline/offline.png`: ağ kapalı tarayıcı kontrolü.

## Acceptance Criteria

- [x] Kaynak kopyasından desteklenen ortamda belgelenmiş kurulum adımları uygulanabilir.
- [x] Gerekli ortam açılır; yeni kimlikler ve boş disklerle aynı makinede doğrulandı.
- [x] Sabit demo verileri yüklenir ve yinelenen istekler kayıtları çoğaltmaz.
- [x] Normal ve şüpheli partiler gerçek uygulamada gösterilebilir.
- [x] Blockchain, PDC, zincir dışı saklama ve simülatör sınırları açıklanabilir.
- [x] 10–12 slayt hedefi, üç dakikalık akış ve ağdan bağımsız alternatif mevcut.
- [x] Mevcut ledger ve kimlikler prova sonrasında korunur.

## Assumptions

- Hedef ortam Linux x86-64; ilk araç indirme aşamasında internet var.
- Kullanıcı mevcut role bağlı geliştirme tokenını yerel dosyadan alır.
- Demo verileri sentetik ve sunumda gösterilmeye uygundur. Bu dışa aktarım
  genel özel ticari veri paylaşım aracı olarak kullanılmaz.
- Konuşma dili Türkçe, hedef teknik jüri; 12 slayt teknik anlatım, üç dakika
  ise ayrı canlı uygulama akışı içindir. Takım adı/gerçek kurum ortaklığı uydurulmadı.

## Known Limitations

- Açık, yerel pilot kabulünü engelleyen hata yoktur.
- Prova aynı fiziksel makinede, yeni kimlik/disklerle ve mevcut araç/bağımlılık
  önbellekleriyle yapıldı. İkinci fiziksel bilgisayarda insan takım arkadaşıyla
  kurulum veya sıfır önbellekle internet indirmesi bu aşamada yürütülmedi.
- Sunum araçla render edilip incelendi; Microsoft PowerPoint uygulamasında
  manuel açılış testi yapılmadı. Son sunum bilgisayarında font/görüntü kontrolü önerilir.
- Video dosyası çekilmedi; çalışır çevrimdışı HTML ve video çekim planı teslim edildi.
- Resmi yarışma portalına yükleme, takım üyeleri bilgisi ve güncel yarışma
  şartnamesine resmi uygunluk onayı bu mühendislik tesliminin kapsamı dışındadır.
- Token/CA/PDC üyesi yönetici erişimi geliştirme pilotu sınırındadır.
  Gerçek kurum entegrasyonu, üretim SSO'su, ülke çapında ölçek ve garanti hile
  tespiti iddiası yoktur. Anomali eşiği yasal sınır değildir.
- İnceleme zincir kaydı eklemelidir. Seed, sonradan sonuçlandırılmış incelemeyi
  silmez veya tekrar açmaz; demo sonraki provada inceleme geçmişini gösterir.

## Deferred to Future Stages

Sekiz aşamalı yerel pilot tamamlandı; otomatik yeni aşama açılmadı.
İkinci bilgisayarda takım provası, isteğe bağlı video çekimi ve resmi başvuru
işlemleri teslim sonrası operasyonlardır.

## Result

**PASS — Aşama 8 yerel yarışma paketi tamamlandı.**

Başlangıç noktası: [yarışma paketi](competition/README.md).
