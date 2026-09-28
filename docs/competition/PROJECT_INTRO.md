# AgroChain — Proje Tanıtımı

## AgroChain nedir?

AgroChain, tarımsal ürünlerin (bu pilotta domates) üreticiden tüketiciye kadar
olan yolculuğunu blok zinciri üzerinde şeffaf ve değiştirilemez biçimde
kaydeden bir TEKNOFEST pilot projesidir. Amaç; tedarik zincirindeki
kuruluşların (üretici, lojistik, perakendeci, düzenleyici kurum) aynı
kayıtlar üzerinde anlaşmasını sağlamak, ticari açıdan hassas bilgileri
(alış/satış fiyatı gibi) yalnızca yetkili taraflara açmak ve olağandışı
fiyat artışlarını otomatik olarak fark edip incelemeye açmaktır.

## Çözmeye çalıştığımız problem

Bir ürün tarladan markete gelene kadar birden fazla kurumun elinden geçer,
ama bu kurumların kayıtları genelde birbirinden kopuktur: kim, ne zaman,
hangi fiyata teslim aldı sorusunun güvenilir ve tek bir kaynaktan cevabı
yoktur. Aynı zamanda fiyatların herkese açık paylaşılması istenmez, ama
fahiş fiyat artışlarının da fark edilebilmesi gerekir. AgroChain bu iki
ihtiyacı birlikte karşılamayı hedefler: gizlilik + izlenebilirlik +
açıklanabilir bir uyarı mekanizması.

## Nasıl çalışıyor?

Sistem, Hyperledger Fabric adlı bir blok zinciri altyapısı üzerine kurulu.
Dört kuruluş (Üretici, Lojistik, Perakendeci, Düzenleyici) aynı ortak
deftere bağlı, ancak her birinin görebileceği bilgi farklı. Örnek bir
domates partisinin (100 kg) yolculuğu şöyle işliyor:

1. Üretici partiyi sisteme kaydeder ve taşınmaya sunar.
2. Lojistik firma teslim alır, nakliye ücretini kaydeder, perakendeciye
   teslime çıkarır. Lojistik firma hiçbir zaman ürünün sahibi olmaz, sadece
   taşır.
3. Perakendeci teslim alır, sahipliği devralır ve rafa koyduğu satış
   fiyatını bildirir.
4. Sistem, alış ve satış fiyatı arasındaki farkı otomatik ve değiştirilemez
   bir kuralla hesaplar. Fark, belirlenen eşiği (bu pilotta %50) aşarsa
   parti "incelemeye açık" olarak işaretlenir.
5. Düzenleyici kurum (regülatör), işaretlenen partileri inceleyip gerekçeli
   bir not ekleyebilir. Bu not orijinal kaydı değiştirmez, sadece üzerine
   eklenir.
6. Tüketici, üründeki QR kodu okutarak partinin nereden geldiğini görebilir
   — fiyat bilgisi tüketiciye hiçbir zaman gösterilmez.

### Somut örnek

İki senaryo ile göstermek gerekirse: her iki partide de üreticinin alış
fiyatı 20 TL/kg.

- **Normal parti:** Raf fiyatı 28 TL/kg → artış %40 → eşiğin (%50) altında,
  sinyal üretilmez.
- **Şüpheli parti:** Raf fiyatı 37 TL/kg → artış %85 → eşiği aştığı için
  sistem otomatik olarak inceleme sinyali üretir.

Bu hesap tamsayı aritmetiğiyle, yuvarlama hatası olmadan yapılır ve iki
farklı doğrulayıcı (arka uç ile akıllı sözleşmenin kendisi) tarafından ayrı
ayrı tekrar hesaplanarak kontrol edilir.

## Kim neyi görebilir?

