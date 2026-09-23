# AgroChain durum özeti — 23 Eylül 2026

**Sekiz aşamalı yerel yarışma pilotunun mühendislik çalışması tamamlandı.**
Bu sonuç, üretim ölçeği veya resmi yarışma başvurusunun tamamlandığı anlamına gelmez.

## Tamamlananlar

| Aşama | Mevcut sonuç |
| --- | --- |
| 1 — Teknik sözleşme | Roller, veri sınırları, yaşam döngüsü ve kabul ölçütleri tanımlı. |
| 2 — Fabric ağı | Dört kuruluş/peer, üç Raft orderer, TLS ve tekrarlanabilir kurulum. |
| 3 — Chaincode | Parti/teslim akışı, yetki, miktar, sürüm ve tekrar kontrolleri. |
| 4 — Gizlilik ve doğrulama | PDC erişimi, imza/saltlı kanıt, özgün ve değiştirilmiş belge testleri. |
| 5 — Backend | Spring Boot, Fabric Gateway, kalıcı işlem kurtarma ve imzalı kurum simülatörleri. |
| 6 — Analiz ve arayüz | Açıklanabilir %40/%85 karşılaştırması, aktör/denetçi ekranları, özel inceleme ve tüketici QR. |
| 7 — Test ve ölçüm | 113 farklı otomatik test, yerel ölçümler ve iki temiz ağ tekrarı raporlandı. |
| 8 — Yarışma paketi | Sabit/idempotent demo, kurulum rehberi, 12 slayt, üç dakikalık akış ve çevrimdışı HTML. |

Aşama 8'de yeni kimlikler ve boş test diskleriyle kurulum ayrıca tekrarlandı;
aynı seed ikinci kez yeni kayıt üretmedi. Özgün ağ ve veriler korundu.
Kurum verileri **SIMULATED**, yerel Fabric işlemleri gerçektir.

## Yarışma öncesinde kalan işler

1. **İkinci fiziksel bilgisayarda takım provası:** Depoyu klonlayıp kurulumu,
   normal/şüpheli senaryoları ve kapatıp açmayı başka ekip üyesiyle deneyin.
   Mevcut temiz provalar aynı makinede, araç önbellekleriyle yapıldı.
2. **Sunum bilgisayarında son kontrol:** PPTX'i PowerPoint/LibreOffice'te açın;
   fontları, projeksiyonu, üç dakikalık anlatımı ve çevrimdışı dosyayı deneyin.
   Telefonla QR gösterilecekse yerel ağ adresini ayarlayın.
3. **Resmi başvuru işlemleri:** Güncel şartnameyi kontrol edin; takım bilgileri,
   istenen belgeler ve portal yüklemesini tamamlayın. Bunlar yapılmış sayılmıyor.
4. **İsteğe bağlı video:** Çekim planı mevcut, video henüz kaydedilmedi.
   Çalışan çevrimdışı HTML alternatifi hazır.

Gerçek kurum API erişimi, üretim SSO'su ve ülke çapında yük testleri mevcut
yarışma pilotunun zorunlu eksikleri değildir; sonraki ürünleşme çalışmalarına aittir.
Anomali eşiği yasal sınır, sinyal de kesin hile kararı değildir.

## Kanıt ve teslim

- [Aşama 7 raporu ve ölçümleri](STAGE7_REPORT.md)
- [Aşama 8 kabul raporu ve sınırlamaları](STAGE8_REPORT.md)
- [Yarışma paketi ve kurulum](competition/README.md)
- [12 slaytlık sunum](competition/AgroChain.pptx)
- [Çevrimdışı gösterim](competition/offline.html)

`make competition-package` kaynak ve sunum arşivini yeniden oluşturur.
Yerel tokenlar, özel anahtarlar, ledger diskleri ve geçici sunum dosyaları
GitHub'a veya dağıtım arşivine dahil edilmez.
