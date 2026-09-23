package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.*;
import java.util.*;

public final class EvidenceService {
    private final Ledger ledger;
    private final Store store;
    private final InstitutionalAdapter adapter;
    public EvidenceService(Ledger ledger,Store store,InstitutionalAdapter adapter){this.ledger=ledger;this.store=store;this.adapter=adapter;}
    public ObjectNode prepare(Actor actor,JsonNode request) {
        JsonNode command=request.get("command");String fn=Json.text(command,"command");
        if(fn.equals("EvaluatePrice"))return Analysis.proposal(ledger,actor,Json.text(command,"batchId"));
        if(fn.equals("OpenReview") || fn.equals("ResolveReview")){
            ObjectNode action=request.get("privateInput").deepCopy();action.put("recordSaltHex",Crypto.salt());
            return Json.obj("reviewActionInput",action,"reviewStateSaltHex",Crypto.salt());
        }
        JsonNode batch=fn.equals("CreateBatch")?command.get("payload"):ledger.query(actor,"GetBatch",Json.text(command,"batchId"));
        JsonNode trust=ledger.query(actor,"GetConfiguration");
        List<String[]> needs=switch(fn){
            case "CreateBatch" -> List.<String[]>of(new String[]{"CKS","PRODUCER_ELIGIBILITY"});
            case "OfferPickup" -> List.of(new String[]{"EFATURA","PURCHASE_INVOICE"},new String[]{"HKS","TRADE_NOTIFICATION"},new String[]{"UETDS","TRANSPORT_MANIFEST"});
            case "RecordFreightCost" -> List.<String[]>of(new String[]{"EFATURA","FREIGHT_INVOICE"});
            default -> List.of();
        };
        var documents=new ArrayList<JsonNode>();
        for(String[] need:needs){
            var doc=adapter.fetchEvidence(new InstitutionalAdapter.Request(need[0],need[1],Json.text(request,"scenario"),command,batch));
            Crypto.verify(doc.bundle(),trust,Json.text(command,"batchId"),Json.text(command,"operationId"),doc.original());
            validateBody(doc.bundle(),batch,need[0],need[1]);
            store.archive(actor,Json.text(command,"operationId"),doc.bundle(),doc.original());
            documents.add(doc.bundle());
        }
        ObjectNode transientData=Json.object();
        if(!documents.isEmpty()){
            documents.sort(Comparator.comparing(d->d.path("envelope").path("header").path("documentId").asText()));
            transientData.set("evidence",Json.MAPPER.valueToTree(documents));
        }
        if(fn.equals("RecordFreightCost"))transientData.put("recordSaltHex",Crypto.salt());
        if(fn.equals("ReportRetailPrice")){
            ObjectNode price=request.get("privateInput").deepCopy();price.put("recordSaltHex",Crypto.salt());transientData.set("retailReportInput",price);
        }
        return transientData;
    }
    static void equal(JsonNode body,String field,String expected){if(!body.path(field).asText().equals(expected))throw ApiError.of("SOURCE_BINDING_MISMATCH");}
    static void validateBody(JsonNode doc,JsonNode batch,String system,String type) {
        JsonNode h=doc.path("envelope").path("header"), b=doc.get("body");equal(h,"sourceSystem",system);equal(h,"documentType",type);
        if(Json.number(b,"quantityGrams")!=Json.number(batch,"quantityGrams"))throw ApiError.of("SOURCE_BINDING_MISMATCH");
        switch(type){
            case "PRODUCER_ELIGIBILITY" -> {
                Json.fields(b,"producerMsp","productCode","gradeCode","quantityGrams","originRegionCode","harvestDate","eligible");
                equal(b,"producerMsp","ProducerMSP");
                for(String field:List.of("productCode","gradeCode","originRegionCode","harvestDate"))equal(b,field,Json.text(batch,field));
                if(!b.get("eligible").isBoolean() || !b.get("eligible").booleanValue())throw ApiError.of("SOURCE_CLAIM_REJECTED");
            }
            case "PURCHASE_INVOICE", "FREIGHT_INVOICE" -> {
                boolean purchase=type.equals("PURCHASE_INVOICE");
                if(purchase){Json.fields(b,"sellerMsp","buyerMsp","quantityGrams","priceKurusPerKg","totalKurus","currency","taxBasis","attachmentDigest","attachmentMediaType");equal(b,"sellerMsp","ProducerMSP");equal(b,"buyerMsp","RetailerMSP");}
                else {Json.fields(b,"carrierMsp","payerMsp","quantityGrams","totalKurus","currency","taxBasis","attachmentDigest","attachmentMediaType");equal(b,"carrierMsp","LogisticsMSP");equal(b,"payerMsp","RetailerMSP");}
                equal(b,"currency","TRY");equal(b,"taxBasis","EXCLUDING_TAX");
                long total=Json.number(b,"totalKurus");if(total<0 || total>10_000_000_000L)throw ApiError.of("INVALID_MONEY");
                if(purchase){long price=Json.number(b,"priceKurusPerKg");if(price<=0 || price>100_000_000L || Json.number(b,"quantityGrams")*price!=total*1000)throw ApiError.of("INVALID_MONEY");}
                if(!Set.of("application/xml","application/pdf","application/json").contains(Json.text(b,"attachmentMediaType")))throw ApiError.of("INVALID_SCHEMA");
            }
            case "TRADE_NOTIFICATION" -> {
                Json.fields(b,"producerMsp","retailerMsp","productCode","quantityGrams","notified");equal(b,"producerMsp","ProducerMSP");equal(b,"retailerMsp","RetailerMSP");equal(b,"productCode",Json.text(batch,"productCode"));
                if(!b.get("notified").isBoolean() || !b.get("notified").booleanValue())throw ApiError.of("SOURCE_CLAIM_REJECTED");
            }
            case "TRANSPORT_MANIFEST" -> {
                Json.fields(b,"carrierMsp","consignorMsp","consigneeMsp","productCode","quantityGrams","declaredDepartureAt");equal(b,"carrierMsp","LogisticsMSP");equal(b,"consignorMsp","ProducerMSP");equal(b,"consigneeMsp","RetailerMSP");equal(b,"productCode",Json.text(batch,"productCode"));
                try{java.time.Instant.parse(Json.text(b,"declaredDepartureAt"));}catch(Exception e){throw ApiError.of("INVALID_SCHEMA");}
            }
            default -> throw ApiError.of("INVALID_SCHEMA");
        }
    }
}
