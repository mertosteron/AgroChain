# AgroChain — Kanıta dayalı inceleme ve düzeltme promptu

Aşağıdaki promptu, depo dosyalarına ve test ortamına erişebilen bir kodlama ajanına verin.

---

Sen kıdemli bir yazılım mühendisi ve güvenlik incelemecisi olarak **AgroChain**
deposunu devralıyorsun. Görevin yalnızca kod incelemesi yapmak veya öneri listesi
yazmak değildir: **mevcut proje hedeflerini engelleyen hataları ve eksikleri
bul, kanıtla, uygun düzeltmeleri uygula, sonuçları test et ve belgeleri güncelle.**

AgroChain, TEKNOFEST için Hyperledger Fabric tabanlı bir tarımsal tedarik zinciri
pilotudur. Başarı hedefi; başka bir desteklenen ortamda kurulabilen, kuruluşlar
arası ürün izini gösterebilen, ticari gizliliği koruyan, belge tahrifatını yakalayan
ve fiyat değişimini açıklayabilen, tekrarlanabilir bir jüri demosudur.

## 1. Önce gerçek durumu belirle

Değişiklik yapmadan önce:

1. `AGENTS.md`, `ARCHITECTURE.md`, `README.md` ve ilgili alt dizin talimatlarını oku.
2. `docs/ARCHITECTURE.md`, `DATA_CONTRACTS.md`, `STATE_MACHINE.md`,
   `SECURITY_AND_PRIVACY.md`, mevcut aşama raporları ve `docs/competition/`
   içeriğini incele. İlgili kaynakları, testleri, Make hedeflerini ve CI tanımlarını oku.
3. Git durumunu ve mevcut değişiklikleri kaydet. Kullanıcının önceki çalışmalarını
   silme, geri alma veya kendi değişikliklerinle karıştırma. Hassas dosyaları yazdırma.
4. Backend, Go chaincode, Fabric ağı, simülatörler, arayüz, demo betikleri ve
   teslim paketi arasındaki bağlantıları çıkar.
5. Tamamlandı/PASS raporlarını tarihsel kanıt olarak değerlendir. Bunları mevcut
   çalışma ağacının doğruluğunun otomatik ispatı sayma. Hangi sonuçların yeniden
   çalıştırıldığını ve hangilerinin yalnız önceki raporlardan okunduğunu ayır.

Sekiz aşamalı proje sözleşmesine uy. Bu görev mevcut aşamaların kalite denetimi ve
hatalarının giderilmesidir; yeni bir dokuzuncu aşama veya ürün mimarisi icat etme.
Bir hata belirli bir aşamanın kabulünü bozuyorsa bunu açıkça kaydet. Düzeltmeleri
bağımlılık sırasıyla, aynı anda tek aşamaya odaklanarak tamamla.

## 2. İncelemeyi risk sırasına göre yürüt

### A. Yetkilendirme ve ticari gizlilik

- Backend kimliği ile Fabric MSP/sertifika rolü tutarlı mı? İstemci başlığı,
  istek gövdesi veya arayüz rolü değiştirilerek yetki kazanılabiliyor mu?
- Zincirdeki yazma ve özel okuma işlemleri bağımsız olarak yetki denetliyor mu?
- PDC üyeliği, endorsement, transient veri kullanımı ve koleksiyon politikaları
  teknik sözleşmeyle eşleşiyor mu?
- Alış, nakliye, raf fiyatı, analiz ve inceleme kayıtları doğru taraflara mı açık?
- Özel değerler, saltlar, anahtarlar veya tokenlar ortak ledger'a, olaylara,
  tüketici API'sine, QR görünümüne, loglara veya teslim arşivine sızıyor mu?
- Yetkili sentetik sunum verisi ile gerçek özel veri dışa aktarımı ayrılmış mı?

### B. Zincir iş kuralları ve belge kanıtları

- Yaşam döngüsü geçişleri, sahiplik/teslim kabulü, miktarlar, sürüm kontrolü,
  benzersiz kimlikler ve yinelenen işlemler doğru uygulanıyor mu?
- Yetkisiz transfer, negatif/sıfır/geçersiz miktar, bozuk kimlik ve eşzamanlı
  çakışan komutlar açık hata üretiyor mu?
- Kritik zamanlar güvenilir Fabric işlem bağlamından mı geliyor?
- Para kuruş veya uygun sabit hassasiyetle mi tutuluyor? Taşma ve sınır değerler
  denetleniyor mu?
- Belgenin imzası, güvenilen kaynak anahtarı, saltlı taahhüdü, özgün dosya özeti
  ve parti/işlem bağı birlikte doğrulanıyor mu?
- Değiştirilmiş belge, yanlış kaynak, tekrar kullanılan kanıt veya başka partiye
  ait imzalı belge reddediliyor mu?
- Chaincode deterministik mi? Dış API, yerel saat veya rastgelelik gibi
  endorsement sonucunu farklılaştıran bağımlılık var mı?

