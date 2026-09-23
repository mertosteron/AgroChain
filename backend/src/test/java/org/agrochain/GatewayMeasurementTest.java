package org.agrochain;

import java.nio.file.*;
import java.util.*;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;
import org.junit.jupiter.api.io.TempDir;
import static org.junit.jupiter.api.Assertions.*;

/** Timings exclude endorsement; submit() waits for actual Fabric VALID status. */
@EnabledIfEnvironmentVariable(named="AGROCHAIN_MEASURE", matches="true")
class GatewayMeasurementTest {
    @TempDir Path data;

    @Test void sequentialSubmitAndQueries() throws Exception {
        Settings settings=Settings.environment();Actor producer=new Actor("producer","producer");
        var output=settings.root().resolve("network/runtime/stage7/gateway-measurements.json");
        Files.createDirectories(output.getParent());
        var report=Json.obj("result","FAIL","concurrency",1,"clock","System.nanoTime",
            "warmupSamplesExcluded",1,"scope","submitAsync through VALID commit status; separate evaluate queries",
            "failures",0);
        var samples=Json.MAPPER.createArrayNode();report.set("samples",samples);
        try(Store store=new Store(data);FabricLedger ledger=new FabricLedger(settings)) {
            String privateBatch=Json.parse(Files.readString(settings.root().resolve("network/runtime/backend-acceptance/stage6-summary.json"))).get(0).path("batch").asText();
            var evidence=new EvidenceService(ledger,store,new Simulators(settings,store,Set.of()));
            for(int i=0;i<31;i++) {
                var request=BackendTest.request();var command=request.get("command");
                Map<String,byte[]> transientData=new TreeMap<>();
                evidence.prepare(producer,request).fields().forEachRemaining(e->transientData.put(e.getKey(),Json.bytes(e.getValue())));
                var tx=ledger.endorse(producer,command,transientData);
                long start=System.nanoTime();var committed=ledger.submit(producer,tx);long submit=System.nanoTime()-start;
                assertEquals(tx.txId(),committed.receipt().path("txId").asText());
                start=System.nanoTime();var batch=ledger.query(producer,"GetBatch",command.path("batchId").asText());long shared=System.nanoTime()-start;
                assertEquals("CREATED",batch.path("state").asText());
                start=System.nanoTime();var purchase=ledger.query(producer,"GetPurchase",privateBatch);long privateQuery=System.nanoTime()-start;
                assertEquals(2000,purchase.path("body").path("priceKurusPerKg").asLong());
                if(i>0)samples.add(Json.obj("submitValidMs",submit/1e6,"sharedQueryMs",shared/1e6,
                    "privateQueryMs",privateQuery/1e6,"endorsedTransactionBytes",tx.bytes().length,
                    "txId",tx.txId(),"blockNumber",committed.block(),
                    "sharedResponseBytes",Json.bytes(batch).length,"privateResponseBytes",Json.bytes(purchase).length));
            }
            report.put("result","PASS");
        } catch(Exception|AssertionError e) {report.put("failures",1);throw e;}
        finally {Files.writeString(output,Json.encode(report));}
    }
}
