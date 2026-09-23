package org.agrochain;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;
import com.google.zxing.*;
import com.google.zxing.common.HybridBinarizer;
import javax.imageio.ImageIO;
import java.io.ByteArrayInputStream;

class AnalysisTest {
 @Test void exactBoundaryAndFloor(){
  assertEquals("NO_SIGNAL",Analysis.score(2000,2800,5000).path("classification").asText());
  assertEquals("NO_SIGNAL",Analysis.score(2000,3000,5000).path("classification").asText());
  assertEquals("REVIEW_REQUIRED",Analysis.score(30000,45001,5000).path("classification").asText());
  assertEquals(5000,Analysis.score(30000,45001,5000).path("increaseBps").asLong());
  assertEquals(8500,Analysis.score(2000,3700,5000).path("increaseBps").asLong());
  assertEquals(-4,Analysis.score(3000,2999,5000).path("increaseBps").asLong());
  assertThrows(ApiError.class,()->Analysis.score(0,100,5000));
  assertThrows(ApiError.class,()->Analysis.score(1,1000000001,5000));
 }
 @Test void configurationIsExact(){assertEquals(5000,Analysis.configuredThreshold("50"));assertEquals(5012,Analysis.configuredThreshold("50.12"));assertThrows(ApiError.class,()->Analysis.configuredThreshold("50.001"));}
 @Test void rolesSeparateScoringAndReview(){
  assertThrows(ApiError.class,()->new Actor("regulator","auditor").authorize("EvaluatePrice"));
  assertThrows(ApiError.class,()->new Actor("regulator","oracle").authorize("OpenReview"));
  assertThrows(ApiError.class,()->new Actor("retailer","retailer").authorize("ResolveReview"));
  new Actor("regulator","reviewer").authorize("ResolveReview");
 }
 @Test void unicodeReviewAndLimit(){assertTrue(Json.canonical(Json.obj("explanation","Ürün incelendi.")).contains("Ürün"));assertThrows(ApiError.class,()->Json.canonical(Json.obj("explanation","x".repeat(2001))));}
 @Test void qrDecodesToPublicLink()throws Exception{
  String url=LotQr.link("http://localhost:8080","LOT-NORMAL01");var image=ImageIO.read(new ByteArrayInputStream(LotQr.render(url)));
  int[] pixels=image.getRGB(0,0,image.getWidth(),image.getHeight(),null,0,image.getWidth());
  String decoded=new MultiFormatReader().decode(new BinaryBitmap(new HybridBinarizer(new RGBLuminanceSource(image.getWidth(),image.getHeight(),pixels)))).getText();assertEquals(url,decoded);
  assertThrows(IllegalArgumentException.class,()->LotQr.link("javascript:alert(1)","LOT-NORMAL01"));
 }
}
