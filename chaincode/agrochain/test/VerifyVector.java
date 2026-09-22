import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.security.*;
import java.security.spec.X509EncodedKeySpec;
import java.util.*;

/** Dependency-free interoperability test; Java 17+. Restricted ASCII/integer JCS schema. */
class VerifyVector {
    static byte[] utf8(String s) { return s.getBytes(StandardCharsets.UTF_8); }
    static String canonical(Map<String, Object> value) {
        var entries = new ArrayList<String>();
        for (var e : new TreeMap<>(value).entrySet()) {
            Object v = e.getValue();
            String encoded;
            if (v instanceof String s) {
                if (!s.chars().allMatch(c -> c >= 32 && c <= 126)) throw new IllegalArgumentException();
                encoded = "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
            } else if (v instanceof Integer || v instanceof Long || v instanceof Boolean) encoded = v.toString();
            else throw new IllegalArgumentException();
            entries.add("\"" + e.getKey() + "\":" + encoded);
        }
        return "{" + String.join(",", entries) + "}";
    }
    static String commitment(String body, byte[] salt) throws Exception {
        var hash = MessageDigest.getInstance("SHA-256");
        hash.update(utf8("AgroChain/document/v1\0"));
        hash.update(salt);
        hash.update(ByteBuffer.allocate(8).putLong(utf8(body).length).array());
        return HexFormat.of().formatHex(hash.digest(utf8(body)));
    }
    static boolean verify(byte[] rawKey, String header, byte[] sig) throws Exception {
        byte[] prefix = HexFormat.of().parseHex("302a300506032b6570032100");
        byte[] der = ByteBuffer.allocate(prefix.length + 32).put(prefix).put(rawKey).array();
        var verifier = Signature.getInstance("Ed25519");
        verifier.initVerify(KeyFactory.getInstance("Ed25519").generatePublic(new X509EncodedKeySpec(der)));
        verifier.update(utf8("AgroChain/source-signature/v1\0"));
        verifier.update(utf8(header));
        return verifier.verify(sig);
    }
    static void check(boolean ok) { if (!ok) throw new AssertionError("interoperability vector failed"); }
    public static void main(String[] args) throws Exception {
        String digest = "a95bd76451b27eaa935bd2bac3586c94165d58a6622154395f9530cdfa0c5a56";
        var body = new HashMap<String, Object>(Map.of("sellerMsp", "ProducerMSP", "buyerMsp", "RetailerMSP",
            "quantityGrams", 100000, "priceKurusPerKg", 2000, "totalKurus", 200000, "currency", "TRY",
            "taxBasis", "EXCLUDING_TAX", "attachmentDigest", digest, "attachmentMediaType", "application/xml"));
        byte[] salt = HexFormat.of().parseHex("202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f");
        String expected = "4e1259fa1bf2ab64d37e9026c276fbe2dc06872f2deb800cbc87f83ffd15940a";
        check(utf8(canonical(body)).length == 288);
        check(commitment(canonical(body), salt).equals(expected));
        String[][] fields = {{"schemaVersion","agrochain.source-document.v1"},{"documentId","DOC-INVOICE01"},
            {"sourceSystem","EFATURA"},{"sourceMode","SIMULATED"},{"issuerId","SIM_EFATURA"},{"keyId","SIM_EFATURA_KEY01"},
            {"sourceDocumentId","DOC-SOURCE001"},{"documentType","PURCHASE_INVOICE"},{"batchId","BAT-NORMAL01"},
            {"boundOperationId","OP-OFFER001"},{"issuedAt","2026-09-16T09:00:00.000Z"},
            {"nonce","000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"},
            {"commitmentAlgorithm","SHA256_SALTED_JCS_V1"},{"commitment",expected},{"signatureAlgorithm","Ed25519"}};
        var header = new HashMap<String,Object>();
        for (var pair : fields) header.put(pair[0],pair[1]);
        byte[] key = Base64.getUrlDecoder().decode("njeHGEyz0EF6HP6iEDUHPftD5gwevFdrlCcOnJgTFkY");
        byte[] signature = Base64.getUrlDecoder().decode("2LPDyrKB3jnVnrDMXr4_I1A1qA8R6Ty-69K8PS67kYM5u2x63GIr1Pm5LETIRAfW_xRSng2yyJraa8Ac41_5DQ");
        check(verify(key, canonical(header), signature));
        header.put("batchId", "BAT-CHANGED1");
        check(!verify(key, canonical(header), signature));
        body.put("totalKurus",200001); check(!commitment(canonical(body),salt).equals(expected));
        salt[0] ^= 1; check(!commitment(canonical(body),salt).equals(expected));
        byte[] attachment = utf8("<invoice id=\"SIM-001\">200000</invoice>");
        check(HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(attachment)).equals(digest));
        check(!HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(utf8(new String(attachment, StandardCharsets.UTF_8)+"\n"))).equals(digest));
        check(canonical(Map.of("z",true,"a","<>&\\\"")).equals("{\"a\":\"<>&\\\\\\\"\",\"z\":true}"));
        System.out.println("PASS Java commitment, Ed25519 signature, canonical subset and tampered document vectors");
    }
}
