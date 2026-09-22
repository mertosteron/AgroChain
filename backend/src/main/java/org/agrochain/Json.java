package org.agrochain;

import com.fasterxml.jackson.core.StreamReadConstraints;
import com.fasterxml.jackson.core.StreamReadFeature;
import com.fasterxml.jackson.databind.*;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.fasterxml.jackson.databind.node.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

/** Exact JSON parsing plus the same ASCII/integer canonical subset as chaincode. */
public final class Json {
    public static final ObjectMapper MAPPER = JsonMapper.builder().enable(StreamReadFeature.STRICT_DUPLICATE_DETECTION).enable(DeserializationFeature.FAIL_ON_TRAILING_TOKENS).build();
    static { MAPPER.getFactory().setStreamReadConstraints(StreamReadConstraints.builder().maxNestingDepth(16).maxStringLength(7_000_000).build()); }
    public static JsonNode parse(String raw) {
        try {
            try(var parser=MAPPER.createParser(raw)) {
                while(parser.nextToken()!=null)if(parser.currentToken().isNumeric() && (parser.currentToken()==com.fasterxml.jackson.core.JsonToken.VALUE_NUMBER_FLOAT || parser.getText().equals("-0")))throw ApiError.of("INVALID_SCHEMA");
            }
            return MAPPER.readTree(raw);
        } catch (Exception e) { throw ApiError.of("INVALID_SCHEMA"); }
    }
    public static JsonNode parse(byte[] raw) { return parse(new String(raw, StandardCharsets.UTF_8)); }
    public static ObjectNode object() { return MAPPER.createObjectNode(); }
    public static ObjectNode obj(Object... pairs) {
        ObjectNode out = object();
        for (int i=0;i<pairs.length;i+=2) out.set((String)pairs[i], MAPPER.valueToTree(pairs[i+1]));
        return out;
    }
    public static String encode(Object v) { try { return MAPPER.writeValueAsString(v); } catch (Exception e) { throw ApiError.of("INTERNAL_ERROR"); } }
    public static byte[] bytes(JsonNode v) { return canonical(v).getBytes(StandardCharsets.UTF_8); }
    public static String canonical(JsonNode v) {
        if (v == null || v.isNull()) throw ApiError.of("INVALID_SCHEMA");
        if (v.isObject()) {
            var sorted = new TreeMap<String,JsonNode>(); v.fields().forEachRemaining(e -> sorted.put(e.getKey(),e.getValue()));
            var fields = new ArrayList<String>(); sorted.forEach((k,x)->fields.add(canonical(MAPPER.valueToTree(k))+":"+canonical(x)));
            return "{"+String.join(",",fields)+"}";
        }
        if (v.isArray()) { var items=new ArrayList<String>(); v.forEach(x->items.add(canonical(x))); return "["+String.join(",",items)+"]"; }
        if (v.isTextual()) {
            if (!v.textValue().chars().allMatch(c->c>=32 && c<=126)) throw ApiError.of("INVALID_SCHEMA");
            return encode(v.textValue());
        }
        if (v.isBoolean() || (v.isIntegralNumber() && v.canConvertToLong())) return v.toString();
        throw ApiError.of("INVALID_SCHEMA");
    }
    public static void fields(JsonNode node, String... fields) {
        if (node==null || !node.isObject()) throw ApiError.of("INVALID_SCHEMA");
        Set<String> actual=new HashSet<>(); node.fieldNames().forEachRemaining(actual::add);
        if (!actual.equals(Set.of(fields))) throw ApiError.of("INVALID_SCHEMA");
    }
    public static String text(JsonNode n, String key) {
        if (n==null || !n.has(key) || !n.get(key).isTextual()) throw ApiError.of("INVALID_SCHEMA"); return n.get(key).textValue();
    }
    public static long number(JsonNode n, String key) {
        if (n==null || !n.has(key) || !n.get(key).isIntegralNumber() || !n.get(key).canConvertToLong()) throw ApiError.of("INVALID_SCHEMA"); return n.get(key).longValue();
    }
    public static String id(String value, String prefix) {
        if (value==null || !value.matches(prefix+"-[A-Z0-9]{8,40}")) throw ApiError.of("INVALID_IDENTIFIER"); return value;
    }
}
