package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.nio.charset.StandardCharsets;
import java.util.*;

/** Four in-process simulator adapters share storage, not verifier trust. */
public final class Simulators implements InstitutionalAdapter {
    private final Settings settings;
    private final Store store;
    private final Set<String> disabled;
    public Simulators(Settings settings,Store store,Set<String> disabled){this.settings=settings;this.store=store;this.disabled=Set.copyOf(disabled);}
    public synchronized Document fetchEvidence(Request r) {
        if(disabled.contains(r.sourceSystem()))throw ApiError.of("SIMULATOR_UNAVAILABLE");
        if(!Set.of("NORMAL","SUSPICIOUS").contains(r.scenario()))throw ApiError.of("INVALID_SCHEMA");
        String key=r.sourceSystem()+":"+r.documentType()+":"+Json.text(r.command(),"batchId")+":"+Json.text(r.command(),"operationId");
        String request=Json.canonical(Json.obj("scenario",r.scenario(),"command",r.command(),"batch",r.batch()));
        var old=store.row("SELECT * FROM simulator WHERE key=?",key);
        if(old!=null){if(!request.equals(old.get("request")))throw ApiError.of("SOURCE_BINDING_MISMATCH");return new Document(Json.parse((String)old.get("bundle")),(byte[])old.get("attachment"));}
        JsonNode b=r.batch();long quantity=Json.number(b,"quantityGrams");
        // These lookup fixtures deliberately support only the frozen tomato pilot.
        if(quantity!=100000 || !b.path("productCode").asText().equals("TOMATO") || !b.path("gradeCode").asText().equals("STANDARD") || !b.path("originRegionCode").asText().equals("07"))throw ApiError.of("SOURCE_CLAIM_REJECTED");
        ObjectNode body;
        byte[] original=null;
        switch(r.sourceSystem()+":"+r.documentType()) {
            case "CKS:PRODUCER_ELIGIBILITY" -> body=Json.obj("producerMsp","ProducerMSP","productCode",b.get("productCode"),"gradeCode",b.get("gradeCode"),"quantityGrams",quantity,"originRegionCode",b.get("originRegionCode"),"harvestDate",b.get("harvestDate"),"eligible",true);
            case "EFATURA:PURCHASE_INVOICE" -> {
                original=("<invoice mode=\"SIMULATED\" kind=\"PURCHASE\" batch=\""+Json.text(r.command(),"batchId")+"\">200000</invoice>").getBytes(StandardCharsets.UTF_8);
                body=Json.obj("sellerMsp","ProducerMSP","buyerMsp","RetailerMSP","quantityGrams",quantity,"priceKurusPerKg",2000,"totalKurus",200000,"currency","TRY","taxBasis","EXCLUDING_TAX","attachmentDigest",Crypto.digest(original),"attachmentMediaType","application/xml");
            }
            case "EFATURA:FREIGHT_INVOICE" -> {
                original=("<invoice mode=\"SIMULATED\" kind=\"FREIGHT\" batch=\""+Json.text(r.command(),"batchId")+"\">20000</invoice>").getBytes(StandardCharsets.UTF_8);
                body=Json.obj("carrierMsp","LogisticsMSP","payerMsp","RetailerMSP","quantityGrams",quantity,"totalKurus",20000,"currency","TRY","taxBasis","EXCLUDING_TAX","attachmentDigest",Crypto.digest(original),"attachmentMediaType","application/xml");
            }
            case "HKS:TRADE_NOTIFICATION" -> body=Json.obj("producerMsp","ProducerMSP","retailerMsp","RetailerMSP","productCode",b.get("productCode"),"quantityGrams",quantity,"notified",true);
            case "UETDS:TRANSPORT_MANIFEST" -> body=Json.obj("carrierMsp","LogisticsMSP","consignorMsp","ProducerMSP","consigneeMsp","RetailerMSP","productCode",b.get("productCode"),"quantityGrams",quantity,"declaredDepartureAt",Crypto.now());
            default -> throw ApiError.of("INVALID_SCHEMA");
        }
        String salt=Crypto.salt();
        JsonNode header=Json.obj("schemaVersion","agrochain.source-document.v1","documentId",Crypto.id("DOC"),"sourceSystem",r.sourceSystem(),"sourceMode","SIMULATED",
            "issuerId","SIM_"+r.sourceSystem(),"keyId","SIM_"+r.sourceSystem()+"_KEY01","sourceDocumentId",Crypto.id("DOC"),"documentType",r.documentType(),
            "batchId",r.command().get("batchId"),"boundOperationId",r.command().get("operationId"),"issuedAt",Crypto.now(),"nonce",Crypto.salt(),
            "commitmentAlgorithm","SHA256_SALTED_JCS_V1","commitment",Crypto.commitment(body,salt),"signatureAlgorithm","Ed25519");
        JsonNode bundle=Json.obj("envelope",Json.obj("header",header,"signature",Crypto.sign(header,settings.sourceKeys().resolve(r.sourceSystem()+".pem"))),"body",body,"saltHex",salt);
        store.update("INSERT INTO simulator(key,request,bundle,attachment) VALUES(?,?,?,?)",key,request,Json.canonical(bundle),original);
        return new Document(bundle,original);
    }
    public boolean verifyOriginal(String id,byte[] bytes) {
        Json.id(id,"DOC");
        var row=store.row("SELECT attachment FROM simulator WHERE json_extract(bundle,'$.envelope.header.documentId')=?",id);
        if(row==null || row.get("attachment")==null)throw ApiError.of("DOCUMENT_NOT_FOUND");
        if(!java.security.MessageDigest.isEqual((byte[])row.get("attachment"),bytes))throw ApiError.of("DOCUMENT_HASH_MISMATCH");
        return true;
    }
}