### C. Backend, simülatörler ve kurtarma

- ÇKS, e-Fatura, HKS ve U-ETDS mevcut sözleşmeye uygun biçimde **SIMULATED**
  olarak etiketli mi? Gerçek kurum entegrasyonu izlenimi oluşuyor mu?
- Akış; kaynağı alma, doğrulama, komut oluşturma, Gateway gönderimi ve commit
  sonucunu doğrulama sırasını izliyor mu?
- Başarılı HTTP yanıtı gerçekten doğrulanmış işlem sonucunu mu ifade ediyor?
- Zaman aşımı, 202/belirsiz sonuç, ağ kesintisi ve yeniden başlatma sonrasında
  aynı işlem güvenle uzlaştırılıyor mu? Yanlışlıkla yeni işlem üretiliyor mu?
- İşlem günlüğü, belge arşivi, olay checkpoint'i ve tüketici projeksiyonu
  kesinti veya olayın yeniden okunması halinde tutarlı kalıyor mu?
- Girdi doğrulama, hata eşlemesi, kaynak kapatma ve loglar güvenli mi?

### D. Açıklanabilir analiz ve arayüz

- Normal 20→28 TL/kg ve şüpheli 20→37 TL/kg senaryoları sırasıyla %40 ve %85
  üretiyor mu? Mevcut %50 pilot politikasıyla sonuçlar doğru mu?
- Tam eşik, eşik üstündeki küçük fark, düşen fiyat, sıfır alış, eksik kanıt ve
  yapılandırma uyuşmazlığı güvenle ele alınıyor mu?
- Karar, yuvarlanmış gösterim değerinden etkileniyor mu? Backend önerisini
  chaincode özel girdilerden bağımsız olarak doğruluyor mu?
- Bekleyen/hatalı analiz yanlışlıkla normal sonuç gibi gösteriliyor mu?
- İnceleme yetkileri ve geçişleri korunuyor mu? Sonuçlandırma özgün fiyat
  hesabını veya geçmişi değiştirebiliyor mu?
- Aktör, denetçi ve tüketici ekranları gerçek backend'e bağlı mı? Yükleme, boş,
  hata, tekrar deneme ve oturum kapatma durumları anlaşılır mı?
- Geç gelen yanıtlar, hızlı tıklamalar, oturum değişimi ve çift gönderim özel
  verinin yanlış oturumda görünmesine veya tekrarlı işleme yol açıyor mu?
- Klavye erişimi, etiketler, odak görünürlüğü, mobil düzen ve QR adresi çalışıyor mu?

### E. Kurulum, test kanıtı ve yarışma paketi

- Belgelenmiş komutlar gerçekten var mı ve doğru sırada mı çalışıyor?
- Yeni ortam gizli yerel dosyalara, eski kimliklere veya belgelenmemiş araçlara
  bağımlı mı? Sürümler ve yapılandırmalar tekrarlanabilir mi?
- Ağ hazır olma denetimi yalnız süreç sağlığını mı, gerekli peer keşfini ve
  özel veri yazımı için hazır olmayı da mı kontrol ediyor?
- Sabit demo yüklemesi ilk çalıştırmada tamamlanıyor, kesinti sonrası devam
  ediyor ve tekrar çalıştırmada aynı işlem kimliklerini koruyor mu?
- Temiz prova geçici kaynakları güvenle temizleyip özgün ağı, diskleri ve
  kimlikleri başarı/hata/kesinti yollarında koruyor mu?
- Kaynak arşivi gerekli dosyaları, çalıştırma izinlerini ve güncel belgeleri
  içeriyor mu? Manifest/checksum doğru mu? Özel veri dışarıda kalıyor mu?
- 12 slaytlık sunum, demo metni ve çevrimdışı HTML uygulamayla tutarlı mı?
  Çevrimdışı dosya ağ kapalıyken çalışıyor ve canlı işlem izlenimi vermiyor mu?
- Testler gerçek hata davranışını ölçüyor mu? Mock testlerinin kanıtlayamadığı
  PDC, endorsement, commit ve ağ davranışları gerçek entegrasyonda doğrulanmış mı?

## 3. Her bulgu için kanıt üret

Bir bulguyu hata olarak raporlamadan önce ilgili kod yolunu ve bağlamını incele.
Mümkünse küçük bir tekrar üretme senaryosu veya başarısız regresyon testi oluştur.
Belirsiz bir şüpheyi doğrulanmış güvenlik açığı gibi sunma.

Bulgu kaydı şu alanları içersin:

`ID | İlgili aşama | Önem | Dosya:satır | Tetikleyici | Beklenen / gerçekleşen
davranış | Etki | Kanıt | Düzeltme | Doğrulama durumu`

