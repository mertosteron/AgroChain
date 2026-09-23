package org.agrochain;

public record Actor(String org, String role) {
    public String msp() { return Character.toUpperCase(org.charAt(0))+org.substring(1)+"MSP"; }
    public String key() { return org+":"+role; }
    public static Actor parse(String value) {
        String[] parts=value.split(":");
        if (parts.length!=2) throw new IllegalArgumentException("Invalid configured actor");
        var a=new Actor(parts[0],parts[1]);
        if (!(a.equals(new Actor("producer","producer")) || a.equals(new Actor("logistics","carrier")) || a.equals(new Actor("retailer","retailer"))
            || (a.org.equals("regulator") && java.util.Set.of("auditor","reviewer","oracle","public-reader").contains(a.role)))) throw new IllegalArgumentException("Invalid configured actor");
        return a;
    }
    public boolean commercial(String kind) {
        return (org.equals("regulator") && !role.equals("public-reader")) || org.equals("retailer")
            || (kind.equals("purchase") && org.equals("producer")) || (kind.equals("freight") && org.equals("logistics"));
    }
    public void authorize(String command) {
        String required=switch(command) {
            case "CreateBatch", "OfferPickup" -> "producer";
            case "AcceptPickup", "RecordFreightCost", "OfferDelivery" -> "logistics";
            case "AcceptDelivery", "ReportRetailPrice" -> "retailer";
            case "EvaluatePrice", "OpenReview", "ResolveReview" -> "regulator";
            default -> throw ApiError.of("UNSUPPORTED_PILOT_OPERATION");
        };
        if (!org.equals(required)) throw ApiError.of("UNAUTHORIZED_ORGANIZATION");
        if(command.equals("EvaluatePrice") && !role.equals("oracle"))throw ApiError.of("UNAUTHORIZED_ROLE");
        if((command.equals("OpenReview") || command.equals("ResolveReview")) && !role.equals("reviewer"))throw ApiError.of("UNAUTHORIZED_ROLE");
    }
}
