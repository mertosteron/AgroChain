package org.agrochain;

import com.fasterxml.jackson.databind.JsonNode;
import java.nio.file.*;
import java.nio.file.attribute.PosixFilePermissions;
import java.sql.*;
import java.util.*;

/** Single-process SQLite journal. Synchronization keeps each update short and atomic. */
public final class Store implements AutoCloseable {
    private final Connection db;
    public final Path documents;
    public Store(Path data) throws Exception {
        Files.createDirectories(data); Files.setPosixFilePermissions(data,PosixFilePermissions.fromString("rwx------"));
        documents=data.resolve("documents");Files.createDirectories(documents);Files.setPosixFilePermissions(documents,PosixFilePermissions.fromString("rwx------"));
        db=DriverManager.getConnection("jdbc:sqlite:"+data.resolve("operations.sqlite"));
        Files.setPosixFilePermissions(data.resolve("operations.sqlite"),PosixFilePermissions.fromString("rw-------"));
        execute("PRAGMA journal_mode=WAL"); execute("PRAGMA synchronous=FULL"); execute("PRAGMA busy_timeout=5000");
        execute("CREATE TABLE IF NOT EXISTS operations (actor TEXT NOT NULL, id TEXT NOT NULL, request TEXT NOT NULL, state TEXT NOT NULL, prepared TEXT, tx BLOB, txid TEXT, receipt TEXT, error TEXT, PRIMARY KEY(actor,id))");
        execute("CREATE TABLE IF NOT EXISTS evidence (id TEXT PRIMARY KEY, operation TEXT NOT NULL, actor TEXT NOT NULL, kind TEXT NOT NULL, bundle TEXT NOT NULL, attachment TEXT, committed INTEGER NOT NULL DEFAULT 0)");
        execute("CREATE TABLE IF NOT EXISTS simulator (key TEXT PRIMARY KEY, request TEXT NOT NULL, bundle TEXT NOT NULL, attachment BLOB)");
        execute("CREATE TABLE IF NOT EXISTS events (txid TEXT PRIMARY KEY, block INTEGER NOT NULL)");
        execute("CREATE TABLE IF NOT EXISTS projections (batch TEXT PRIMARY KEY, lot TEXT UNIQUE, body TEXT NOT NULL)");
    }
    private void execute(String sql) throws SQLException { try(var s=db.createStatement()){s.execute(sql);} }
    private PreparedStatement statement(String sql, Object... args) throws SQLException {
        var s=db.prepareStatement(sql); for(int i=0;i<args.length;i++) s.setObject(i+1,args[i]); return s;
    }
    public synchronized Map<String,Object> row(String sql,Object... args) {
        try(var s=statement(sql,args);var r=s.executeQuery()) {
            if(!r.next())return null;var out=new HashMap<String,Object>();
            for(int i=1;i<=r.getMetaData().getColumnCount();i++)out.put(r.getMetaData().getColumnName(i),r.getObject(i));return out;
        } catch(SQLException e){throw ApiError.of("INTERNAL_ERROR");}
    }
    public synchronized List<Map<String,Object>> rows(String sql,Object... args) {
        try(var s=statement(sql,args);var r=s.executeQuery()) {
            var out=new ArrayList<Map<String,Object>>();
            while(r.next()){var item=new HashMap<String,Object>();for(int i=1;i<=r.getMetaData().getColumnCount();i++)item.put(r.getMetaData().getColumnName(i),r.getObject(i));out.add(item);}return out;
        }catch(SQLException e){throw ApiError.of("INTERNAL_ERROR");}
    }
    public synchronized void update(String sql,Object... args) {
        try(var s=statement(sql,args)){s.executeUpdate();}catch(SQLException e){throw ApiError.of("INTERNAL_ERROR");}
    }
    public synchronized Map<String,Object> stage(Actor actor,String id,JsonNode request) {
        var existing=operation(actor,id);String encoded=Json.canonical(request);
        if(existing!=null){if(!encoded.equals(existing.get("request")))throw ApiError.of("IDEMPOTENCY_CONFLICT");return existing;}
        update("INSERT INTO operations(actor,id,request,state) VALUES(?,?,?,'PENDING')",actor.key(),id,encoded);
        return operation(actor,id);
    }
    public Map<String,Object> operation(Actor actor,String id){return row("SELECT * FROM operations WHERE actor=? AND id=?",actor.key(),id);}
    public synchronized void archive(Actor actor,String op,JsonNode bundle,byte[] attachment) {
        String id=Json.id(Json.text(bundle.path("envelope").path("header"),"documentId"),"DOC");
        String filename=null;
        if(attachment!=null){
            filename=id+".bin";Path dest=documents.resolve(filename);
            try {if(!Files.exists(dest)){Path tmp=Files.createTempFile(documents,"stage-",".tmp");Files.write(tmp,attachment);Files.move(tmp,dest,StandardCopyOption.ATOMIC_MOVE);}}
            catch(Exception e){throw ApiError.of("INTERNAL_ERROR");}
        }
        update("INSERT OR IGNORE INTO evidence(id,operation,actor,kind,bundle,attachment) VALUES(?,?,?,?,?,?)",id,op,actor.key(),Json.text(bundle.path("envelope").path("header"),"documentType"),Json.canonical(bundle),filename);
    }
    public synchronized void committed(Actor actor,String id,JsonNode receipt) {
        try {
            db.setAutoCommit(false);
            update("UPDATE operations SET state='COMMITTED',receipt=?,txid=?,error=NULL WHERE actor=? AND id=?",Json.canonical(receipt),receipt.path("txId").asText(),actor.key(),id);
            update("UPDATE evidence SET committed=1 WHERE actor=? AND operation=?",actor.key(),id);
            db.commit();
        }catch(Exception e){try{db.rollback();}catch(SQLException ignored){}throw ApiError.of("INTERNAL_ERROR");}
        finally{try{db.setAutoCommit(true);}catch(SQLException e){throw ApiError.of("INTERNAL_ERROR");}}
    }
    public synchronized boolean seen(String txid){return row("SELECT txid FROM events WHERE txid=?",txid)!=null;}
    public synchronized long checkpoint(){var r=row("SELECT COALESCE(MAX(block),0) AS b FROM events");return ((Number)r.get("b")).longValue();}
    public synchronized void project(String txid,long block,String batch,String lot,JsonNode body) {
        try {
            db.setAutoCommit(false);
            update("INSERT OR IGNORE INTO events(txid,block) VALUES(?,?)",txid,block);
            update("INSERT INTO projections(batch,lot,body) VALUES(?,?,?) ON CONFLICT(batch) DO UPDATE SET lot=excluded.lot,body=excluded.body",batch,lot,Json.canonical(body));
            db.commit();
        }catch(Exception e){try{db.rollback();}catch(SQLException ignored){}throw ApiError.of("INTERNAL_ERROR");}
        finally{try{db.setAutoCommit(true);}catch(SQLException e){throw ApiError.of("INTERNAL_ERROR");}}
    }
    public void close() throws SQLException {db.close();}
}
