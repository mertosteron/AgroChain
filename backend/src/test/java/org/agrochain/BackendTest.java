package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.junit.jupiter.api.*;
import org.junit.jupiter.api.io.TempDir;
import org.springframework.mock.web.*;
import java.nio.file.*;
import java.security.*;
import java.util.*;
import static org.junit.jupiter.api.Assertions.*;

class BackendTest {
    @TempDir Path dir;
    Store store;
    Settings settings;
    ObjectNode trust;
    final Actor producer=new Actor("producer","producer");
    @BeforeEach void setup() throws Exception {
        settings=new Settings(dir,dir.resolve("data"),dir.resolve("tokens.json"),dir.resolve("keys"),"agrochannel","agrochain");
        Files.createDirectories(settings.sourceKeys());store=new Store(settings.data());
        var sources=Json.MAPPER.createArrayNode();
        for(String sys:List.of("CKS","EFATURA","HKS","UETDS")){
            var key=KeyPairGenerator.getInstance("Ed25519").generateKeyPair();
            Files.writeString(settings.sourceKeys().resolve(sys+".pem"),"-----BEGIN PRIVATE KEY-----\n"+Base64.getEncoder().encodeToString(key.getPrivate().getEncoded())+"\n-----END PRIVATE KEY-----\n");
            byte[] encoded=key.getPublic().getEncoded();byte[] raw=Arrays.copyOfRange(encoded,encoded.length-32,encoded.length);
            List<String> types=switch(sys){case "CKS"->List.of("PRODUCER_ELIGIBILITY");case "EFATURA"->List.of("FREIGHT_INVOICE","PURCHASE_INVOICE");case "HKS"->List.of("TRADE_NOTIFICATION");default->List.of("TRANSPORT_MANIFEST");};
            sources.add(Json.obj("issuerId","SIM_"+sys,"keyId","SIM_"+sys+"_KEY01","sourceSystem",sys,"enabled",true,"documentTypes",types,"publicKeyRawBase64url",Base64.getUrlEncoder().withoutPadding().encodeToString(raw)));
        }
        trust=Json.obj("sources",sources);
    }
    @AfterEach void close() throws Exception {store.close();}
    static ObjectNode batch(){return Json.obj("productCode","TOMATO","gradeCode","STANDARD","quantityGrams",100000,"originRegionCode","07","harvestDate","2026-09-15","logisticsMsp","LogisticsMSP","intendedRetailerMsp","RetailerMSP");}
    static ObjectNode request(){return Json.obj("command",Json.obj("schemaVersion","agrochain.command.v1","operationId",Crypto.id("OP"),"batchId",Crypto.id("BAT"),"expectedVersion",0,"command","CreateBatch","payload",batch()),"scenario","NORMAL","privateInput",Json.object());}
    Simulators simulator(){return new Simulators(settings,store,Set.of());}
    InstitutionalAdapter.Request source(JsonNode r,String sys,String type){return new InstitutionalAdapter.Request(sys,type,"NORMAL",r.get("command"),batch());}
    static void code(String code,org.junit.jupiter.api.function.Executable action){assertEquals(code,assertThrows(ApiError.class,action).code);}

