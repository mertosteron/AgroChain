# Aşama 7 — Test ve ölçüm kılavuzu

Bu aşama yeni ürün özellikleri eklemez. Önceki aşamaların gerçek Fabric, PDC,
backend ve tarayıcı kontrollerini tek komutta çalıştırır; ölçüm ve temiz kurulum
kanıtlarını kaydeder. Kurum verileri SIMULATED; zincir işlemleri gerçek yerel Fabric'tir.

## Önkoşullar ve çalıştırma

Linux x86-64, Python ≥3.10, Docker/Compose erişimi; README'deki Fabric, Go,
Java/Maven araçları ve Python cryptography hazırlanmış olmalıdır. Aşama 6 ağı
hazırlanmış, sözleşme deploy edilmiş, rol sertifikaları/kaynak anahtarları ve backend
paketi hazır olmalıdır. Playwright 1.62.1 ve Chromium, Aşama 6 kılavuzundaki gibi
hazırlanır. Uygulamanın kendisi Node gerektirmez.

```bash
make backend-build
# Playwright yerel kurulum örneği; başka kurulumda NODE_PATH'i ona göre ayarlayın.
NODE_PATH="$PWD/network/runtime/ui-tools/node_modules" \
CHROMIUM_PATH=/usr/bin/chromium \
make stage7-check JAVA="$PWD/network/tools/jdk-21.0.12.1+1/bin/java"
```

Birden fazla Python kurulumu varsa desteklenen Python'un bulunduğu dizini PATH'in
başına alın. Aynı anda başka ağ yaşam döngüsü/test komutu çalıştırmayın. Aşama 7
komutları aralarında dosya kilidi kullanır; genel Make komutları bu kilidi kullanmaz.

`stage7-check` sırayla:

1. Ağ güvenlik/yapılandırma testleri, Go race/vet/kapsam, Java birim testleri,
   çapraz dil belge vektörü ve ölçüm istatistiği testleri.
2. Mevcut ağı açma, TLS/MSP/kanal/topoloji, gerçek chaincode reddetmeleri, gerçek PDC üyesi/üye
   olmayan istemci ve peer sorguları, belge değişikliği, MVCC çakışması ve kalıcılık.
3. HTTP/simülatör/işlem kurtarma, anomali, inceleme ve tüketici gizliliği testleri.
4. Gerçek tarayıcıda aktör akışı, denetçi karşılaştırması ve mobil tüketici ekranı.
5. HTTP uçtan uca ve Gateway ölçümleri.
6. İki bağımsız temiz ağ kurulumu ve aynı kabul senaryolarının tekrarı.

Herhangi bir zorunlu adım hata verirse komut sıfır dışı kodla biter ve rapor FAIL
kalır. Hata/atlanmış adım başarılı sonuç gibi gösterilmez. Birim turunda canlı testler
bilerek atlanır; özel canlı adımlarda ayrıca çalıştırılır.

## Temiz kurulum ve mevcut verilerin korunması

Test iki kez yeni `/tmp/agrochain-stage7-*` çalışma dizini, rastgele geliştirme
kimlikleri, genesis, sıfır ledger diskleri ve boş backend veritabanı oluşturur.
Kaynak kod mevcut çalışma ağacından kopyalanır; kimlikler, token'lar ve veritabanları
kopyalanmaz. Araçlar ve indirilen Go bağımlılıkları tekrar kullanılır. Bu, bağımlılık
indirmesinden başlayarak internet bağlantısı/temiz işletim sistemi testi değildir.

Portlar ve konteyner adları sabit olduğundan mevcut ağ geçici olarak durur.
**Varsayılan yedi `agrochain-*-data` diski ve mevcut kimlikler silinmez.** Her test
kurulumunun disk öneki `agrochain-stage7-<12 hex>` olur. Bu değer kimlik manifestine
bağlanır; farklı önekle başlatma/temizleme reddedilir. Temizleme mevcut güvenli
`clean-generated` kodunu yalnız geçici çalışma dizini ve ona ait yedi test diski
üzerinde çalıştırır. Sonunda mevcut ağ açılır ve sözleşme kontrol edilir.

Test sırasında açık UI kısa süreli ağ hatası gösterebilir. Programın yakalayabildiği
hata, SIGINT veya SIGTERM durumunda geri açma denenir. Elektrik kesintisi/SIGKILL
sonrası otomatik geri açma garanti edilemez. Böyle bir durumda kayıtlı test dizini ve
`network/runtime/stage7/*-stop.log` dosyalarını inceleyin; genel Docker temizleme
komutları kullanmayın. Özgün çalışma dizininde `make network-up chaincode-check`
ile geri açmadan önce geçici test konteynerlerini kendi çalışma dizininden durdurun.

İki yeni genesis'in hash'leri farklı olmalıdır; anahtar/salt/kimlik rastgeleliği
nedeniyle byte eşitliği beklenmez. İş sonuçları her iki turda aynı olmalıdır:

