# AgroChain yarışma paketi

TEKNOFEST teknik jürisine yönelik yerel pilot teslimi. Gerçek yerel Fabric
işlemleri ile **SIMULATED** kurum verileri açıkça ayrılır. Bu paket portal
başvurusu veya yarışma şartnamesine resmi uygunluk onayı değildir.

- [Kurulum ve tekrar başlatma](INSTALLATION.md)
- [Üç dakikalık canlı gösterim](DEMO_3MIN.md)
- [Kısa teknik açıklama ve mimari](TECHNICAL.md)
- [12 slaytlık düzenlenebilir sunum](AgroChain.pptx)
- [Çevrimdışı sunum ve kayıtlı ekranlar](offline.html)
- [Çevrimdışı geçiş ve video planı](OFFLINE.md)
- [Konuşmacı notları dahil sunum kaynağı](slides.json)
- [Aşama 7 ölçüm raporu](../STAGE7_REPORT.md)
- [Aşama 8 kabul raporu](../STAGE8_REPORT.md)

Sabit normal parti `BAT-DEMONORMAL01`, şüpheli parti `BAT-DEMOSUSPICIOUS01`.
Tüketici lotları sırasıyla `LOT-DEMONORMAL01` ve `LOT-DEMOSUSPICIOUS01`.
`make demo-seed demo-check` gerçek API/Fabric üzerinden yükler ve doğrular.
Yinelenen çalıştırma aynı operationId'leri kullanır. Anahtarlar, saltlar ve
kimlik dosyaları teslim paketinin parçası değildir.

`make competition-package` tüm kaynakları ve bu belgeleri Git dışı
`dist/AgroChain-competition.tar.gz` dosyasında toplar. Yanındaki `.sha256`
dosyasını `sha256sum -c AgroChain-competition.sha256` ile doğrulayın. Arşivdeki
`PACKAGE_MANIFEST.json` her dosyanın özetini içerir. Araç önbellekleri, özel
anahtarlar, tokenlar ve ledger verileri dahil değildir. Yeni bilgisayarda
kurulum rehberini izleyerek yeni geliştirme kimlikleri oluşturun.

`make stage8-rehearse` aynı makinede temiz bir kaynak kopyası, yeni kimlikler ve
boş geçici disklerle `demo-setup`, iki `demo-seed` ve `demo-check` çalıştırır.
Araç/bağımlılık önbelleklerini yeniden kullanır. Özgün ağ geçici durur ve sonunda
geri açılır; çalışan backend'i önceden kapatın, ardından `make backend-run` ile
yeniden açın. Bu komut ikinci fiziksel bilgisayarda test yapıldığı anlamına gelmez.

## Sunumu yeniden üretme

`slides.json` içeriği tek kaynaktır. `scripts/build-presentation.mjs` PPTX ve
slayt önizlemeleri, `scripts/build-offline.py` bağımsız HTML üretir.
PPTX üretimi Codex'in sağladığı `@oai/artifact-tool` çalışma zamanı gerektirir;
teslim edilmiş PPTX/HTML dosyalarını açmak için bu araç gerekmez.
HTML üretimi Python standart kütüphanesiyle yapılır.

Çevrimdışı HTML hiçbir API veya dış varlık istemez. İçindeki sentetik ekran
kayıtları son doğrulama anına aittir; canlı ağ durumu gibi yorumlanmamalıdır.
