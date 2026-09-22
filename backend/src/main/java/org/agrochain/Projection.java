package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.*;
import java.util.*;

/** Explicit public DTO built only from valid committed events and immutable batch fields. */
public final class Projection {
    private final Store store;
    private final Ledger ledger;
    public Projection(Store store,Ledger ledger){this.store=store;this.ledger=ledger;}
    public synchronized void accept(Ledger.Event event) {
        if(store.seen(event.txId()))return;
        JsonNode e=event.body();String batch=Json.id(Json.text(e,"batchId"),"BAT");String type=Json.text(e,"eventType");
        if(!Set.of("BatchCreated","PickupOffered","PickupAccepted","FreightCostRecorded","DeliveryOffered","DeliveryAccepted","RetailPriceReported").contains(type))throw ApiError.of("INVALID_SCHEMA");
        var existing=store.row("SELECT body FROM projections WHERE batch=?",batch);
        ObjectNode body;
        if(existing==null){
            JsonNode b=ledger.query(new Actor("regulator","public-reader"),"GetBatch",batch);
            body=Json.obj("schemaVersion","agrochain.public-trace.v1","batchId",batch);
            for(String key:List.of("productCode","gradeCode","quantityGrams","originRegionCode","harvestDate"))body.set(key,b.get(key));
            body.set("organizations",Json.MAPPER.valueToTree(List.of(Json.obj("role","PRODUCER","displayName","Producer"),Json.obj("role","LOGISTICS","displayName","Logistics"),Json.obj("role","RETAILER","displayName","Retailer"))));
            body.set("history",Json.MAPPER.createArrayNode());body.set("sources",Json.MAPPER.createArrayNode());
        }else body=(ObjectNode)Json.parse((String)existing.get("body"));
        body.set("state",e.get("newState"));body.set("version",e.get("resultVersion"));
        body.put("updatedTxId",event.txId());body.put("asOfBlock",Long.toString(event.block()));
        if(type.equals("BatchCreated") || type.equals("PickupAccepted") || type.equals("DeliveryAccepted")){
            ObjectNode history=Json.obj("eventType",type,"recordedAt",e.get("txTime"),"txId",event.txId(),"toRole",type.equals("BatchCreated")?"PRODUCER":type.equals("PickupAccepted")?"LOGISTICS":"RETAILER");
            if(!type.equals("BatchCreated"))history.put("fromRole",type.equals("PickupAccepted")?"PRODUCER":"LOGISTICS");
            ((ArrayNode)body.get("history")).add(history);
        }
        List<String> sources=type.equals("BatchCreated")?List.of("CKS"):type.equals("PickupOffered")?List.of("EFATURA","HKS","UETDS"):List.of();
        for(String source:sources)((ArrayNode)body.get("sources")).add(Json.obj("sourceSystem",source,"sourceMode","SIMULATED","verificationStatus","SIGNED_SIMULATOR_ASSERTION"));
        if(type.equals("RetailPriceReported"))for(JsonNode id:e.path("objectIds"))if(id.asText().startsWith("LOT-"))body.set("retailLotId",id);
        store.project(event.txId(),event.block(),batch,body.has("retailLotId")?body.get("retailLotId").asText():null,body);
    }
    public JsonNode lot(String id){Json.id(id,"LOT");var row=store.row("SELECT body FROM projections WHERE lot=?",id);if(row==null)throw ApiError.of("BATCH_NOT_FOUND");return Json.parse((String)row.get("body"));}
}
