package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import io.grpc.ManagedChannel;
import io.grpc.Grpc;
import io.grpc.TlsChannelCredentials;
import org.hyperledger.fabric.client.*;
import org.hyperledger.fabric.client.identity.*;
import java.nio.file.*;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

/** No generic submit endpoint: private read functions cannot be ordered by this backend. */
public final class FabricLedger implements Ledger {
    private static final Set<String> READS=Set.of("Health","GetConfiguration","GetBatch","BatchExists","GetBatchHistory","GetTransfer","GetOperation","GetDocument","GetPurchase","GetFreightCost","GetRetailReport","GetAnomaly","GetReviewHistory");
    private final Settings settings;
    private final Map<Actor,Gateway> gateways=new HashMap<>();
    private final ManagedChannel channel;
    public FabricLedger(Settings settings) throws Exception {
        this.settings=settings;
        var ca=settings.root().resolve("network/organizations/peerOrganizations/retailer.agrochain.test/peers/peer0.retailer.agrochain.test/tls/ca.crt");
        var credentials=TlsChannelCredentials.newBuilder().trustManager(ca.toFile()).build();
        channel=Grpc.newChannelBuilder(System.getenv().getOrDefault("AGROCHAIN_GATEWAY_ENDPOINT","localhost:9051"),credentials)
            .overrideAuthority("peer0.retailer.agrochain.test").build();
    }
    private synchronized Gateway gateway(Actor actor) {
        return gateways.computeIfAbsent(actor,a->{
            var path=settings.root().resolve("network/runtime/identities/"+a.org()+"/"+a.role()+"/msp");
            try(var cert=Files.newBufferedReader(path.resolve("signcerts/cert.pem"));var key=Files.newBufferedReader(path.resolve("keystore/key.pem"))) {
                return Gateway.newInstance().identity(new X509Identity(a.msp(),Identities.readX509Certificate(cert)))
                    .signer(Signers.newPrivateKeySigner(Identities.readPrivateKey(key))).connection(channel)
                    .evaluateOptions(o->o.withDeadlineAfter(5,TimeUnit.SECONDS))
                    .endorseOptions(o->o.withDeadlineAfter(10,TimeUnit.SECONDS))
                    .submitOptions(o->o.withDeadlineAfter(5,TimeUnit.SECONDS))
                    .commitStatusOptions(o->o.withDeadlineAfter(15,TimeUnit.SECONDS)).connect();
            }catch(Exception e){throw ApiError.of("FABRIC_UNAVAILABLE");}
        });
    }
    private Contract contract(Actor actor){return gateway(actor).getNetwork(settings.channel()).getContract(settings.contract());}
    public JsonNode query(Actor actor,String fn,String... args) {
        if(!READS.contains(fn))throw ApiError.of("UNSUPPORTED_PILOT_OPERATION");
        try{return Json.parse(contract(actor).newProposal(fn).addArguments(args).setEndorsingOrganizations("RetailerMSP").build().evaluate());}
        catch(GatewayException e){throw mapped(e);}
    }
    public Prepared endorse(Actor actor,JsonNode command,Map<String,byte[]> transientData) {
        String fn=Json.text(command,"command");actor.authorize(fn);
        try {
            Transaction tx=contract(actor).newProposal(fn).addArguments(Json.canonical(command)).putAllTransient(transientData)
                .setEndorsingOrganizations("RetailerMSP","RegulatorMSP").build().endorse();
            return new Prepared(tx.getBytes(),tx.getTransactionId());
        }catch(GatewayException e){throw mapped(e);}
    }
    public Committed submit(Actor actor,Prepared prepared) {
        try {
            Transaction tx=gateway(actor).newTransaction(prepared.bytes());
            var commit=tx.submitAsync();var status=commit.getStatus();
            if(!status.isSuccessful())throw ApiError.of(status.getCode().getNumber()==11?"VERSION_CONFLICT":"FABRIC_TRANSACTION_INVALID");
            return new Committed(Json.parse(tx.getResult()),status.getBlockNumber());
        }catch(ApiError e){throw e;}catch(Exception e){throw new UnknownCommit();}
    }
    public Committed reconcile(Actor actor,Prepared prepared) {
        Gateway g=gateway(actor);
        var identity=org.hyperledger.fabric.protos.msp.SerializedIdentity.newBuilder()
            .setMspid(g.getIdentity().getMspId())
            .setIdBytes(com.google.protobuf.ByteString.copyFrom(g.getIdentity().getCredentials())).build();
        var request=org.hyperledger.fabric.protos.gateway.CommitStatusRequest.newBuilder()
            .setChannelId(settings.channel()).setTransactionId(prepared.txId()).setIdentity(identity.toByteString()).build();
        var signed=org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest.newBuilder().setRequest(request.toByteString()).build();
        try {
            Status status=g.newCommit(signed.toByteArray()).getStatus(o->o.withDeadlineAfter(2,TimeUnit.SECONDS));
            if(!status.isSuccessful())throw ApiError.of(status.getCode().getNumber()==11?"VERSION_CONFLICT":"FABRIC_TRANSACTION_INVALID");
            return new Committed(Json.parse(g.newTransaction(prepared.bytes()).getResult()),status.getBlockNumber());
        }catch(CommitStatusException e){return null;}
    }
    public void events(long startBlock,Consumer<Event> consume) {
        var network=gateway(new Actor("regulator","public-reader")).getNetwork(settings.channel());
        try(var events=network.newChaincodeEventsRequest(settings.contract()).startBlock(startBlock).build().getEvents()) {
            events.forEachRemaining(e->consume.accept(new Event(e.getTransactionId(),e.getBlockNumber(),Json.parse(e.getPayload()))));
        }
    }
    private ApiError mapped(GatewayException error) {
        // Only a domain code in structured peer details may leave this boundary.
        for(var detail:error.getDetails()) {
            String message=detail.getMessage();int start=message.indexOf('{');
            if(start>=0)try{
                JsonNode d=Json.parse(message.substring(start));String code=d.path("code").asText();
                if(d.path("schemaVersion").asText().equals("agrochain.error.v1") && code.matches("[A-Z_]{3,64}"))return ApiError.of(code);
            }catch(ApiError ignored){}
        }
        return ApiError.of("FABRIC_UNAVAILABLE");
    }
    public void close(){gateways.values().forEach(Gateway::close);channel.shutdownNow();}
}
