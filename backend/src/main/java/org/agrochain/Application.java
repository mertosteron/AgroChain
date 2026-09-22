package org.agrochain;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.boot.context.event.ApplicationReadyEvent;
import org.springframework.context.event.EventListener;
import jakarta.annotation.PreDestroy;
import java.util.*;
import java.util.concurrent.*;

@SpringBootApplication
public class Application {
    public static void main(String[] args){SpringApplication.run(Application.class,args);}
    @Bean Settings settings(){return Settings.environment();}
    @Bean Store store(Settings s) throws Exception{return new Store(s.data());}
    @Bean Ledger ledger(Settings s) throws Exception{return new FabricLedger(s);}
    @Bean AuthFilter auth(Settings s) throws Exception{return new AuthFilter(s);}
    @Bean InstitutionalAdapter adapter(Settings s,Store store){return new Simulators(s,store,new HashSet<>(Arrays.asList(System.getenv().getOrDefault("AGROCHAIN_DISABLED_SIMULATORS","").split(","))));}
    @Bean EvidenceService evidence(Ledger l,Store s,InstitutionalAdapter a){return new EvidenceService(l,s,a);}
    @Bean Workflow workflow(Store s,Ledger l,EvidenceService e){return new Workflow(s,l,e);}
    @Bean Projection projection(Store s,Ledger l){return new Projection(s,l);}
    @Bean Workers workers(Ledger l,Store s,Workflow w,Projection p){return new Workers(l,s,w,p);}
    static final class Workers {
        private final Ledger ledger;private final Store store;private final Workflow workflow;private final Projection projection;
        private final ScheduledExecutorService recovery=Executors.newSingleThreadScheduledExecutor();
        private volatile boolean stopped;
        Workers(Ledger l,Store s,Workflow w,Projection p){ledger=l;store=s;workflow=w;projection=p;}
        @EventListener(ApplicationReadyEvent.class) public void start(){
            recovery.scheduleWithFixedDelay(()->{try{workflow.recover();}catch(Exception e){org.slf4j.LoggerFactory.getLogger(Workers.class).warn("Operation recovery will retry");}},2,2,TimeUnit.SECONDS);
            Thread.ofVirtual().name("agrochain-public-projection").start(()->{
                while(!stopped){
                    try{ledger.events(store.checkpoint(),projection::accept);}
                    catch(Exception e){if(!stopped)org.slf4j.LoggerFactory.getLogger(Workers.class).warn("Projection disconnected; checkpoint retained");}
                    try{Thread.sleep(2000);}catch(InterruptedException e){Thread.currentThread().interrupt();return;}
                }
            });
        }
        @PreDestroy public void stop(){stopped=true;recovery.shutdownNow();ledger.close();}
    }
}
