package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.util.*;

/** Durable stage -> endorse -> persist transaction -> submit -> VALID -> promote. */
public final class Workflow {
    private final Store store;
    private final Ledger ledger;
    private final EvidenceService evidence;
    public record Result(int http,JsonNode body) {}
    public Workflow(Store store,Ledger ledger,EvidenceService evidence){this.store=store;this.ledger=ledger;this.evidence=evidence;}
    static void validate(JsonNode request,String fn,String batch,String key) {
        Json.fields(request,"command","scenario","privateInput");
        if(!Set.of("NORMAL","SUSPICIOUS").contains(Json.text(request,"scenario")))throw ApiError.of("INVALID_SCHEMA");
        JsonNode c=request.get("command");
        Json.fields(c,"schemaVersion","operationId","batchId","expectedVersion","command","payload");
        Json.id(Json.text(c,"operationId"),"OP");Json.id(Json.text(c,"batchId"),"BAT");
        if(!Json.text(c,"operationId").equals(key) || !Json.text(c,"command").equals(fn) || (batch!=null && !Json.text(c,"batchId").equals(batch)))throw ApiError.of("INVALID_SCHEMA");
        if(!Json.text(c,"schemaVersion").equals("agrochain.command.v1"))throw ApiError.of("UNSUPPORTED_SCHEMA_VERSION");
        long version=Json.number(c,"expectedVersion");if(version<0 || version>2147483646)throw ApiError.of("INVALID_SCHEMA");
        JsonNode p=c.get("payload");
        switch(fn){
            case "EvaluatePrice" -> {Json.fields(p,"anomalyId","reportId");Json.id(Json.text(p,"anomalyId"),"ANM");Json.id(Json.text(p,"reportId"),"RPT");}
            case "OpenReview","ResolveReview" -> {Json.fields(p,"anomalyId");Json.id(Json.text(p,"anomalyId"),"ANM");}
            case "CreateBatch" -> {
                Json.fields(p,"productCode","gradeCode","quantityGrams","originRegionCode","harvestDate","logisticsMsp","intendedRetailerMsp");
                if(!Json.text(p,"logisticsMsp").equals("LogisticsMSP") || !Json.text(p,"intendedRetailerMsp").equals("RetailerMSP"))throw ApiError.of("UNAUTHORIZED_ORGANIZATION");
            }
            case "OfferPickup","AcceptPickup","OfferDelivery","AcceptDelivery" -> {Json.fields(p,"transferId","quantityGrams");Json.id(Json.text(p,"transferId"),"TRF");}
            case "RecordFreightCost" -> {Json.fields(p,"costId");Json.id(Json.text(p,"costId"),"CST");}
            case "ReportRetailPrice" -> {
                Json.fields(p,"reportId","retailLotId","policyId");Json.id(Json.text(p,"reportId"),"RPT");Json.id(Json.text(p,"retailLotId"),"LOT");Json.id(Json.text(p,"policyId"),"CFG");
            }
            default -> throw ApiError.of("UNSUPPORTED_PILOT_OPERATION");
        }
        if(p.has("quantityGrams") && (Json.number(p,"quantityGrams")<1 || Json.number(p,"quantityGrams")>100000000))throw ApiError.of("INVALID_QUANTITY");
        JsonNode privateInput=request.get("privateInput");
        if(fn.equals("ReportRetailPrice")){
            Json.fields(privateInput,"offeredPriceKurusPerKg","currency","taxBasis","reportedAt");
            if(Json.number(privateInput,"offeredPriceKurusPerKg")<0 || Json.number(privateInput,"offeredPriceKurusPerKg")>100000000)throw ApiError.of("INVALID_MONEY");
            EvidenceService.equal(privateInput,"currency","TRY");EvidenceService.equal(privateInput,"taxBasis","EXCLUDING_TAX");
        }else if(fn.equals("OpenReview") || fn.equals("ResolveReview")){
            if(fn.equals("OpenReview"))Json.fields(privateInput,"reviewerRef");else Json.fields(privateInput,"reviewerRef","outcome","explanation");
            Json.id(Json.text(privateInput,"reviewerRef"),"REV");
            if(fn.equals("ResolveReview") && (!Set.of("EXPLAINED","FOLLOW_UP_RECOMMENDED","INSUFFICIENT_EVIDENCE").contains(Json.text(privateInput,"outcome")) || Json.text(privateInput,"explanation").isBlank() || Json.text(privateInput,"explanation").length()>2000))throw ApiError.of("INVALID_SCHEMA");
        }else Json.fields(privateInput);
        if(Json.bytes(request).length>65536)throw ApiError.of("INVALID_SCHEMA");
    }
    public synchronized Result submit(Actor actor,String fn,String batch,String key,JsonNode request) {
        actor.authorize(fn);validate(request,fn,batch,key);
        boolean existing=store.operation(actor,key)!=null;
        var row=store.stage(actor,key,request);
        if("FAILED".equals(row.get("state"))){
            ApiError error=ApiError.of((String)row.get("error"));
            boolean correctedPolicy=fn.equals("EvaluatePrice") && error.code.equals("POLICY_MISMATCH") && row.get("tx")==null;
            if(error.status!=503 && !correctedPolicy)throw error;
            store.update("UPDATE operations SET state='PENDING',error=NULL WHERE actor=? AND id=?",actor.key(),key);
        }
        advance(actor,key);
        row=store.operation(actor,key);
        if("FAILED".equals(row.get("state")))throw ApiError.of((String)row.get("error"));
        return new Result("COMMITTED".equals(row.get("state"))?(existing?200:201):202,status(actor,key));
    }
    private void advance(Actor actor,String id) {
        var row=store.operation(actor,id);
        if(row==null || Set.of("COMMITTED","FAILED").contains(row.get("state")))return;
        JsonNode request=Json.parse((String)row.get("request")), command=request.get("command");
        try {
            try {
                JsonNode receipt=ledger.query(actor,"GetOperation",id);
                // Operation IDs are scoped by MSP on chain. Never adopt another role's
                // or an independently committed command's receipt without matching it.
                if(!receipt.path("batchId").equals(command.get("batchId")) || !receipt.path("command").equals(command.get("command")))throw ApiError.of("IDEMPOTENCY_CONFLICT");
                if(row.get("txid")!=null && !receipt.path("txId").asText().equals(row.get("txid")))throw ApiError.of("IDEMPOTENCY_CONFLICT");
                if(row.get("txid")==null)throw ApiError.of("IDEMPOTENCY_CONFLICT");
                store.committed(actor,id,receipt);return;
            }catch(ApiError e){if(!e.code.equals("OPERATION_NOT_FOUND"))throw e;}
            Ledger.Prepared prepared;
            if(row.get("tx")!=null){
                prepared=new Ledger.Prepared((byte[])row.get("tx"),(String)row.get("txid"));
                Ledger.Committed known=ledger.reconcile(actor,prepared);
                if(known!=null){store.committed(actor,id,known.receipt());return;}
            }
            else {
                JsonNode transientJson;
                if(row.get("prepared")!=null)transientJson=Json.parse((String)row.get("prepared"));
                else {
                    transientJson=evidence.prepare(actor,request);
                    store.update("UPDATE operations SET prepared=? WHERE actor=? AND id=?",Json.canonical(transientJson),actor.key(),id);
                }
                Map<String,byte[]> transientData=new TreeMap<>();transientJson.fields().forEachRemaining(e->transientData.put(e.getKey(),Json.bytes(e.getValue())));
                prepared=ledger.endorse(actor,command,transientData);
                store.update("UPDATE operations SET tx=?,txid=? WHERE actor=? AND id=?",prepared.bytes(),prepared.txId(),actor.key(),id);
            }
            // Persist before sending; a crash here has a safe, same-tx retry path.
            store.update("UPDATE operations SET state='SUBMITTED_UNKNOWN' WHERE actor=? AND id=?",actor.key(),id);
            Ledger.Committed committed=ledger.submit(actor,prepared);
            store.committed(actor,id,committed.receipt());
            org.slf4j.LoggerFactory.getLogger(Workflow.class).info("Committed operation={} batch={} organization={} command={} tx={}",id,command.path("batchId").asText(),actor.msp(),command.path("command").asText(),prepared.txId());
        }catch(Ledger.UnknownCommit ignored){/* Keep the durable pending outcome; never invent a fresh operation. */}
        catch(ApiError e){
            // A transport outage after transaction persistence cannot prove failure.
            var current=store.operation(actor,id);
            if(current.get("tx")!=null && (e.status==503 || e.code.equals("INTERNAL_ERROR")))return;
            store.update("UPDATE operations SET state='FAILED',error=? WHERE actor=? AND id=?",e.code,actor.key(),id);
        }
    }
    public synchronized void recover() {
        for(var row:store.rows("SELECT actor,id FROM operations WHERE state IN ('PENDING','SUBMITTED_UNKNOWN')"))advance(Actor.parse((String)row.get("actor")),(String)row.get("id"));
    }
    public JsonNode status(Actor actor,String id) {
        Json.id(id,"OP");var row=store.operation(actor,id);if(row==null)throw ApiError.of("OPERATION_NOT_FOUND");
        ObjectNode out=Json.obj("schemaVersion","agrochain.operation-status.v1","operationId",id,"status",row.get("state"),"statusUrl","/api/v1/operations/"+id);
        if(row.get("txid")!=null)out.put("txId",(String)row.get("txid"));
        if("COMMITTED".equals(row.get("state")))out.set("receipt",Json.parse((String)row.get("receipt")));
        if("FAILED".equals(row.get("state")))out.set("error",Json.MAPPER.valueToTree(ApiError.of((String)row.get("error")).body()));
        return out;
    }
}