| Bilgi | Kim görür |
| --- | --- |
| Alış fiyatı | Üretici, Perakendeci, Düzenleyici |
| Nakliye ücreti | Lojistik, Perakendeci, Düzenleyici |
| Satış fiyatı ve inceleme notu | Perakendeci, yetkili Düzenleyici |
| İzlenebilirlik özeti (QR) | Herkes (tüketici dahil), fiyat bilgisi olmadan |

## Güvenilirlik ve sahtecilik koruması

Kaynak belgeler (fatura, teslim kaydı gibi) dijital imza ve kriptografik bir
özet ("parmak izi") ile korunur. Bir belge sonradan değiştirilirse veya aynı
belge ikinci kez kullanılmaya çalışılırsa sistem bunu fark edip reddeder.
Böylece hem "bu belge gerçekten bu kuruma mı ait" hem de "bu belge
değiştirilmiş mi" sorularının yanıtı doğrulanabilir hale gelir.

Şu an kurumsal veri kaynakları (ÇKS, e-Fatura, HKS, U-ETDS) gerçek devlet
sistemlerine bağlı değildir; bu pilotta imzalı ve gerçekçi biçimde üretilmiş
örnek (simüle) verilerdir. Ancak bu verilerin blok zincirine işlenmesi,
doğrulanması ve saklanması tamamen gerçek ve yerel olarak çalışan bir
Hyperledger Fabric ağı üzerinde gerçekleşir.

## Kullanılan teknolojiler

- **Hyperledger Fabric 2.5.15** — dört kuruluşun ortak deftere bağlandığı
  blok zinciri ağı
- **Go** — zincir üzerinde çalışan akıllı sözleşme (chaincode); yetki,
  miktar ve fiyat kurallarını uygular
- **Spring Boot / Java** — kurumlarla ve blok zinciriyle konuşan arka uç
  servisi
- **Basit web arayüzü** — aktörler, denetçi ve tüketici için ayrı ekranlar

## Proje şu an nerede?

Proje, sekiz aşamalı bir mühendislik sürecinde tamamlandı: teknik
tasarımdan başlayıp blok zinciri ağının kurulmasına, akıllı sözleşmenin
yazılmasına, gizlilik ve doğrulama katmanlarına, arka uç ve arayüze, son
olarak da test ve yarışma paketine kadar uzanan bir yol izlendi. 113 farklı
otomatik test ile normal ve şüpheli senaryolar, yetkisiz erişim denemeleri
ve sahte/değiştirilmiş belge durumları doğrulandı. Kurulumdan gösterime
kadar tüm adımlar temiz bir ortamda iki kez tekrarlanarak test edildi.
Sunum için 12 slaytlık bir gösterim ve internete ya da canlı ağa ihtiyaç
duymayan tek dosyalık bir yedek gösterim de hazırlandı.

## Bu proje ne değildir (dürüst sınırlar)

AgroChain, gerçekleri olduğu gibi anlatmaya özen gösteren bir pilot
projedir:

- Gerçek bir devlet kurumu entegrasyonu içermez; kurumsal veriler şu an
  için simüledir.
- Ulusal ölçekte bir sistem değildir — tek bir yerel test ağı üzerinde
  çalışan bir pilottur.
- Ürettiği "fiyat artışı sinyali" hukuki bir karar ya da kesin bir hile
  tespiti değildir; yalnızca incelemeye değer bir işarettir.
- Ödeme, cüzdan veya kripto para birimi içermez.
- Fiziksel ürünün gerçekten iddia edilen miktar/kalitede olduğunu garanti
  etmez — bunu doğrulamak ayrı bir saha denetimi gerektirir.

## Özetle

AgroChain; tarımsal tedarik zincirinde farklı kurumların ortak ve
değiştirilemez bir kayıt üzerinde buluşmasını, ticari verilerin gizli
kalmasını ve olağandışı fiyat hareketlerinin otomatik olarak fark
edilmesini aynı anda sağlayan, uçtan uca çalışan ve test edilmiş bir blok
zinciri pilotudur.
