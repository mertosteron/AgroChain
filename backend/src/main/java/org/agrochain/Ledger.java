package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import java.util.Map;
import java.util.function.Consumer;

public interface Ledger extends AutoCloseable {
    record Prepared(byte[] bytes,String txId) {}
    record Committed(JsonNode receipt,long block) {}
    record Event(String txId,long block,JsonNode body) {}
    JsonNode query(Actor actor,String function,String... args);
    Prepared endorse(Actor actor,JsonNode command,Map<String,byte[]> transientData);
    Committed submit(Actor actor,Prepared prepared);
    /** Null means status is still unknown, never a confirmed failure. */
    default Committed reconcile(Actor actor,Prepared prepared) { return null; }
    void events(long startBlock,Consumer<Event> consume);
    default void close() {}
    final class UnknownCommit extends RuntimeException {}
}