| Senaryo | Beklenen |
| --- | --- |
| Normal, 20 → 28 TL/kg | %40; NO_SIGNAL; pilot eşiği %50 |
| Şüpheli, 20 → 37 TL/kg | %85; REVIEW_REQUIRED; yetkili insan incelemesi |
| Değiştirilmiş belge | Orijinal geçer; değiştirilmiş içerik DOCUMENT_HASH_MISMATCH ile reddedilir |

Yinelenen operationId HTTP tarafında aynı makbuzu döndürür; doğrudan chaincode
yinelenen komutu reddeder. İki eşzamanlı öneride yalnız biri VALID olur. Bu davranışlar
ayrı testlerdir; ölçümün eşzamanlılık seviyesi 1'dir.

## Ölçüm yöntemi

`make stage7-measure` yalnız ölçümleri; `make stage7-reproduce` yalnız iki temiz
kurulumu çalıştırır. Bunlar bütün test kapısının geçtiğini tek başına kanıtlamaz.
Gateway ölçümü Aşama 6 kabul testinin ürettiği normal partiye ihtiyaç duyar.

| Ölçüm | Başlangıç / bitiş | Örneklem |
| --- | --- | --- |
| Gateway submit → VALID | Önceden endorse edilmiş işlem gönderilmeden hemen önce → gerçek commit-status başarılı dönünce | 1 ısınma hariç 30 ardışık CreateBatch |
| Ortak sorgu | Gateway GetBatch evaluate çağrısı | 30 |
| Özel sorgu | Yetkili üreticinin GetPurchase evaluate çağrısı | 30 |
| HTTP komutu | İstemci HTTP POST başlangıcı → COMMITTED yanıtı | 42; 7 komut türünden 6'şar örnek |
| Tam rota | CreateBatch başlangıcı → 7 komut, anomali ve tüketici projeksiyonu görünür | 6; 3 normal + 3 şüpheli |

HTTP'de bir normal ve bir şüpheli rota ısınma olarak dışlanır. Tam rota sorgulama
beklemesi 100 ms aralıklıdır; insan bekleme süresi içermez. TLS, yerel disk,
simülatör doğrulaması ve endorsement HTTP süresine dahildir. Gateway submit
ölçümü ise endorsement/simülatör hazırlığını dışlar. Her iki ölçüm gerçek VALID
onayına dayanır; HTTP kabulü veya endorsement tek başına commit sayılmaz.

Ortak sorgu her örnekte yeni oluşturulan partiyi; özel sorgu aynı normal kabul
partisinin alış belgesini okur. Özel sorguda tekrar okuma/önbellek etkisi vardır.
Masaüstü makinenin CPU'su başka süreçlerden izole edilmemiştir.

Süreler monoton saat (`perf_counter_ns` / `System.nanoTime`) ile ölçülür. Medyan
çift örneklemde ortadaki iki sayının ortalaması; p95 sıralı örneklemde
`ceil(0.95*n)` konumudur. Küçük örneklemde p95'in kararlılığı sınırlıdır.
JSON kaydı tekil süreleri, istek/yanıt veya endorse edilmiş işlem byte boyutlarını,
örneklem sayısını ve hata sayısını saklar. Özel payload, salt, token ve anahtar
ölçüm raporuna yazılmaz. HTTP üstbilgi/TLS çerçevesi boyutlara dahil değildir.

Tek makine, önceden çalışan ledger, ısınmış araç/işletim sistemi önbellekleri,
LevelDB ve düşük eşzamanlılık koşulları geçerlidir. Kayıtlar üretim performansı,
ulusal kapasite, azami TPS veya hizmet garantisi değildir. Fiyat eşiği hukuki sınır
olarak sunulmaz. Yük/ölçek testleri bu pilot ölçümünden çıkarılamaz.

## Kanıtlar

`network/runtime/stage7/` Git dışıdır ve kısıtlı dosya izinleriyle oluşturulur:

- `check-report.json`: ortam, her komutun durumu/süresi, ölçüm özetleri, temiz tur sonuçları.
- `progress.json` ve adım başına `.log`: kesinti halinde en son durum ve teşhis.
- `http-measurements.json`, `gateway-measurements.json`: ham süre/boyut kayıtları.
- `measure-report.json`, `reproduce-report.json`: ayrı çalıştırılan alt kapıların sonuçları.

Paylaşılacak tamamlanma/ölçüm özeti `docs/STAGE7_REPORT.md` içindedir. Uzak CI
başarısı yerel sonuçlardan varsayılmaz. Sunum paketi, sabit demo seed komutu ve
başka bilgisayardaki kurulum provası Aşama 8'e bırakılır.

Bu çalışmanın özel içeriklerden arındırılmış süre, byte boyutu ve işlem kimliği
örnekleri `docs/STAGE7_MEASUREMENTS.json` içinde sürüm kontrolüne alınabilir.
Bu kopya yeni bir ölçüm çalıştırıldığında otomatik güncellenmez; tarihli raporun
kanıtıdır. Yeni çalışmanın sonuçları kendi runtime JSON dosyalarından okunmalıdır.
