package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.security.*;
import java.security.spec.*;
import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.*;

public final class Crypto {
    private static final SecureRandom RANDOM=new SecureRandom();
    public static byte[] random(int size) { byte[] b=new byte[size]; RANDOM.nextBytes(b); return b; }
    public static String salt() { return HexFormat.of().formatHex(random(32)); }
    public static String id(String prefix) { return prefix+"-"+HexFormat.of().formatHex(random(16)).toUpperCase(); }
    public static String now() { return java.time.format.DateTimeFormatter.ofPattern("uuuu-MM-dd'T'HH:mm:ss.SSS'Z'").withZone(java.time.ZoneOffset.UTC).format(Instant.now().truncatedTo(ChronoUnit.MILLIS)); }
    public static String digest(byte[] bytes) { try { return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(bytes)); } catch(Exception e) { throw new IllegalStateException(e); } }
    public static String commitment(JsonNode body, String salt) {
        if (!salt.matches("[0-9a-f]{64}")) throw ApiError.of("INVALID_SCHEMA");
        byte[] b=Json.bytes(body), domain="AgroChain/document/v1\0".getBytes(StandardCharsets.UTF_8);
        return digest(ByteBuffer.allocate(domain.length+32+8+b.length).put(domain).put(HexFormat.of().parseHex(salt)).putLong(b.length).put(b).array());
    }
    public static byte[] message(JsonNode h) { return ("AgroChain/source-signature/v1\0"+Json.canonical(h)).getBytes(StandardCharsets.UTF_8); }
    public static String sign(JsonNode h, Path path) {
        try {
            String pem=Files.readString(path).replace("-----BEGIN PRIVATE KEY-----","").replace("-----END PRIVATE KEY-----","").replaceAll("\\s","");
            PrivateKey key=KeyFactory.getInstance("Ed25519").generatePrivate(new PKCS8EncodedKeySpec(Base64.getDecoder().decode(pem)));
            Signature s=Signature.getInstance("Ed25519"); s.initSign(key); s.update(message(h)); return Base64.getUrlEncoder().withoutPadding().encodeToString(s.sign());
        } catch(Exception e) { throw ApiError.of("SIMULATOR_UNAVAILABLE"); }
    }
    public static void verify(JsonNode bundle, JsonNode trust, String batch, String operation, byte[] attachment) {
        Json.fields(bundle,"envelope","body","saltHex");
        JsonNode env=bundle.get("envelope"); Json.fields(env,"header","signature"); JsonNode h=env.get("header");
        Json.fields(h,"schemaVersion","documentId","sourceSystem","sourceMode","issuerId","keyId","sourceDocumentId","documentType","batchId","boundOperationId","issuedAt","nonce","commitmentAlgorithm","commitment","signatureAlgorithm");
        if (!h.path("sourceMode").asText().equals("SIMULATED") || !h.path("schemaVersion").asText().equals("agrochain.source-document.v1") || !h.path("signatureAlgorithm").asText().equals("Ed25519") || !h.path("commitmentAlgorithm").asText().equals("SHA256_SALTED_JCS_V1")) throw ApiError.of("INVALID_SCHEMA");
        Json.id(Json.text(h,"documentId"),"DOC");Json.id(Json.text(h,"sourceDocumentId"),"DOC");Json.id(Json.text(h,"batchId"),"BAT");Json.id(Json.text(h,"boundOperationId"),"OP");
        for(String key:List.of("nonce","commitment"))if(!Json.text(h,key).matches("[0-9a-f]{64}"))throw ApiError.of("INVALID_SCHEMA");
        for(String key:List.of("issuerId","keyId"))if(!Json.text(h,key).matches("[A-Z0-9_]{1,64}"))throw ApiError.of("INVALID_SCHEMA");
        String issued=Json.text(h,"issuedAt");
        try{if(!issued.matches("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\\.[0-9]{3}Z"))throw new IllegalArgumentException();Instant.parse(issued);}catch(Exception e){throw ApiError.of("INVALID_SCHEMA");}
        JsonNode source=null;
        for (JsonNode candidate:trust.path("sources")) {
            boolean type=false; for(JsonNode t:candidate.path("documentTypes")) type |= t.equals(h.get("documentType"));
            if(candidate.path("enabled").asBoolean() && candidate.path("issuerId").equals(h.get("issuerId")) && candidate.path("keyId").equals(h.get("keyId")) && candidate.path("sourceSystem").equals(h.get("sourceSystem")) && type) source=candidate;
        }
        if(source==null) throw ApiError.of("UNTRUSTED_SOURCE_KEY");
        try {
            byte[] raw=rawUrl(Json.text(source,"publicKeyRawBase64url"),32);
            byte[] prefix=HexFormat.of().parseHex("302a300506032b6570032100");
            var pub=KeyFactory.getInstance("Ed25519").generatePublic(new X509EncodedKeySpec(ByteBuffer.allocate(prefix.length+raw.length).put(prefix).put(raw).array()));
            Signature verifier=Signature.getInstance("Ed25519");verifier.initVerify(pub);verifier.update(message(h));
            if(!verifier.verify(rawUrl(Json.text(env,"signature"),64))) throw ApiError.of("INVALID_SOURCE_SIGNATURE");
        } catch(ApiError e) { throw e; } catch(Exception e) { throw ApiError.of("INVALID_SOURCE_SIGNATURE"); }
        if(!h.path("batchId").asText().equals(batch) || !h.path("boundOperationId").asText().equals(operation)) throw ApiError.of("SOURCE_BINDING_MISMATCH");
        if(!commitment(bundle.get("body"),Json.text(bundle,"saltHex")).equals(Json.text(h,"commitment"))) throw ApiError.of("DOCUMENT_HASH_MISMATCH");
        if(bundle.get("body").has("attachmentDigest") && (attachment==null || attachment.length>5*1024*1024 || !digest(attachment).equals(bundle.get("body").path("attachmentDigest").asText()))) throw ApiError.of("DOCUMENT_HASH_MISMATCH");
    }
    private static byte[] rawUrl(String value,int size) {
        byte[] bytes=Base64.getUrlDecoder().decode(value);
        if(bytes.length!=size || !Base64.getUrlEncoder().withoutPadding().encodeToString(bytes).equals(value))throw new IllegalArgumentException();
        return bytes;
    }
}
