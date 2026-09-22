package org.agrochain;

import java.nio.file.Path;

public record Settings(Path root, Path data, Path tokens, Path sourceKeys, String channel, String contract) {
    public static Settings environment() {
        Path root=Path.of(System.getenv().getOrDefault("AGROCHAIN_ROOT", ".")).toAbsolutePath().normalize();
        Path data=Path.of(System.getenv().getOrDefault("AGROCHAIN_BACKEND_DATA",root.resolve("network/runtime/backend").toString()));
        return new Settings(root,data,Path.of(System.getenv().getOrDefault("AGROCHAIN_TOKEN_FILE",data.resolve("tokens.json").toString())),
            Path.of(System.getenv().getOrDefault("AGROCHAIN_SOURCE_KEYS",root.resolve("network/runtime/source-keys").toString())),
            System.getenv().getOrDefault("AGROCHAIN_CHANNEL","agrochannel"), System.getenv().getOrDefault("AGROCHAIN_CHAINCODE","agrochain"));
    }
}
