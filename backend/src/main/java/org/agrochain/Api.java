package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;
import java.nio.file.Files;
import java.util.*;

@RestController
@RequestMapping("/api/v1")
public final class Api {
    private final Ledger ledger;
    private final Store store;
    private final Workflow workflow;
    private final Projection projection;
    public Api(Ledger ledger,Store store,Workflow workflow,Projection projection){this.ledger=ledger;this.store=store;this.workflow=workflow;this.projection=projection;}
    static Actor actor(HttpServletRequest req){return (Actor)req.getAttribute("actor");}
    static JsonNode body(HttpServletRequest req){
        try{byte[] bytes=req.getInputStream().readNBytes(8*1024*1024+1);if(bytes.length>8*1024*1024)throw new ApiError("INVALID_SCHEMA",413);return Json.parse(bytes);}
        catch(java.io.IOException e){throw ApiError.of("INVALID_SCHEMA");}
    }
    @GetMapping("/health") public Object health(){return Map.of("status","UP","sourceMode","SIMULATED","blockchainMode","FABRIC","projectionAsOfBlock",Long.toString(store.checkpoint()));}
    @GetMapping("/session") public Object session(HttpServletRequest req){return actor(req);}
    @GetMapping("/batches") public Object batches(HttpServletRequest req){
        return store.rows("SELECT batch FROM projections ORDER BY batch").stream().map(r->ledger.query(actor(req),"GetBatch",(String)r.get("batch"))).toList();
    }
    @GetMapping("/batches/{id}/anomaly") public JsonNode anomaly(@PathVariable String id,HttpServletRequest req){return ledger.query(actor(req),"GetAnomaly",Json.id(id,"BAT"));}
    @GetMapping("/batches/{id}/reviews") public JsonNode reviews(@PathVariable String id,HttpServletRequest req){return ledger.query(actor(req),"GetReviewHistory",Json.id(id,"BAT"));}
    @PostMapping("/anomalies/{id}/review-actions") public ResponseEntity<JsonNode> review(@PathVariable String id,HttpServletRequest req){
        JsonNode input=body(req), c=input.path("command");Json.id(id,"ANM");
        if(!id.equals(c.path("payload").path("anomalyId").asText()) || !Set.of("OpenReview","ResolveReview").contains(c.path("command").asText()))throw ApiError.of("INVALID_SCHEMA");
        var result=workflow.submit(actor(req),Json.text(c,"command"),Json.text(c,"batchId"),req.getHeader("Idempotency-Key"),input);
        return ResponseEntity.status(result.http()).body(result.body());
    }
    @PostMapping("/batches") public ResponseEntity<JsonNode> create(HttpServletRequest req){return mutate(req,"CreateBatch",null);}
    @PostMapping("/batches/{id}/{action}") public ResponseEntity<JsonNode> command(@PathVariable String id,@PathVariable String action,HttpServletRequest req){
        Json.id(id,"BAT");String fn=switch(action){
            case "pickup-offers"->"OfferPickup";case "pickup-acceptances"->"AcceptPickup";case "freight-costs"->"RecordFreightCost";
            case "delivery-offers"->"OfferDelivery";case "delivery-acceptances"->"AcceptDelivery";case "retail-reports"->"ReportRetailPrice";
            case "anomaly-evaluations"->"EvaluatePrice";
            default->throw ApiError.of("UNSUPPORTED_PILOT_OPERATION");};
        return mutate(req,fn,id);
    }
    private ResponseEntity<JsonNode> mutate(HttpServletRequest req,String fn,String batch){
        var result=workflow.submit(actor(req),fn,batch,req.getHeader("Idempotency-Key"),body(req));
        return ResponseEntity.status(result.http()).header("X-Source-Mode","SIMULATED").body(result.body());
    }
    @GetMapping("/operations/{id}") public JsonNode status(@PathVariable String id,HttpServletRequest req){return workflow.status(actor(req),id);}
    @GetMapping("/batches/{id}") public JsonNode batch(@PathVariable String id,HttpServletRequest req){return ledger.query(actor(req),"GetBatch",Json.id(id,"BAT"));}
    @GetMapping("/batches/{id}/history") public JsonNode history(@PathVariable String id,@RequestParam(defaultValue="") String bookmark,HttpServletRequest req){
        return ledger.query(actor(req),"GetBatchHistory",Json.canonical(Json.obj("schemaVersion","agrochain.page-request.v1","filter",Json.id(id,"BAT"),"pageSize",100,"bookmark",bookmark)));
    }
    @GetMapping("/batches/{id}/commercial") public JsonNode commercial(@PathVariable String id,@RequestParam String kind,HttpServletRequest req){
        String fn=switch(kind){case "purchase"->"GetPurchase";case "freight"->"GetFreightCost";case "retail"->"GetRetailReport";default->throw ApiError.of("INVALID_SCHEMA");};
        if(!actor(req).commercial(kind))throw ApiError.of("PRIVATE_DATA_ACCESS_DENIED");
        return ledger.query(actor(req),fn,Json.id(id,"BAT"));
    }
    private Map<String,Object> document(String id,Actor actor){
        Json.id(id,"DOC");var row=store.row("SELECT * FROM evidence WHERE id=? AND committed=1",id);
        if(row==null)throw ApiError.of("DOCUMENT_NOT_FOUND");
        String kind=row.get("kind").equals("FREIGHT_INVOICE")?"freight":"purchase";
        if(!actor.commercial(kind))throw ApiError.of("PRIVATE_DATA_ACCESS_DENIED");return row;
    }
    @GetMapping("/documents/{id}") public ResponseEntity<byte[]> download(@PathVariable String id,HttpServletRequest req) throws Exception {
        var row=document(id,actor(req));if(row.get("attachment")==null)throw ApiError.of("DOCUMENT_NOT_FOUND");
        // Filename is assigned internally from a validated document ID, never a request path.
        return ResponseEntity.ok().contentType(MediaType.APPLICATION_OCTET_STREAM).header("X-Source-Mode","SIMULATED")
            .body(Files.readAllBytes(store.documents.resolve((String)row.get("attachment"))));
    }
    @PostMapping("/documents/{id}/verify") public JsonNode verify(@PathVariable String id,HttpServletRequest req){
        var row=document(id,actor(req));JsonNode input=body(req);Json.fields(input,"originalBase64");
        byte[] bytes;try{bytes=Base64.getDecoder().decode(Json.text(input,"originalBase64"));}catch(IllegalArgumentException e){throw ApiError.of("INVALID_SCHEMA");}
        JsonNode cached=Json.parse((String)row.get("bundle"));JsonNode envelope=ledger.query(actor(req),"GetDocument",id);
        JsonNode bundle=Json.obj("envelope",envelope,"body",cached.get("body"),"saltHex",cached.get("saltHex"));JsonNode h=envelope.get("header");
        if(row.get("attachment")==null)throw ApiError.of("UNSUPPORTED_PILOT_OPERATION");
        Crypto.verify(bundle,ledger.query(actor(req),"GetConfiguration"),Json.text(h,"batchId"),Json.text(h,"boundOperationId"),bytes);
        var receipt=store.row("SELECT receipt FROM operations WHERE actor=? AND id=?",row.get("actor"),row.get("operation"));
        return Json.obj("schemaVersion","agrochain.integrity-result.v1","documentId",id,"commitmentMatched",true,"signatureVerified",true,"attachmentMatched",true,"sourceMode","SIMULATED","verifiedAt",Crypto.now(),"committedTxId",Json.parse((String)receipt.get("receipt")).get("txId"));
    }
    @GetMapping("/public/lots/{id}") public ResponseEntity<JsonNode> publicLot(@PathVariable String id){
        return ResponseEntity.ok().header("X-Projection-Checkpoint",Long.toString(store.checkpoint())).header("X-Projection-Consistency","EVENTUAL").body(projection.lot(id));
    }
    @ExceptionHandler(ApiError.class) public ResponseEntity<Object> error(ApiError e){return ResponseEntity.status(e.status).body(e.body());}
    @ExceptionHandler(org.springframework.web.bind.MissingServletRequestParameterException.class)
    public ResponseEntity<Object> missingParameter(Exception e){return ResponseEntity.badRequest().body(ApiError.of("INVALID_SCHEMA").body());}
    @ExceptionHandler(org.springframework.web.HttpRequestMethodNotSupportedException.class)
    public ResponseEntity<Object> wrongMethod(Exception e){return ResponseEntity.status(405).body(new ApiError("METHOD_NOT_ALLOWED",405).body());}
    @ExceptionHandler(Exception.class) public ResponseEntity<Object> unexpected(Exception e){
        // Log exception types/locations only; never messages that may contain inputs.
        Throwable cause=e;for(int i=0;i<4 && cause.getCause()!=null;i++)cause=cause.getCause();
        org.slf4j.LoggerFactory.getLogger(Api.class).error("Unhandled type={} location={}",cause.getClass().getName(),cause.getStackTrace().length==0?"unknown":cause.getStackTrace()[0]);
        return ResponseEntity.status(500).body(ApiError.of("INTERNAL_ERROR").body());
    }
}