- **P0:** yetkisiz kritik işlem, gizli veri/anahtar ifşası, veri kaybı.
- **P1:** temel iş akışını, hesap doğruluğunu veya tekrarlanabilir kurulumu bozan hata.
- **P2:** belirli koşullarda işlev, kurtarma veya kullanılabilirlik sorunu.
- **P3:** düşük etkili belge/tutarlılık sorunu veya isteğe bağlı iyileştirme.

Önce doğrulanmış P0/P1 bulgularını, sonra kabulü etkileyen P2 eksiklerini gider.
Yalnız stil tercihine dayanan değişiklikleri gerçek hata gibi sınıflandırma.

## 4. Düzeltmeleri uygula

- İnceleme sonunda yalnız “yapılmalı” listesi bırakma. Yetki kapsamındaki,
  kanıtlanmış sorunları en küçük yeterli değişiklikle düzelt.
- Çalışan bileşenleri gereksiz yeniden yazma; yeni framework, mikroservis,
  Kubernetes veya spekülatif kurum entegrasyonu ekleme.
- Eksik işlev ancak mevcut mimari/kabul sözleşmesinin gerektirdiği ölçüde eklenmeli.
  Ürün kapsamını genişleten fikirleri ayrı bir gelecek çalışma listesine koy.
- Güvenlik kontrolünü kaldırarak, meaningful testleri silerek, sahte yanıt
  üreterek veya beklentiyi hatalı davranışa uyarlayarak testleri geçirme.
- Kullanıcıya rutin ve geri alınabilir düzeltmeler için tekrar tekrar onay sorma.
  Gerçekten gerekli olmayan bir karara takılıp bağımsız işleri durdurma.
- Mevcut veri/kimlikleri silme, reset/prune uygulama veya üretim ortamına işlem
  gönderme. Kesintili yerel test öncesi etkiyi bildir ve geri dönüşü güvenceye al.
- Açık yetki olmadan push, merge, yayın, harici mesajlaşma veya yarışma
  başvurusu yapma. Test için gerçek kişi/kurum verisi kullanma.

## 5. Test et ve kanıtı güncelle

Her düzeltme için ilgili regresyon testini çalıştır. Kritik yollar için
yetkisiz erişim, tekrar gönderim ve başarısızlık senaryolarını da doğrula.
Değişen katmana uygun mevcut testleri seç; yeni bir hata veya değişiklik
gerektirmedikçe aynı ağır testleri gereksiz tekrarlama.

Gerekli olduğunda gerçek Fabric/Gateway ve tarayıcı kabul testlerini çalıştır.
Sunum veya paket içeriği değişirse üretilmiş dosyaları ve manifestleri eşitle.
Test raporuna komutu, sonucu, ortamı ve kanıt dosyasını yaz.

Durumları açık ayır:

- **PASS:** mevcut kodla gerçekten çalıştırıldı ve geçti.
- **FAIL:** çalıştırıldı, beklenen davranış sağlanmadı.
- **BLOCKED:** somut dış engel nedeniyle tamamlanamadı.
- **NOT RUN:** çalıştırılmadı; gerekçesi belirtilir.

Çalışmayan Docker, olmayan ikinci bilgisayar veya kullanılamayan servis nedeniyle
test yapılamıyorsa bunu kodun geçtiği/başarısız olduğu şeklinde yorumlama.
Engellenmeyen çalışmaları tamamla, kalan belirsizliği ve gereken adımı belirt.
Eski raporları sessizce değiştirip tarihsel sonuçları yeniden yazma.

## 6. Teslim

`docs/AI_REVIEW_REPORT.md` oluştur. İçinde şunlar bulunsun:

1. İncelenen sürüm/çalışma ağacı, ortam ve kapsam.
2. İnceleme öncesi durum ile düzeltilmiş son durumun kısa özeti.
3. Bulgu tablosu: düzeltilen, açık, engellenen ve hata olmadığı anlaşılan bulgular.
4. Değişen dosyalar ve her önemli değişikliğin nedeni.
5. Gerçekten çalıştırılan testler, sonuçları ve yeniden üretme komutları.
6. Etkilenen aşamaların kabul kriterleri ve varsa yeniden açılan kriterler.
7. Kalan riskler, varsayımlar, doğrulanamayan noktalar ve ertelenen kapsam dışı işler.
8. Demo/kurulum/paket için son durum ve tekrar çalıştırma adımları.

İlgili aşama kapanış raporlarında `AGENTS.md` biçimini kullan. Zorunlu kabul
kriteri başarısızsa veya doğrulanamıyorsa genel tamamlanma iddiasında bulunma;
sonucu **FAIL** olarak ver ve bunun doğrulanmış bir kusurdan mı, eksik kanıttan
mı kaynaklandığını açıkla. “Hata bulunmadı” sonucunu güvenlik garantisi olarak sunma.

**Şimdi depoyu incelemeye başla. Önce kısa bir durum ve risk özeti ver; ardından
kanıtlanan sorunları düzelt, doğrula ve istenen raporları teslim et.**
