package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.math.BigDecimal;

/** Reconciles committed reports; the ledger, not local arithmetic, attests results. */
public final class Analysis {
    static ObjectNode score(long purchase,long retail,long threshold){
        if(purchase<=0 || purchase>1_000_000_000L || retail<0 || retail>1_000_000_000L)throw ApiError.of("INVALID_MONEY");
        if(threshold<0 || threshold>100_000)throw ApiError.of("POLICY_MISMATCH");
        long n=(retail-purchase)*10000;boolean signal=n>purchase*threshold;
        return Json.obj("increaseBps",Math.floorDiv(n,purchase),"classification",signal?"REVIEW_REQUIRED":"NO_SIGNAL","reasonCode",signal?"ABOVE_PILOT_THRESHOLD":"WITHIN_PILOT_THRESHOLD");
    }
    static long configuredThreshold(String value){
        try{long n=new BigDecimal(value).movePointRight(2).longValueExact();if(n<0 || n>100000)throw new ArithmeticException();return n;}
        catch(RuntimeException e){throw ApiError.of("POLICY_MISMATCH");}
    }
    static ObjectNode proposal(Ledger ledger,Actor actor,String batch){
        JsonNode config=ledger.query(actor,"GetConfiguration"),purchase=ledger.query(actor,"GetPurchase",batch).path("body"),report=ledger.query(actor,"GetRetailReport",batch),freight=ledger.query(actor,"GetFreightCost",batch);
        long threshold=Json.number(config,"thresholdBps");
        if(threshold!=configuredThreshold(System.getenv().getOrDefault("ANOMALY_PRICE_INCREASE_THRESHOLD_PERCENT","50")) || !Json.text(report,"policyId").equals(Json.text(config,"policyId")))throw ApiError.of("POLICY_MISMATCH");
        for(JsonNode n:java.util.List.of(purchase,report,freight)){
            EvidenceService.equal(n,"currency","TRY");EvidenceService.equal(n,"taxBasis","EXCLUDING_TAX");
            if(Json.number(n,"quantityGrams")!=Json.number(report,"quantityGrams"))throw ApiError.of("SOURCE_BINDING_MISMATCH");
        }
        if(Json.number(freight,"totalKurus")<0)throw ApiError.of("INVALID_MONEY");
        return Json.obj("proposedResult",score(Json.number(purchase,"priceKurusPerKg"),Json.number(report,"offeredPriceKurusPerKg"),threshold),"recordSaltHex",Crypto.salt());
    }
    public static void reconcile(Store store,Ledger ledger,Workflow workflow){
        Actor oracle=new Actor("regulator","oracle");
        for(var row:store.rows("SELECT batch FROM projections WHERE lot IS NOT NULL")){
            String batch=(String)row.get("batch");
            try{
                if(!ledger.query(oracle,"GetAnomaly",batch).path("status").asText().equals("EVALUATION_PENDING"))continue;
                JsonNode b=ledger.query(oracle,"GetBatch",batch),report=ledger.query(oracle,"GetRetailReport",batch);
                String suffix=Crypto.digest((batch+":"+b.path("version").asText()).getBytes(java.nio.charset.StandardCharsets.UTF_8)).substring(0,24).toUpperCase();
                String op="OP-"+suffix;
                JsonNode request=Json.obj("scenario","NORMAL","privateInput",Json.object(),"command",Json.obj("schemaVersion","agrochain.command.v1","command","EvaluatePrice","operationId",op,"batchId",batch,"expectedVersion",b.get("version"),"payload",Json.obj("anomalyId","ANM-"+suffix,"reportId",report.get("reportId"))));
                workflow.submit(oracle,"EvaluatePrice",batch,op,request);
            }catch(ApiError e){org.slf4j.LoggerFactory.getLogger(Analysis.class).warn("Evaluation pending batch={} code={}",batch,e.code);}
        }
    }
}