    @Test void allFourAdaptersAndBothInvoiceTypesVerifyAndRetryStably() {
        var r=request();var sim=simulator();
        for(String[] pair:new String[][]{{"CKS","PRODUCER_ELIGIBILITY"},{"EFATURA","PURCHASE_INVOICE"},{"EFATURA","FREIGHT_INVOICE"},{"HKS","TRADE_NOTIFICATION"},{"UETDS","TRANSPORT_MANIFEST"}}){
            var query=source(r,pair[0],pair[1]);var doc=sim.fetchEvidence(query);var again=sim.fetchEvidence(query);
            assertArrayEquals(Json.bytes(doc.bundle()),Json.bytes(again.bundle()));assertArrayEquals(doc.original(),again.original());
            Crypto.verify(doc.bundle(),trust,r.path("command").path("batchId").asText(),r.path("command").path("operationId").asText(),doc.original());
            EvidenceService.validateBody(doc.bundle(),batch(),pair[0],pair[1]);
            assertEquals("SIMULATED",doc.bundle().path("envelope").path("header").path("sourceMode").asText());
            if(doc.original()!=null){assertTrue(sim.verifyOriginal(doc.bundle().path("envelope").path("header").path("documentId").asText(),doc.original()));}
        }
    }
    @Test void rejectsModifiedBodySaltHeaderSignatureKeyAndOriginal() {
        var r=request();var doc=simulator().fetchEvidence(source(r,"EFATURA","PURCHASE_INVOICE"));
        String batch=r.path("command").path("batchId").asText(), op=r.path("command").path("operationId").asText();
        for(String target:List.of("body","salt","header","signature","key","original")){
            ObjectNode changed=doc.bundle().deepCopy();byte[] original=doc.original().clone();String expected;
            switch(target){
                case "body"->{((ObjectNode)changed.get("body")).put("totalKurus",1);expected="DOCUMENT_HASH_MISMATCH";}
                case "salt"->{changed.put("saltHex",Crypto.salt());expected="DOCUMENT_HASH_MISMATCH";}
                case "header"->{((ObjectNode)changed.path("envelope").path("header")).put("batchId",Crypto.id("BAT"));expected="INVALID_SOURCE_SIGNATURE";}
                case "signature"->{((ObjectNode)changed.get("envelope")).put("signature",Base64.getUrlEncoder().withoutPadding().encodeToString(new byte[64]));expected="INVALID_SOURCE_SIGNATURE";}
                case "key"->{((ObjectNode)changed.path("envelope").path("header")).put("keyId","UNKNOWN");expected="UNTRUSTED_SOURCE_KEY";}
                default->{original[0]^=1;expected="DOCUMENT_HASH_MISMATCH";}
            }
            code(expected,()->Crypto.verify(changed,trust,batch,op,original));
        }
        code("SOURCE_BINDING_MISMATCH",()->Crypto.verify(doc.bundle(),trust,Crypto.id("BAT"),op,doc.original()));
    }
    @Test void simulatorOutageNeverCreatesUnsignedEvidence() {
        for(String[] pair:new String[][]{{"CKS","PRODUCER_ELIGIBILITY"},{"EFATURA","PURCHASE_INVOICE"},{"HKS","TRADE_NOTIFICATION"},{"UETDS","TRANSPORT_MANIFEST"}}){
            var sim=new Simulators(settings,store,Set.of(pair[0]));
            code("SIMULATOR_UNAVAILABLE",()->sim.fetchEvidence(source(request(),pair[0],pair[1])));
        }
        assertTrue(store.rows("SELECT * FROM simulator").isEmpty());
    }
    @Test void bodyValidationRejectsSignedFalseClaimsAndWrongQuantities() {
        var r=request();var doc=simulator().fetchEvidence(source(r,"CKS","PRODUCER_ELIGIBILITY"));ObjectNode bundle=doc.bundle().deepCopy();
        ((ObjectNode)bundle.get("body")).put("eligible",false);
        code("SOURCE_CLAIM_REJECTED",()->EvidenceService.validateBody(bundle,batch(),"CKS","PRODUCER_ELIGIBILITY"));
        ((ObjectNode)bundle.get("body")).put("quantityGrams",1);
        code("SOURCE_BINDING_MISMATCH",()->EvidenceService.validateBody(bundle,batch(),"CKS","PRODUCER_ELIGIBILITY"));
    }
    @Test void strictSchemasDoNotCoerceMoneyOrAcceptActorHeaders() throws Exception {
        code("INVALID_SCHEMA",()->Json.parse("{\"a\":1,\"a\":2}"));
        code("INVALID_SCHEMA",()->Json.parse("{\"a\":-0}"));
        code("INVALID_SCHEMA",()->Json.parse("{} {}"));
        code("INVALID_SCHEMA",()->Json.canonical(Json.parse("{\"price\":1.1}")));
        var r=request();((ObjectNode)r.path("command").path("payload")).put("quantityGrams","100000");
        code("INVALID_SCHEMA",()->Workflow.validate(r,"CreateBatch",null,r.path("command").path("operationId").asText()));
        var tokens=Json.obj("producer:producer",Crypto.salt(),"logistics:carrier",Crypto.salt(),"retailer:retailer",Crypto.salt(),"regulator:auditor",Crypto.salt());
        Files.writeString(settings.tokens(),Json.encode(tokens));AuthFilter filter=new AuthFilter(settings);
        MockHttpServletRequest req=new MockHttpServletRequest("POST","/api/v1/batches");req.addHeader("X-MSP","ProducerMSP");req.addHeader("X-Role","producer");
        MockHttpServletResponse response=new MockHttpServletResponse();filter.doFilter(req,response,new MockFilterChain());assertEquals(401,response.getStatus());
        req=new MockHttpServletRequest("POST","/api/v1/batches");req.addHeader("Authorization","Bearer "+tokens.path("producer:producer").asText());req.addHeader("X-Role","admin");response=new MockHttpServletResponse();filter.doFilter(req,response,new MockFilterChain());assertEquals(producer,req.getAttribute("actor"));
    }
    class FakeLedger implements Ledger {
        Map<String,JsonNode> receipts=new HashMap<>();int endorsements,submissions;boolean loseResponse,dropBeforeCommit,offline,invalid,invalidAfterLostResponse;
        public JsonNode query(Actor actor,String fn,String... args){
            if(offline)throw ApiError.of("FABRIC_UNAVAILABLE");
            return switch(fn){case "GetConfiguration"->trust;case "GetBatch"->batch();case "GetOperation"->{var receipt=receipts.get(args[0]);if(receipt==null)throw ApiError.of("OPERATION_NOT_FOUND");yield receipt;}default->throw new IllegalArgumentException(fn);};
        }
        public Prepared endorse(Actor actor,JsonNode command,Map<String,byte[]> trans){endorsements++;assertTrue(trans.containsKey("evidence"));return new Prepared(Json.bytes(command),Crypto.salt());}
        public Committed submit(Actor actor,Prepared prepared){
            submissions++;if(dropBeforeCommit)throw new UnknownCommit();if(invalid)throw ApiError.of("VERSION_CONFLICT");JsonNode c=Json.parse(prepared.bytes());
            var receipt=Json.obj("schemaVersion","agrochain.receipt.v1","operationId",c.get("operationId"),"batchId",c.get("batchId"),"command",c.get("command"),"txId",prepared.txId(),"resultVersion",1);
            receipts.put(c.path("operationId").asText(),receipt);if(loseResponse)throw new UnknownCommit();return new Committed(receipt,1);
        }
        public void events(long start,java.util.function.Consumer<Event> c){}
        public Committed reconcile(Actor actor,Prepared p){if(invalidAfterLostResponse)throw ApiError.of("VERSION_CONFLICT");return null;}
    }
    Workflow workflow(FakeLedger l){return new Workflow(store,l,new EvidenceService(l,store,simulator()));}
    @Test void timeoutAfterCommitRecoversReceiptWithoutSecondSubmitAcrossRestart() throws Exception {
        var l=new FakeLedger();l.loseResponse=true;var w=workflow(l);var request=request();String op=request.path("command").path("operationId").asText();
        assertEquals(202,w.submit(producer,"CreateBatch",null,op,request).http());assertEquals("SUBMITTED_UNKNOWN",w.status(producer,op).path("status").asText());
        store.close();store=new Store(settings.data());w=workflow(l);w.recover();
        assertEquals("COMMITTED",w.status(producer,op).path("status").asText());assertEquals(1,l.submissions);assertEquals(1,l.endorsements);
        assertEquals(1,((Number)store.row("SELECT COUNT(*) AS n FROM evidence WHERE committed=1").get("n")).intValue());
        assertEquals(200,w.submit(producer,"CreateBatch",null,op,request).http());assertEquals(1,l.submissions);
    }
    @Test void crashBeforeCommitRetriesExactTransactionAndSameEvidence() throws Exception {
        var l=new FakeLedger();l.dropBeforeCommit=true;var w=workflow(l);var request=request();String op=request.path("command").path("operationId").asText();
        assertEquals(202,w.submit(producer,"CreateBatch",null,op,request).http());String txid=(String)store.operation(producer,op).get("txid");
        store.close();store=new Store(settings.data());w=workflow(l);l.dropBeforeCommit=false;w.recover();
        assertEquals(txid,w.status(producer,op).path("receipt").path("txId").asText());assertEquals(1,l.endorsements);assertEquals(2,l.submissions);
    }
    @Test void duplicateConflictAndWrongRoleDoNotSubmit() {
        var l=new FakeLedger();var w=workflow(l);var r=request();String op=r.path("command").path("operationId").asText();
        assertEquals(201,w.submit(producer,"CreateBatch",null,op,r).http());
        r.put("scenario","SUSPICIOUS");code("IDEMPOTENCY_CONFLICT",()->w.submit(producer,"CreateBatch",null,op,r));
        code("UNAUTHORIZED_ORGANIZATION",()->w.submit(new Actor("retailer","retailer"),"CreateBatch",null,op,r));
        code("OPERATION_NOT_FOUND",()->w.status(new Actor("retailer","retailer"),op));assertEquals(1,l.submissions);
    }
    @Test void unavailableFabricAndInvalidCommitAreNeverSuccess() {
        var l=new FakeLedger();var w=workflow(l);var r=request();String op=r.path("command").path("operationId").asText();
        l.offline=true;code("FABRIC_UNAVAILABLE",()->w.submit(producer,"CreateBatch",null,op,r));assertEquals(0,l.submissions);
        l.offline=false;l.invalid=true;code("VERSION_CONFLICT",()->w.submit(producer,"CreateBatch",null,op,r));assertEquals("FAILED",w.status(producer,op).path("status").asText());
    }
    @Test void invalidCommitAfterLostResponseIsReconciledWithoutResubmission() throws Exception {
        var l=new FakeLedger();l.dropBeforeCommit=true;var w=workflow(l);var r=request();String op=r.path("command").path("operationId").asText();
        assertEquals(202,w.submit(producer,"CreateBatch",null,op,r).http());
        store.close();store=new Store(settings.data());l.invalidAfterLostResponse=true;w=workflow(l);w.recover();
        assertEquals("FAILED",w.status(producer,op).path("status").asText());
        assertEquals("VERSION_CONFLICT",w.status(producer,op).path("error").path("code").asText());
        assertEquals(1,l.submissions);assertEquals(1,l.endorsements);
    }
    @Test void fabricProtobufRuntimeMatchesGeneratedClasses() throws Exception {
        var id=org.hyperledger.fabric.protos.msp.SerializedIdentity.newBuilder().setMspid("ProducerMSP").build();
        var req=org.hyperledger.fabric.protos.gateway.CommitStatusRequest.newBuilder().setChannelId("agrochannel").setTransactionId("test").setIdentity(id.toByteString()).build();
        assertEquals("ProducerMSP",org.hyperledger.fabric.protos.msp.SerializedIdentity.parseFrom(req.getIdentity()).getMspid());
    }
    @Test void untrustedAdapterEvidenceNeverReachesEndorsement() {
        var l=new FakeLedger();var actual=simulator();
        InstitutionalAdapter corrupt=new InstitutionalAdapter(){
            public Document fetchEvidence(Request r){
                var d=actual.fetchEvidence(r);ObjectNode changed=d.bundle().deepCopy();
                ((ObjectNode)changed.get("body")).put("eligible",false);
                return new Document(changed,d.original());
            }
            public boolean verifyOriginal(String id,byte[] bytes){return false;}
        };
        var w=new Workflow(store,l,new EvidenceService(l,store,corrupt));var r=request();
        code("DOCUMENT_HASH_MISMATCH",()->w.submit(producer,"CreateBatch",null,r.path("command").path("operationId").asText(),r));
        assertEquals(0,l.endorsements);assertEquals(0,l.submissions);assertTrue(store.rows("SELECT * FROM evidence").isEmpty());
    }
    @Test void duplicateAuthenticationTokensFailStartup() throws Exception {
        String token=Crypto.salt();
        Files.writeString(settings.tokens(),Json.encode(Json.obj("producer:producer",token,"logistics:carrier",token,"retailer:retailer",Crypto.salt(),"regulator:auditor",Crypto.salt())));
        assertThrows(IllegalArgumentException.class,()->new AuthFilter(settings));
    }
    @Test void projectionDeduplicatesEventsAndExcludesPrivateData() {
        var ledger=new FakeLedger();var p=new Projection(store,ledger);String batch=Crypto.id("BAT"),lot=Crypto.id("LOT");
        var event=new Ledger.Event(Crypto.salt(),9,Json.obj("batchId",batch,"eventType","RetailPriceReported","newState","RETAIL_REPORTED","resultVersion",7,"objectIds",List.of(batch,lot),"txTime",Json.obj("seconds","123","nanos",0),"offeredPriceKurusPerKg",123456));
        p.accept(event);p.accept(event);JsonNode dto=p.lot(lot);assertFalse(dto.has("offeredPriceKurusPerKg"));assertEquals("9",dto.path("asOfBlock").asText());
        assertEquals(1,store.rows("SELECT * FROM events").size());assertEquals(9,store.checkpoint());
    }
    @Test void errorEnvelopeNeverContainsInternalDetails() {
        var error=ApiError.of("FABRIC_UNAVAILABLE").body();assertEquals(Set.of("schemaVersion","code","message","correlationId","retryable"),error.keySet());
        assertEquals(503,ApiError.of("FABRIC_UNAVAILABLE").status);assertEquals(422,ApiError.of("DOCUMENT_HASH_MISMATCH").status);
    }
}
