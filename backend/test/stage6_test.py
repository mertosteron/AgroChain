#!/usr/bin/env python3
"""Stage 6 real HTTP/Fabric acceptance, using the existing lifecycle fixture."""
import copy
import json
import time
import unittest
import urllib.request
import integration_test as base


class Stage6Integration(base.BackendIntegration):
    def test_11_analysis_permissions_and_boundary_results(self):
        base.ACTORS.update(reviewer="regulator:reviewer", oracle="regulator:oracle")
        for s in self.scenarios:
            batch=s["batch"]
            for _ in range(100):
                status,result=base.call("GET",f"/batches/{batch}/anomaly","regulator")
                if status==200 and result.get("classification"):break
                time.sleep(.3)
            self.assertEqual(status,200,str(result))
            self.assertEqual(result["increaseBps"],4000 if s["price"]==2800 else 8500)
            self.assertEqual(result["classification"],"NO_SIGNAL" if s["price"]==2800 else "REVIEW_REQUIRED")
            self.assertEqual(result["thresholdBps"],5000)
            for actor in ["logistics","producer","public-reader"]:
                for route in ["anomaly","reviews"]:
                    self.assertEqual(base.call("GET",f"/batches/{batch}/{route}",actor)[0],403)
            self.assertEqual(base.call("GET",f"/batches/{batch}/anomaly","retailer")[0],200)
            s["anomaly"]=result

    def review_request(self,s,fn,version):
        r=base.create_request();r["command"].update(command=fn,batchId=s["batch"],expectedVersion=version,payload={"anomalyId":s["anomaly"]["anomalyId"]})
        r["privateInput"]={"reviewerRef":"REV-TEST0001"}
        if fn=="ResolveReview":r["privateInput"].update(outcome="EXPLAINED",explanation="Ürün kalitesi ve nakliye koşulları incelendi; bu bir hukuki ihlal kararı değildir.")
        return r

    def test_12_review_transitions_immutable_result_and_replay(self):
        normal,suspicious=self.scenarios
        for s in self.scenarios:
            s["version"]=base.call("GET",f'/batches/{s["batch"]}',"reviewer")[1]["version"]
        r=self.review_request(normal,"OpenReview",normal["version"])
        self.assertEqual(self.submit("reviewer",f'/anomalies/{normal["anomaly"]["anomalyId"]}/review-actions',r,409)["code"],"INVALID_REVIEW_TRANSITION")
        route=f'/anomalies/{suspicious["anomaly"]["anomalyId"]}/review-actions'
        resolve=self.review_request(suspicious,"ResolveReview",suspicious["version"])
        self.assertEqual(self.submit("reviewer",route,resolve,409)["code"],"INVALID_REVIEW_TRANSITION")
        opening=self.review_request(suspicious,"OpenReview",suspicious["version"])
        for actor in ["regulator","oracle","retailer"]:self.assertEqual(self.submit(actor,route,opening,403)["code"],"UNAUTHORIZED_ORGANIZATION" if actor=="retailer" else "UNAUTHORIZED_ROLE")
        first=self.submit("reviewer",route,opening)
        self.assertEqual(self.submit("reviewer",route,opening,200)["txId"],first["txId"])
        resolve=self.review_request(suspicious,"ResolveReview",suspicious["version"]+1)
        self.submit("reviewer",route,resolve)
        result=base.call("GET",f'/batches/{suspicious["batch"]}/anomaly',"regulator")[1]
        self.assertEqual(result["reviewState"],"RESOLVED")
        for field in ["classification","increaseBps","thresholdBps","operationId","txId"]:self.assertEqual(result[field],suspicious["anomaly"][field])
        history=base.call("GET",f'/batches/{suspicious["batch"]}/reviews',"reviewer")[1]
        self.assertEqual(len(history),2);self.assertEqual({h["action"] for h in history},{"OPEN_REVIEW","RESOLVE_REVIEW"})
        self.assertTrue(any("Ürün" in h.get("explanation","") for h in history))
        repeat=self.review_request(suspicious,"OpenReview",suspicious["version"]+2)
        self.assertEqual(self.submit("reviewer",route,repeat,409)["code"],"INVALID_REVIEW_TRANSITION")

    def test_13_public_ui_qr_and_restart(self):
        for path in ["/","/app.js","/style.css","/consumer.html","/consumer.js"]:
            with urllib.request.urlopen(f"http://127.0.0.1:{base.PORT}"+path) as response:
                self.assertEqual(response.status,200)
                self.assertIn("frame-ancestors 'none'",response.headers["Content-Security-Policy"])
                self.assertNotIn(b"Bearer "+base.TOKENS["regulator:reviewer"].encode(),response.read())
        for s in self.scenarios:
            status,image=base.call("GET",f'/public/lots/{s["lot"]}/qr')
            self.assertEqual(status,200);self.assertTrue(image.startswith(b'\x89PNG'))
            trace=base.call("GET",f'/public/lots/{s["lot"]}')[1]
            self.assertEqual(len(trace["history"]),3)
            for private in ["classification","reviewState","anomalyId","increaseBps","recordSaltHex","explanation","Kurus"]:self.assertNotIn(private,json.dumps(trace))
        self.stop();self.start()
        s=self.scenarios[1];result=base.call("GET",f'/batches/{s["batch"]}/anomaly',"regulator")[1]
        self.assertEqual(result["reviewState"],"RESOLVED")
        self.assertEqual(len(base.call("GET",f'/batches/{s["batch"]}/reviews',"reviewer")[1]),2)
        (base.DATA/"stage6-summary.json").write_text(json.dumps([{k:s[k] for k in ["batch","lot","price","anomaly"]} for s in self.scenarios],indent=2))
    def test_14_pending_survives_restart_and_policy_mismatch(self):
        previous=self.scenarios
        self.stop();self.start(threshold="75")
        self.test_02_normal_and_suspicious_full_workflows()
        waiting=self.scenarios
        for s in waiting:
            for _ in range(60):
                status,_=base.call("GET",f'/public/lots/{s["lot"]}')
                if status==200:break
                time.sleep(.25)
            time.sleep(2)
            status,result=base.call("GET",f'/batches/{s["batch"]}/anomaly',"regulator")
            self.assertEqual((status,result["status"]),(200,"EVALUATION_PENDING"))
        self.stop();self.start()
        for s in waiting:
            for _ in range(100):
                status,result=base.call("GET",f'/batches/{s["batch"]}/anomaly',"regulator")
                if result.get("classification"):break
                time.sleep(.3)
            self.assertEqual(result["increaseBps"],4000 if s["price"]==2800 else 8500)
        self.__class__.scenarios=previous


if __name__=="__main__":unittest.main(verbosity=2)
