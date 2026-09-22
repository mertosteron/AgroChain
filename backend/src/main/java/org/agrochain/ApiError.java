package org.agrochain;

import java.util.Map;
import java.util.Set;
import java.util.UUID;

public final class ApiError extends RuntimeException {
    public final String code;
    public final int status;
    public ApiError(String code, int status) { super(code); this.code = code; this.status = status; }
    public static ApiError of(String code) {
        int status = 400;
        if (Set.of("UNAUTHORIZED_ORGANIZATION", "UNAUTHORIZED_ROLE", "PRIVATE_DATA_ACCESS_DENIED", "WRONG_TRANSFER_RECIPIENT").contains(code)) status = 403;
        else if (code.endsWith("_NOT_FOUND")) status = 404;
        else if (code.endsWith("_ALREADY_EXISTS") || code.endsWith("_ALREADY_USED") || Set.of("DUPLICATE_TRANSACTION", "INVALID_STATE_TRANSITION", "VERSION_CONFLICT", "IDEMPOTENCY_CONFLICT").contains(code)) status = 409;
        else if (Set.of("INVALID_SOURCE_SIGNATURE", "UNTRUSTED_SOURCE_KEY", "SOURCE_BINDING_MISMATCH", "SOURCE_CLAIM_REJECTED", "DOCUMENT_HASH_MISMATCH", "MISSING_EVIDENCE", "POLICY_MISMATCH").contains(code)) status = 422;
        else if (code.endsWith("_UNAVAILABLE") || code.equals("CONFIGURATION_REQUIRED")) status = 503;
        else if (code.equals("FABRIC_TRANSACTION_INVALID")) status = 502;
        else if (code.equals("AUTHENTICATION_REQUIRED")) status = 401;
        else if (code.equals("INTERNAL_ERROR")) status = 500;
        return new ApiError(code, status);
    }
    public Map<String,Object> body() {
        return Map.of("schemaVersion", "agrochain.error.v1", "code", code, "message", code.replace('_', ' '),
                "correlationId", "REQ_" + UUID.randomUUID().toString().replace("-", "").toUpperCase(), "retryable", status == 503);
    }
}
