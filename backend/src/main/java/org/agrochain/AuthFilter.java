package org.agrochain;

import jakarta.servlet.*;
import jakarta.servlet.http.*;
import org.springframework.web.filter.OncePerRequestFilter;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.*;

public final class AuthFilter extends OncePerRequestFilter {
    private final Map<Actor,byte[]> tokens=new HashMap<>();
    public AuthFilter(Settings settings) throws IOException {
        var config=Json.parse(Files.readString(settings.tokens()));
        Set<String> unique=new HashSet<>();
        config.fields().forEachRemaining(e->{
            String token=e.getValue().asText();if(!token.matches("[0-9a-f]{64}"))throw new IllegalArgumentException("Invalid token configuration");
            if(!unique.add(token))throw new IllegalArgumentException("Duplicate configured token");
            tokens.put(Actor.parse(e.getKey()),token.getBytes(StandardCharsets.UTF_8));
        });
        for(String actor:List.of("producer:producer","logistics:carrier","retailer:retailer","regulator:auditor"))
            if(!tokens.containsKey(Actor.parse(actor)))throw new IllegalArgumentException("Missing pilot actor tokens");
    }
    protected void doFilterInternal(HttpServletRequest req,HttpServletResponse res,FilterChain chain) throws IOException,ServletException {
        res.setHeader("Cache-Control","no-store");res.setHeader("X-Content-Type-Options","nosniff");
        res.setHeader("Content-Security-Policy","default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'");
        res.setHeader("Referrer-Policy","no-referrer");
        res.setHeader("X-Source-Mode","SIMULATED");res.setHeader("X-Blockchain-Mode","FABRIC");
        String path=req.getRequestURI();
        boolean publicRead=req.getMethod().equals("GET") && (Set.of("/","/index.html","/app.js","/style.css","/consumer.html","/consumer.js").contains(path) || path.equals("/api/v1/health") || path.startsWith("/api/v1/public/lots/"));
        if(!publicRead){
            String authorization=req.getHeader("Authorization");Actor actor=null;
            if(authorization!=null && authorization.startsWith("Bearer ") && authorization.length()==71){
                byte[] supplied=authorization.substring(7).getBytes(StandardCharsets.UTF_8);
                for(var entry:tokens.entrySet())if(MessageDigest.isEqual(supplied,entry.getValue()))actor=entry.getKey();
            }
            if(actor==null){res.setStatus(401);res.setContentType("application/json");res.getWriter().write(Json.encode(ApiError.of("AUTHENTICATION_REQUIRED").body()));return;}
            req.setAttribute("actor",actor);
        }
        chain.doFilter(req,res);
    }
}
