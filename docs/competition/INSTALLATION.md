# Kurulum ve prova

## Desteklenen ortam

Linux x86-64, Docker Engine >=24, Compose >=2.20, Bash >=4.4, Make, curl, jq,
tar, OpenSSL ve Python >=3.10. Python'da PyYAML ve cryptography gerekir.
Fabric CLI **2.5.15**, Java **21.0.12.1+1**, Maven **3.9.11**, Go **1.27.1**
proje sürümleridir. Bilinen yerel ortam ve ölçümler [Aşama 7 raporundadır](../STAGE7_REPORT.md).
Başka işletim sistemi desteği veya ölçülmüş minimum RAM iddiası yoktur.

1. Depoyu klonlayın ya da kaynak kopyasını ayrı bir dizine açın. Komutları kökte
   çalıştırın. Docker erişimini `docker info` ile doğrulayın.
2. [Ağ rehberindeki](../../network/README.md#prerequisites-and-versions) resmi,
   checksum doğrulamalı Fabric CLI kurulumunu tamamlayın. Ağ rehberi gerekli
   sistem paketlerini de listeler. Betikler sistem paketlerini veya sudo'yu yönetmez.
3. Yeni çalışma dizininde aşağıdaki adımları çalıştırın. İlk indirme internet
   gerektirir; sahnede indirme yapılmamalıdır.

```bash
# Arch'ta eski Python önceliğini engellemek için:
export PATH=/usr/bin:/bin:$PATH
python3 -c 'import yaml, cryptography; print("Python bağımlılıkları hazır")'
make demo-setup
```

Bu komut sabit araçları kurar, dört peer/üç orderer ağını başlatır, chaincode'u
dağıtır, rastgele geliştirme kimlikleri ve tokenları üretir, kurum kaynaklarını
başlatan mevcut gizlilik kabul testini çalıştırır ve backend'i derler. Özel veri
yazımından önce `demo-ready` her peer'in diğer kuruluşların yüklü sözleşmesini
keşfettiğini bekler; yalnız konteyner sağlık kontrolü yeterli değildir. Mevcut
kimlik manifesti varsa durur; verileri sıfırlamaz. Tamamlanınca:

```bash
# Terminal 1: açık bırakın
make backend-run
# Terminal 2:
make demo-seed demo-check
```

`http://localhost:8080` adresini açın. Giriş anahtarını yalnız yerel
`network/runtime/backend/tokens.json` dosyasından alın. `regulator:reviewer`
denetçi karşılaştırması ve inceleme, `retailer:retailer` raf/analiz,
`producer:producer` üretici, `logistics:carrier` taşıyıcı içindir.
Anahtarı sunuma, ekran kaydına, Git'e veya mesajlaşmaya kopyalamayın.

## Beklenen sonuç ve tekrar çalıştırma

| Parti | Alış / kg | Raf / kg | Artış | Sonuç |
| --- | ---: | ---: | ---: | --- |
| BAT-DEMONORMAL01 | 20 TL | 28 TL | %40 | NO_SIGNAL |
| BAT-DEMOSUSPICIOUS01 | 20 TL | 37 TL | %85 | REVIEW_REQUIRED |

İki parti 100 kg standart domatestir. Nakliye toplamı 200 TL'dir. Değerler
sentetiktir. Eşik %50 pilot parametresidir, yasal sınır değildir.
Kriptografik anahtar/salt/işlem zamanı rastgeledir; sabit olan iş akışı ve
sonuçlardır. `demo-seed` aynı istekleri yeniden gönderir, yeni parti oluşturmaz.
Kesintiden sonra aynı komutu kullanın; backend SQLite günlüğünü silmeyin.
Günlük kaybolduysa mevcut zincire yeni kimliklerle aynı kayıtları zorlamayın.
İşlem günlüğü ve belge arşivini uygun yedekten geri getirin.

Kapatma: backend terminalinde Ctrl-C, ardından `make network-down`.
Bu işlem ledger disklerini korur. Sonraki açılış:

```bash
make network-up chaincode-check demo-ready backend-build
make backend-run
# Başka terminal:
make demo-seed demo-check
```

Eski sözleşme sürümü varsa [Aşama 6 yükseltme rehberini](../STAGE6_UI.md) izleyin.
Sıra 6'yı farklı bir ağa körlemesine uygulamayın. Yeni kanal sıra 1 kullanır.
`make clean-generated` anahtar, belge ve ledger verilerini siler; demo açma
veya arıza giderme komutu olarak kullanmayın.

## Sunumdan önce

- Ağ, backend ve `demo-check` başarılı olmalı. Değerlendirme bekliyorsa normal
  kabul etmeyin; ağ, kaynak anahtarları ve politika/eşik eşleşmesini kontrol edin.
- İncelemeciyle iki sabit partiyi seçin. Karşılaştırmayı bir kez açın.
- Tüketici bağlantısını hazırlayın: `http://localhost:8080/consumer.html?lot=LOT-DEMONORMAL01`.
- Dizüstü ekranında tüketici bağlantısı yeterlidir. Telefon için
  `AGROCHAIN_PUBLIC_BASE_URL` ve `SERVER_ADDRESS` yerel ağ adresine ayarlanmalıdır.
  Varsayılan QR'daki localhost telefonda sunucu değildir.
- [Çevrimdışı dosyayı](offline.html) internet kapalıyken açın. Token girişi
  veya backend gerektirmediğini kontrol edin.

## Sık karşılaşılan sorunlar

| Belirti | İzlenecek yol |
| --- | --- |
| 8080 bağlantısı reddediliyor | Backend terminalini ve port çakışmasını kontrol edin. |
| Docker yetkisi yok | Kullanıcının Docker erişimini düzeltin; veri temizlemeyin. |
| Kaynak anahtarı uyuşmuyor | Mevcut ağla birlikte oluşturulan kaynak anahtarlarını geri yükleyin. |
| Bekleyen analiz | Eşik %50 ve CFG-PRICE001 eşleşmesini, peer bağlantısını kontrol edin. |
| Seed kimlik çakışması | Aynı backend günlüğünü kullanın; sabit demo tanımını değiştirmeyin. |
| Ağ sunumda çalışmıyor | OFFLINE.md geçişini kullanın; canlı işlem yapıldığını söylemeyin. |
