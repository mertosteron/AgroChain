package org.agrochain;

import com.google.zxing.*;
import com.google.zxing.qrcode.QRCodeWriter;
import org.springframework.web.bind.annotation.*;
import org.springframework.http.*;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import javax.imageio.ImageIO;
import java.net.URI;

@RestController
public final class LotQr {
    private final Projection projection;
    public LotQr(Projection projection){this.projection=projection;}
    static String link(String base,String lot){
        URI uri=URI.create(base);
        if(!java.util.Set.of("http","https").contains(uri.getScheme()) || uri.getHost()==null || uri.getUserInfo()!=null || uri.getQuery()!=null || uri.getFragment()!=null || !(uri.getPath().isEmpty() || uri.getPath().equals("/")))throw new IllegalArgumentException("Invalid public base URL");
        return base.replaceAll("/$","")+"/consumer.html?lot="+Json.id(lot,"LOT");
    }
    static byte[] render(String url) throws Exception {
        var matrix=new QRCodeWriter().encode(url,BarcodeFormat.QR_CODE,320,320);
        var image=new BufferedImage(320,320,BufferedImage.TYPE_INT_RGB);
        for(int y=0;y<320;y++)for(int x=0;x<320;x++)image.setRGB(x,y,matrix.get(x,y)?0x000000:0xffffff);
        var out=new ByteArrayOutputStream();ImageIO.write(image,"png",out);return out.toByteArray();
    }
    @GetMapping(value="/api/v1/public/lots/{id}/qr",produces=MediaType.IMAGE_PNG_VALUE)
    public ResponseEntity<byte[]> qr(@PathVariable String id){
        projection.lot(id);
        try{return ResponseEntity.ok().body(render(link(System.getenv().getOrDefault("AGROCHAIN_PUBLIC_BASE_URL","http://localhost:8080"),id)));}
        catch(Exception e){throw ApiError.of("INTERNAL_ERROR");}
    }
    @ExceptionHandler(ApiError.class) public ResponseEntity<Object> error(ApiError e){return ResponseEntity.status(e.status).contentType(MediaType.APPLICATION_JSON).body(e.body());}
}
