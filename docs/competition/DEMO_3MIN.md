# Üç dakikalık canlı gösterim

Hazırlık sahne süresine dahil değildir: `demo-seed` ve `demo-check` geçmeli,
incelemeci oturumu açık, iki sabit parti karşılaştırmada seçili ve tüketici
sayfası ayrı sekmede hazır olmalıdır. Anahtar girişini projeksiyona vermeyin.

| Süre | Ekran / eylem | Konuşma |
| --- | --- | --- |
| 0:00–0:30 | AgroChain çalışma alanı, SIMULATED etiketi | “Bu pilotta ürün teslimlerini gerçek yerel Fabric ağına kaydediyoruz. ÇKS, e-Fatura, HKS ve U-ETDS verileri açıkça etiketlenmiş simülasyonlardır.” |
| 0:30–1:05 | BAT-DEMONORMAL01 ve BAT-DEMOSUSPICIOUS01 karşılaştırması | “Her iki partide alış 20 TL/kg. Solda raf 28 TL, artış yüzde 40. Sağda raf 37 TL, artış yüzde 85. Yüzde 50 pilot eşiğini sadece sağdaki parti aşıyor.” |
| 1:05–1:35 | İki geçmişi ve işlem kimliklerini göster | “Teslim tekliflerini ilgili kuruluş kabul ediyor. Nakliye ve perakende adımları zincirde. Backend hesabını sözleşme özel girdilerden yeniden doğruluyor.” |
| 1:35–2:15 | Şüpheli parti ayrıntısı, inceleme paneli | “Yetkili incelemeci gerekçeli karar ekleyebilir. Bu, özgün hesabı değiştirmez. Eşik yasal sınır, sinyal de hile kanıtı değildir. Ticari kayıtları yalnız yetkili kuruluşlar okuyabilir.” |
| 2:15–2:45 | Normal lotun tüketici sekmesi | “Tüketici kaynağı ve teslimleri görür. Özel fiyatlar ve inceleme metni bu yanıtın içinde yoktur. QR aynı sayfaya götürür.” |
| 2:45–3:00 | Sonuç | “Bu, tek makinede tekrarlanabilir bir yarışma pilotu. Gizlilik, belge tahrifatı ve iki fiyat senaryosu otomatik testlerle doğrulanıyor.” |

İnceleme yazma işlemini üç dakikalık akışın zorunlu parçası yapmayın. Jüri
isterse ek sürede “İncelemeyi başlat” ve gerekçeyle sonuçlandır adımlarını
gösterin. Bu eklemeli, gerçek zincir işlemidir; tekrar açılmaz/silinmez. Seed
inceleme durumunu sıfırlamaz. Sonraki provada sonuçlanmış inceleme geçmişini gösterin.

## Kısa teknik soru cevap

- **Neden blockchain?** Kuruluş kimlikleriyle onaylanan ortak teslim ve kanıt
  kaydı için. Büyük belge dosyaları ve uygulama listeleri zincir dışında kalır.
- **Fiziksel ürünün doğruluğu garanti mi?** Hayır. Zincir kaydedilen iddiayı
  ve kaydın bütünlüğünü korur; saha denetimi ve güvenilir kaynak ayrı gereksinimdir.
- **Kurum entegrasyonları gerçek mi?** Hayır, imzalı ve değiştirilebilir adaptör
  sınırının arkasındaki yerel simülatörlerdir.
- **Kimler alış fiyatını görür?** Producer, Retailer ve Regulator. Taşıyıcının
  alış kaydı yetkisi yoktur. Raf/analiz yalnız Retailer ve uygun Regulator rolleri.
- **Anomali ne anlama gelir?** Yapılandırılmış pilot eşiğinin aşılması. Nakliye
  sadece bağlamdır; ekonomik veya hukuki ihlal hükmü değildir.
- **Ölçümler ölçek kanıtı mı?** Hayır. Tek makine, eşzamanlılık 1 ve küçük örneklem.

## Kesinti

10 saniyeyi aşan bağlantı sorununun sahnede çözümüne çalışmayın. “Canlı ağ
erişiminde sorun var; şimdi önceki doğrulanmış sentetik kayıtların çevrimdışı
gösterimine geçiyorum” deyip `offline.html` dosyasını açın. Aynı zaman akışını
kullanın, ekrandaki çevrimdışı uyarısını saklamayın.
