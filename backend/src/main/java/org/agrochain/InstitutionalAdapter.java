package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;

/** Replaceable institutional boundary; the current implementation is SIMULATED only. */
public interface InstitutionalAdapter {
    record Request(String sourceSystem,String documentType,String scenario,JsonNode command,JsonNode batch) {}
    record Document(JsonNode bundle,byte[] original) {}
    Document fetchEvidence(Request request);
    boolean verifyOriginal(String documentId,byte[] original);
}
