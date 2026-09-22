package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import java.nio.file.Path;
import java.util.*;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;
import org.junit.jupiter.api.io.TempDir;
import static org.junit.jupiter.api.Assertions.*;

/** Explicitly opt-in: uses real identities, endorsements and commit validation. */
@EnabledIfEnvironmentVariable(named="AGROCHAIN_LIVE_TESTS", matches="true")
class GatewayRecoveryTest {
    @TempDir Path data;

    @Test void restoredGatewayChecksValidAndInvalidCommitStatus() throws Exception {
        Settings s=Settings.environment();Actor actor=new Actor("producer","producer");
        Ledger.Prepared first,second;
        try(Store store=new Store(data); FabricLedger ledger=new FabricLedger(s)) {
            var evidence=new EvidenceService(ledger,store,new Simulators(s,store,Set.of()));
            ObjectNode a=BackendTest.request(),b=a.deepCopy();
            ((ObjectNode)b.get("command")).put("operationId",Crypto.id("OP"));
            first=endorse(ledger,evidence,actor,a);second=endorse(ledger,evidence,actor,b);
            assertEquals("CreateBatch",ledger.submit(actor,first).receipt().path("command").asText());
            assertEquals("VERSION_CONFLICT",assertThrows(ApiError.class,()->ledger.submit(actor,second)).code);
        }
        // A fresh Gateway reconstructs status requests solely from durable tx bytes.
        try(FabricLedger restored=new FabricLedger(s)) {
            assertEquals(first.txId(),restored.reconcile(actor,first).receipt().path("txId").asText());
            assertEquals("VERSION_CONFLICT",assertThrows(ApiError.class,()->restored.reconcile(actor,second)).code);
            assertEquals("UNSUPPORTED_PILOT_OPERATION",assertThrows(ApiError.class,()->restored.query(actor,"CreateBatch")).code);
            assertEquals("UNSUPPORTED_PILOT_OPERATION",assertThrows(ApiError.class,()->restored.endorse(actor,Json.obj("command","GetPurchase"),Map.of())).code);
        }
    }
    private Ledger.Prepared endorse(Ledger ledger,EvidenceService evidence,Actor actor,JsonNode request) {
        Map<String,byte[]> transientData=new TreeMap<>();
        evidence.prepare(actor,request).fields().forEachRemaining(e->transientData.put(e.getKey(),Json.bytes(e.getValue())));
        return ledger.endorse(actor,request.get("command"),transientData);
    }
}
